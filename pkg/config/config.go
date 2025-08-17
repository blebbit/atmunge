package config

import (
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
	PlcUpstream    string `split_words:"true" default:"https://plc.directory/export"`
	PlcFilter      bool   `split_words:"true" default:"false"`
	PlcMirrorDelay int    `split_words:"true" default:"6"`

	// repo config
	RepoDataDir string `split_words:"true" default:"./data/repos"`

	// server config
	RunPlcMirror  bool   `split_words:"true" default:"true"`
	RunRepoMirror bool   `split_words:"true" default:"false"`
	RunServer     bool   `split_words:"true" default:"true"`
	HTTPPort      string `split_words:"true" default:"1323"`

	// AI config
	OllamaHost string `split_words:"true" default:"http://localhost:11434"`
}

var cfg *Config

func GetConfig() (*Config, error) {
	if cfg == nil {
		err := LoadConfig()
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func LoadConfig() error {
	var c Config
	err := envconfig.Process("atmirror", &c)
	if err != nil {
		return err
	}

	cfg = &c
	return nil
}
