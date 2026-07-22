package mapper

import (
	"testing"

	"github.com/AliyunContainerService/image-syncer/pkg/config"
)

func TestGenerate(t *testing.T) {
	cfg := &config.Config{
		Target: config.TargetConfig{
			Registry:  "registry.example.com",
			Namespace: "mirror",
			Flatten:   true,
		},
		Images: map[string]map[string]any{
			"docker.io": {
				"library/registry": "3",
				"library/nginx":    nil,
			},
			"ghcr.io": {
				"sagernet/sing-box": []any{
					"+latest-testing",
				},
			},
		},
	}

	result := Generate(cfg)

	expected := map[string]string{
		"docker.io/library/registry:3": "registry.example.com/mirror/registry:3",

		"docker.io/library/nginx:latest": "registry.example.com/mirror/nginx:latest",

		"ghcr.io/sagernet/sing-box:latest": "registry.example.com/mirror/sing-box:latest",

		"ghcr.io/sagernet/sing-box:latest-testing": "registry.example.com/mirror/sing-box:latest-testing",
	}

	if len(result) != len(expected) {
		t.Fatalf(
			"expected %d mappings, got %d",
			len(expected),
			len(result),
		)
	}

	for _, item := range result {
		if target, ok := expected[item.Source]; ok {
			if target != item.Target {
				t.Fatalf(
					"%s target mismatch: expected %s got %s",
					item.Source,
					target,
					item.Target,
				)
			}
			delete(expected, item.Source)
		}
	}

	if len(expected) != 0 {
		t.Fatalf(
			"missing mappings: %v",
			expected,
		)
	}
}
