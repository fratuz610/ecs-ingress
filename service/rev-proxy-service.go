package service

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
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
	latestHash   string
}

// NewRevProxyService Creates a new rev proxy service
func NewRevProxyService(cfg *shared.Config, ecsService *EcsService, nginxMonitor *NginxMonitor) *RevProxyService {

	ret := &RevProxyService{
		cfg:          cfg,
		ecsService:   ecsService,
		nginxMonitor: nginxMonitor,
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
		log.Info().Msgf("Reading template %v", r.cfg.Nginx.UpstreamsTemplateFile)
	}

	tmpl, err := template.ParseFiles(r.cfg.Nginx.UpstreamsTemplateFile)

	if err != nil {
		return fmt.Errorf("Template parsing failed: %v", err.Error())
	}

	buffer := new(bytes.Buffer)
	tmpl.Execute(buffer, descrMap)

	currentUpstreamHash := util.HashBuffer(buffer)

	// we download the nginx config file
	nginxConf, err := util.HTTPDownloadFile(r.cfg.Nginx.ConfigFileURL, r.cfg.Nginx.ConfigFileURLHeader)

	if err != nil {
		return fmt.Errorf("Unable to download NGINX config: %v", err.Error())
	}

	if verbose {
		log.Info().Msgf("Nginx config downloaded: %v bytes", len(nginxConf))
	}

	currentNginxHash := util.HashString(nginxConf)
	currentHash := fmt.Sprintf("%v-%v", currentUpstreamHash, currentNginxHash)

	// nothing has changed
	if currentHash == r.latestHash {
		return err
	}

	// we store a reference
	r.latestHash = currentHash

	log.Info().Msgf("Change detected %v. Nginx file size: %v bytes", r.latestHash, len(nginxConf))

	// we update the upstreams file
	err = ioutil.WriteFile(r.cfg.Nginx.UpstreamsConfigFile, buffer.Bytes(), 0644)

	if err != nil {
		return fmt.Errorf("Nginx upstream file update failed: %v", err.Error())
	}

	// we update the main config file
	err = ioutil.WriteFile(r.cfg.Nginx.MainConfigFile, []byte(nginxConf), 0644)

	if err != nil {
		return fmt.Errorf("Nginx main config file update failed: %v", err.Error())
	}

	// we test the new configuration first

	output, err := r.nginxMonitor.TestConfig()

	if err != nil {
		return fmt.Errorf("Error testing NGINX configuration: %v - Unable to proceed", output)
	}

	log.Info().Msg("NGINX configuration test SUCCESS")

	// we send a reload
	r.nginxMonitor.Reload()

	if verbose {
		log.Info().Msg("QueryAndUpdate SUCCESS")
	}

	return nil
}
