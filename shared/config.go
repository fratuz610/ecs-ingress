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
	ConfigFolder          string
	UpstreamsTemplateFile string
	UpstreamsConfigFile   string
	MainConfigFile        string
	ConfigBundleURL       string
	ConfigBundleURLHeader string
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
			ConfigFolder:          "/app/nginx",
			UpstreamsTemplateFile: "upstreams.conf.tmpl",
			UpstreamsConfigFile:   "upstreams.conf",
			MainConfigFile:        "nginx.conf",
			ConfigBundleURL:       "https://github.com/fratuz610/ecs-ingress/blob/master/bundle/nginx.zip?raw=true",
			ConfigBundleURLHeader: "",
		},
	}

	viper.BindEnv("AWS.Clustername", "AWS_CLUSTER_NAME")
	viper.BindEnv("AWS.Region", "AWS_REGION")
	viper.BindEnv("Nginx.MainConfigFile", "NGINX_CONFIG_FILE_NAME")
	viper.BindEnv("Nginx.ConfigBundleURL", "NGINX_CONFIG_BUNDLE_URL")
	viper.BindEnv("Nginx.ConfigBundleURLHeader", "NGINX_CONFIG_BUNDLE_URL_HEADER")

	if err := viper.Unmarshal(&config); err != nil {
		log.Panic().Msgf("Error unmarshaling config, %s", err)
	}

	log.Info().
		Str("AWS Clustername", config.AWS.ClusterName).
		Str("AWS Region", config.AWS.Region).
		Str("NGINX ConfigFolder", config.Nginx.ConfigFolder).
		Str("NGINX MainConfigFile", config.Nginx.MainConfigFile).
		Str("NGINX ConfigBundleURL", config.Nginx.ConfigBundleURL).
		Str("NGINX ConfigBundleURLHeader", config.Nginx.ConfigBundleURLHeader).
		Msgf("Config loaded successfully")

	return &config, nil
}
