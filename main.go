package main

import (
	"os"
	"sync"

	"bitbucket.org/nnnco/rev-proxy/service"
	"bitbucket.org/nnnco/rev-proxy/shared"
	"bitbucket.org/nnnco/rev-proxy/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ServiceTasks an ecs task cut down view
type ServiceTasks struct {
	ServiceName string
	TaskArnList []string
}

func main() {

	// we define the global logger
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	zerolog.TimestampFieldName = "t"
	zerolog.LevelFieldName = "l"
	zerolog.MessageFieldName = "m"
	zerolog.TimeFieldFormat = ""

	config, _ := shared.NewConfig()
	ecsService := service.NewEcsService(config)

	nginxMonitor := service.NewNginxMonitor(config)
	s3Client := util.NewS3Client(config)

	revProxyService := service.NewRevProxyService(config, ecsService, nginxMonitor, s3Client)

	var wg sync.WaitGroup
	wg.Add(1)

	// we generate the first configuration
	err := revProxyService.QueryAndUpdate(true)

	if err != nil {
		panic(err.Error())
	}

	log.Info().Msg("First ECS configuration loaded successfully")

	// we start NGINX
	go nginxMonitor.Start(&wg)

	// we start our updating service
	go revProxyService.Start()

	// we wait until the nginx monitor dies
	wg.Wait()
}
