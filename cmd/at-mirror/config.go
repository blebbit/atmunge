package main

import (
	"fmt"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// ops config
	LogFormat   string `split_words:"true" default:"text"`
	LogLevel    int64  `split_words:"true" default:"1"`
	MetricsPort string `split_words:"true"`
	DBUrl       string `envconfig:"POSTGRES_URL"`

	// plc config
	PlcUpstream  string `split_words:"true" default:"https://plc.directory"`
	RunPlcMirror bool   `split_words:"true" default:"true"`

	// repo config
	RunRepoMirror bool `split_words:"true" default:"false"`

	// server config
	RunServer bool   `split_words:"true" default:"true"`
	HTTPPort  string `split_words:"true" default:"1323"`
}

var config Config

func loadConfig() error {
	if err := envconfig.Process("plc", &config); err != nil {
		log.Fatalf("envconfig.Process: %s", err)
	}

	fmt.Println(config)

	return nil
}
