package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:"localhost:8080"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel   string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPConfig HTTPConfig `yaml:"api_server"`
	DBUrl      string     `yaml:"db_url" env:"DB_URL"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
