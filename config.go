package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Domain string
		Port   string
	}

	Database struct {
		Name     string
		Endpoint string
		Port     string
		SSLMode  string
		Username string
		Password string
	}

	AWS struct {
		Region         string
		AccessKeyID    string
		SecretAcessKey string
	}
}

func NewConfig() *Config {
	cfg := &Config{}

	if Profile == "local" {
		err := godotenv.Load()
		if err != nil {
			panic(err)
		}

		cfg.App.Domain = os.Getenv("APP_DOMAIN")
		cfg.App.Port = os.Getenv("APP_PORT")
		cfg.Database.Name = os.Getenv("DB_NAME")
		cfg.Database.Endpoint = os.Getenv("DB_ENDPOINT")
		cfg.Database.SSLMode = os.Getenv("DB_SSL_MODE")
		cfg.Database.Username = os.Getenv("DB_USERNAME")
		cfg.Database.Password = os.Getenv("DB_PASSWORD")
		cfg.AWS.Region = os.Getenv("AWS_REGION")
		cfg.AWS.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		cfg.AWS.SecretAcessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	} else {
		cfg.App.Domain = getSecretFromDockerSwarm("go_template_domain")
		cfg.App.Port = getSecretFromDockerSwarm("go_template_port")
		cfg.Database.Name = getSecretFromDockerSwarm("go_template_db_name")
		cfg.Database.Endpoint = getSecretFromDockerSwarm("go_template_db_endpoint")
		cfg.Database.SSLMode = getSecretFromDockerSwarm("go_template_db_ssl_mode")
		cfg.Database.Username = getSecretFromDockerSwarm("go_template_db_username")
		cfg.Database.Password = getSecretFromDockerSwarm("go_template_db_password")
		cfg.AWS.Region = getSecretFromDockerSwarm("go_template_aws_region")
		cfg.AWS.AccessKeyID = getSecretFromDockerSwarm("go_template_aws_access_key_id")
		cfg.AWS.SecretAcessKey = getSecretFromDockerSwarm("go_template_aws_secret_access_key")
	}

	return cfg
}

func getSecretFromDockerSwarm(secretName string) string {
	secretFile, err := os.Open("/run/secrets/" + secretName)
	if err != nil {
		panic(fmt.Errorf("can't open secret \"%s\" in docker swarm", secretName))
	}
	defer secretFile.Close()

	secretContent, err := os.ReadFile("/run/secrets/" + secretName)
	if err != nil {
		panic(fmt.Errorf("can't find secret \"%s\" in docker swarm", secretName))
	}

	return string(secretContent)
}
