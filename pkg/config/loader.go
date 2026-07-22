package config

import (
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var envPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

func Load(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	content := envPattern.ReplaceAllStringFunc(string(data), func(s string) string {
		match := envPattern.FindStringSubmatch(s)
		if len(match) != 2 {
			return s
		}

		value := os.Getenv(match[1])
		if value == "" {
			return s
		}

		return value
	})

	var cfg Config

	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func IsEmpty(value any) bool {
	if value == nil {
		return true
	}

	if s, ok := value.(string); ok {
		return strings.TrimSpace(s) == ""
	}

	return false
}
