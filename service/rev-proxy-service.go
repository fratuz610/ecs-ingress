package service

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"time"

	"bitbucket.org/nnnco/rev-proxy/shared"
	"bitbucket.org/nnnco/rev-proxy/util"
	"github.com/rs/zerolog/log"
)

// RevProxyService simplified client to access ECS resources on AWS
type RevProxyService struct {
	cfg          *shared.Config
	ecsService   *EcsService
	nginxMonitor *NginxMonitor
	s3Client     *util.S3Client
	latestHash   string
}

// NewRevProxyService Creates a new rev proxy service
func NewRevProxyService(cfg *shared.Config, ecsService *EcsService, nginxMonitor *NginxMonitor, s3Client *util.S3Client) *RevProxyService {

	ret := &RevProxyService{
		cfg:          cfg,
		ecsService:   ecsService,
		nginxMonitor: nginxMonitor,
		s3Client:     s3Client,
	}

	return ret
}

// Start starts this
func (r *RevProxyService) Start() {

	for range time.Tick(10 * time.Second) {

		err := r.QueryAndUpdate(false)

		if err != nil {
			log.Error().Err(err).Send()
		}
	}
}

// QueryAndUpdate runs the main routine
func (r *RevProxyService) QueryAndUpdate(verbose bool) error {

	if verbose {
		log.Info().Msg("QueryAndUpdate START")
	}

	// we then describe all tasks
	descrMap, err := r.ecsService.GetServicesAndPorts(r.cfg.AWS.ClusterName)

	if err != nil {
		return fmt.Errorf("GetServicesAndPorts failed: %v", err.Error())
	}

	if verbose {
		log.Info().Msgf("GetServicesAndPorts SUCCEEDED: %v services found", len(descrMap))
	}

	upstreamsTemplatePath := filepath.Join(r.cfg.Nginx.ConfigFolder, r.cfg.Nginx.UpstreamsTemplateFile)
	upstreamsConfigPath := filepath.Join(r.cfg.Nginx.ConfigFolder, r.cfg.Nginx.UpstreamsConfigFile)

	if verbose {
		log.Info().Msgf("Reading template %v", upstreamsTemplatePath)
	}

	tmpl, err := template.ParseFiles(upstreamsTemplatePath)

	if err != nil {
		return fmt.Errorf("Template parsing failed: %v", err.Error())
	}

	templateBuffer := new(bytes.Buffer)
	tmpl.Execute(templateBuffer, descrMap)

	currentUpstreamHash := util.HashBuffer(templateBuffer)

	// we download the nginx config file
	nginxConfBundleBytes, err := r.s3Client.DownloadFileInMemory(r.cfg.Nginx.ConfigBundleS3Bucket, r.cfg.Nginx.ConfigBundleS3Key)

	if err != nil {
		return fmt.Errorf("Unable to download NGINX config bundle: %v", err.Error())
	}

	if verbose {
		log.Info().Msgf("Nginx config bundle downloaded: %v bytes", len(nginxConfBundleBytes))
	}

	currentNginxHash := util.HashBytes(nginxConfBundleBytes)
	currentHash := fmt.Sprintf("%v-%v", currentUpstreamHash, currentNginxHash)

	// nothing has changed
	if currentHash == r.latestHash {
		return err
	}

	// we store a reference
	r.latestHash = currentHash

	log.Info().Msgf("Change detected %v. Nginx file size: %v bytes", r.latestHash, len(nginxConfBundleBytes))

	// we update the upstreams file
	err = ioutil.WriteFile(upstreamsConfigPath, templateBuffer.Bytes(), 0644)

	if verbose {
		log.Info().Msgf("Upstreams config: %v", string(templateBuffer.Bytes()))
	}

	if err != nil {
		return fmt.Errorf("Nginx upstream file update failed: %v", err)
	}

	// we unzip the main config bundle into the config folder
	fileList, err := util.UnzipFileFromMemory(nginxConfBundleBytes, r.cfg.Nginx.ConfigFolder)

	if err != nil {
		return fmt.Errorf("Unable to extract bundle: %v", err)
	}

	log.Info().Msgf("%v files extracted.", len(fileList))

	if len(fileList) == 0 {
		return fmt.Errorf("Bundle contained NO files")
	}

	for _, filePath := range fileList {
		log.Info().Msgf("Extracted file: '%v'", filePath)
	}

	// we look for a main config file there
	mainConfigFile := filepath.Join(r.cfg.Nginx.ConfigFolder, r.cfg.Nginx.MainConfigFile)

	if !util.FileExists(mainConfigFile) {
		return fmt.Errorf("Nginx config file NOT found under '%v'", mainConfigFile)
	}

	// we test the new configuration first
	output, err := r.nginxMonitor.TestConfig()

	if err != nil {
		return fmt.Errorf("Error testing NGINX configuration: %v - Unable to proceed", output)
	}

	log.Info().Msg("NGINX configuration test SUCCESS. Sending reload message")

	// we send a reload
	r.nginxMonitor.Reload()

	if verbose {
		log.Info().Msg("QueryAndUpdate SUCCESS")
	}

	return nil
}
