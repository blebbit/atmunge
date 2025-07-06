package config

import (
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
	PlcUpstream           string `split_words:"true" default:"https://plc.directory"`
	PlcFilter             bool   `split_words:"true" default:"false"`
	PlcFilterKeep         bool   `split_words:"true" default:"false"`
	PlcConflictUpdate     bool   `split_words:"true" default:"false"`
	PlcConflictUpdateKeep bool   `split_words:"true" default:"false"`

	// server config
	RunPlcMirror  bool   `split_words:"true" default:"true"`
	RunRepoMirror bool   `split_words:"true" default:"false"`
	RunServer     bool   `split_words:"true" default:"true"`
	HTTPPort      string `split_words:"true" default:"1323"`
}

var cfg *Config

func GetConfig() *Config {
	if cfg == nil {
		LoadConfig()
	}
	return cfg
}

func LoadConfig() error {
	var c Config

	if err := envconfig.Process("atmirror", &c); err != nil {
		log.Fatalf("envconfig.Process: %s", err)
	}

	cfg = &c

	return nil
}
