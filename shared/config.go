package shared

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config The shared goroutine-safe config object
type Config struct {
	AWS   configAWS
	Nginx configNginx
}

type configAWS struct {
	ClusterName string
	Region      string
}

type configNginx struct {
	UpstreamsTemplateFile string
	UpstreamsConfigFile   string
	MainConfigFile        string
	ConfigFileURL         string
	ConfigFileURLHeader   string
}

// NewConfig is used to generate a configuration instance which will be passed around the codebase
func NewConfig() (*Config, error) {

	// we create a default
	config := Config{
		AWS: configAWS{
			ClusterName: "default",
			Region:      "ap-southeast-2",
		},
		Nginx: configNginx{
			UpstreamsTemplateFile: "/app/upstreams.conf.tmpl",
			UpstreamsConfigFile:   "/app/upstreams.conf",
			MainConfigFile:        "/app/nginx.conf",
			ConfigFileURL:         "https://fratuz610.s3.amazonaws.com/ecs-ingress/nginx.conf",
			ConfigFileURLHeader:   "",
		},
	}

	viper.BindEnv("AWS.Clustername", "AWS_CLUSTER_NAME")
	viper.BindEnv("AWS.Region", "AWS_REGION")
	viper.BindEnv("Nginx.ConfigFileURL", "NGINX_CONFIG_FILE_URL")
	viper.BindEnv("Nginx.ConfigFileURLHeader", "NGINX_CONFIG_FILE_URL_HEADER")

	if err := viper.Unmarshal(&config); err != nil {
		log.Panic().Msgf("Error unmarshaling config, %s", err)
	}

	log.Info().
		Str("AWS Clustername", config.AWS.ClusterName).
		Str("AWS Region", config.AWS.Region).
		Str("NGINX ConfigFileURL", config.Nginx.ConfigFileURL).
		Str("NGINX ConfigFileURLHeader", config.Nginx.ConfigFileURLHeader).
		Msgf("Config loaded successfully")

	return &config, nil
}
