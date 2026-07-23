package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

func Load(
	path string,
) (*Config, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := yaml.Unmarshal(
		data,
		cfg,
	); err != nil {
		return nil, err
	}

	return cfg, nil
}
