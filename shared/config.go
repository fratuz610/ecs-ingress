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
	ConfigBundleS3Bucket  string
	ConfigBundleS3Key     string
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
			ConfigBundleS3Bucket:  "",
			ConfigBundleS3Key:     "",
		},
	}

	viper.BindEnv("AWS.Clustername", "AWS_CLUSTER_NAME")
	viper.BindEnv("AWS.Region", "AWS_REGION")
	viper.BindEnv("Nginx.MainConfigFile", "NGINX_CONFIG_FILE_NAME")
	viper.BindEnv("Nginx.ConfigBundleS3Bucket", "NGINX_CONFIG_BUNDLE_S3_BUCKET")
	viper.BindEnv("Nginx.ConfigBundleS3Key", "NGINX_CONFIG_BUNDLE_S3_KEY")

	if err := viper.Unmarshal(&config); err != nil {
		log.Panic().Msgf("Error unmarshaling config, %s", err)
	}

	log.Info().
		Str("AWS Clustername", config.AWS.ClusterName).
		Str("AWS Region", config.AWS.Region).
		Str("NGINX ConfigFolder", config.Nginx.ConfigFolder).
		Str("NGINX MainConfigFile", config.Nginx.MainConfigFile).
		Str("NGINX ConfigBundleS3Bucket", config.Nginx.ConfigBundleS3Bucket).
		Str("NGINX ConfigBundleS3Key", config.Nginx.ConfigBundleS3Key).
		Msgf("Config loaded successfully")

	return &config, nil
}
