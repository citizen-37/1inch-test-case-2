package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	yamlDoc "gopkg.in/yaml.v3"
	"os"
)

type (
	Config struct {
		Envs
		Yaml
	}

	Envs struct {
		InfuraKey string `envconfig:"infura_key"`
	}

	Yaml struct {
		PoolV2Addresses []string `yaml:"pool_v2_addresses"`
		PoolV3Addresses []string `yaml:"pool_v3_addresses"`
	}
)

func Load() (Config, error) {
	envCfg := Envs{}

	err := envconfig.Process("", &envCfg)
	if err != nil {
		return Config{}, fmt.Errorf("env config failed: %w", err)
	}

	yamlCfg := Yaml{}
	file, err := os.Open("config.yaml")
	if err != nil {
		return Config{}, fmt.Errorf("config.yaml open failed: %w", err)
	}

	err = yamlDoc.NewDecoder(file).Decode(&yamlCfg)
	if err != nil {
		return Config{}, fmt.Errorf("yaml config decode failed: %w", err)
	}

	return Config{
		Envs: envCfg,
		Yaml: yamlCfg,
	}, nil
}
