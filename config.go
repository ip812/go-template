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
		Host     string
		Port     string
		SSL      string
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
		cfg.Database.Host = os.Getenv("DB_HOST")
		cfg.Database.Port = os.Getenv("DB_PORT")
		cfg.Database.SSL = os.Getenv("DB_SSL")
		cfg.Database.Username = os.Getenv("DB_USERNAME")
		cfg.Database.Password = os.Getenv("DB_PASSWORD")
		cfg.AWS.Region = os.Getenv("AWS_REGION")
		cfg.AWS.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
		cfg.AWS.SecretAcessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	} else {
		// TODO: Rename them later to go_indie_hacking_starter_*
		cfg.App.Domain = getSecretFromDockerSwarm("blog_domain")
		cfg.App.Port = getSecretFromDockerSwarm("blog_port")
		cfg.Database.Name = getSecretFromDockerSwarm("blog_db_name")
		cfg.Database.Host = getSecretFromDockerSwarm("blog_db_host")
		cfg.Database.Port = getSecretFromDockerSwarm("blog_db_port")
		cfg.Database.SSL = getSecretFromDockerSwarm("blog_db_ssl")
		cfg.Database.Username = getSecretFromDockerSwarm("blog_db_user")
		cfg.Database.Password = getSecretFromDockerSwarm("blog_db_pass")
		cfg.AWS.Region = getSecretFromDockerSwarm("blog_aws_region")
		cfg.AWS.AccessKeyID = getSecretFromDockerSwarm("blog_aws_access_key_id")
		cfg.AWS.SecretAcessKey = getSecretFromDockerSwarm("blog_aws_secret_access_key")
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
