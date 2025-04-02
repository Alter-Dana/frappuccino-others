package main

import (
	"os"
)

type Config struct {
	Postgres struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
	}
}

func getConfig() (*Config, error) {
	var config Config
	config.Postgres = struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
	}{}

	config.Postgres.Host = os.Getenv("DB_HOST")
	config.Postgres.Port = os.Getenv("DB_PORT")
	config.Postgres.User = os.Getenv("DB_USER")
	config.Postgres.Password = os.Getenv("DB_PASSWORD")
	config.Postgres.Database = os.Getenv("DB_NAME")

	return &config, nil
}
