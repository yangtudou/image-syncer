package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

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

type AuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type ImageRule struct {
	Registry   string
	Repository string
	Tags       []string
}

func Load(
	path string,
) (*Config, error) {

	data, err := os.ReadFile(
		path,
	)

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

func (c *Config) NormalizeRules() []ImageRule {

	var rules []ImageRule

	for registry, source := range c.Sources {

		for repository, value := range source.Images {

			rule := ImageRule{
				Registry:   registry,
				Repository: repository,
				Tags:       []string{},
			}

			switch v := value.(type) {

			case nil:

				rule.Tags = append(
					rule.Tags,
					"latest",
				)

			case string:

				tag := os.ExpandEnv(
					v,
				)

				if tag == "" {
					tag = "latest"
				}

				rule.Tags = append(
					rule.Tags,
					tag,
				)

			case []interface{}:

				for _, item := range v {

					tag := os.ExpandEnv(
						fmt.Sprintf(
							"%v",
							item,
						),
					)

					if tag == "" {
						tag = "latest"
					}

					rule.Tags = append(
						rule.Tags,
						tag,
					)
				}

			case []string:

				for _, item := range v {

					tag := os.ExpandEnv(
						item,
					)

					if tag == "" {
						tag = "latest"
					}

					rule.Tags = append(
						rule.Tags,
						tag,
					)
				}
			}

			if len(rule.Tags) == 0 {

				rule.Tags = append(
					rule.Tags,
					"latest",
				)
			}

			rules = append(
				rules,
				rule,
			)
		}
	}

	return rules
}
