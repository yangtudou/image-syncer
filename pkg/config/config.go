package config

type Config struct {
	Dest    DestConfig              `yaml:"dest"`
	Sources map[string]SourceConfig `yaml:"sources"`
}

type DestConfig struct {
	Registry  string      `yaml:"registry"`
	Namespace string      `yaml:"namespace"`
	Flatten   bool        `yaml:"flatten"`
	Auth      *AuthConfig `yaml:"auth,omitempty"`
}

type SourceConfig struct {
	Auth   *AuthConfig    `yaml:"auth,omitempty"`
	Images map[string]any `yaml:"images"`
}
