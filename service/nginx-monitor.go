package service

import (
	"os"
	"os/exec"
	"sync"

	"bitbucket.org/nnnco/rev-proxy/shared"
	"github.com/rs/zerolog/log"
)

// NginxMonitor simplified client to access ECS resources on AWS
type NginxMonitor struct {
	cfg *shared.Config
}

// NewNginxMonitor Creates a new rev proxy service
func NewNginxMonitor(cfg *shared.Config) *NginxMonitor {

	ret := &NginxMonitor{
		cfg: cfg,
	}

	return ret
}

// Start starts and monitor the NGINX process
func (n *NginxMonitor) Start(wg *sync.WaitGroup) {

	log.Info().Msg("Nginx Monitor START")

	// log.Info().Str("errorMsg", response.ErrorMsg).Int("numReadings", len(response.ReadingList)).Float32("durationMs", response.DurationMs).Msg("")

	// we start nginx
	log.Info().Msg("Starting nginx executable...")

	mainCmd := exec.Command("nginx", "-c", "/app/nginx.conf", "-g", "daemon off;")

	// we redirect stdout and err to ourself
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr

	// we run the executable (in the main thread)
	exitErr := mainCmd.Run()

	if exitErr != nil {
		log.Error().Msgf("Nginx exited WITH ERROR: %v", exitErr)
	} else {
		log.Warn().Msg("Nginx exited without error")
	}

	wg.Done()
}

// Reload reloads Nginx config
func (n *NginxMonitor) Reload() {
	log.Info().Msg("Sending reload message to NGINX")
	mainCmd := exec.Command("nginx", "-s", "reload")
	mainCmd.Run()
	log.Info().Msg("Reload sent")
}

// TestConfig Tests the Nginx configuration
func (n *NginxMonitor) TestConfig() (string, error) {
	mainCmd := exec.Command("nginx", "-t")
	byteOut, err := mainCmd.Output()
	return string(byteOut), err
}
