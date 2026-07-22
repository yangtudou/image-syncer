package config

type Config struct {
	Target TargetConfig              `yaml:"target"`
	Images map[string]map[string]any `yaml:"images"`
}

type TargetConfig struct {
	Registry  string   `yaml:"registry"`
	Namespace string   `yaml:"namespace"`
	Flatten   bool     `yaml:"flatten"`
	Platforms []string `yaml:"platforms"`
}
