package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Helper: 获取示例配置文件相对路径
func getExampleConfigPath() string {
	return filepath.Join("..", "..", "examples", "sync.yaml")
}

func TestLoad(t *testing.T) {
	exampleFile := getExampleConfigPath()

	// 检查文件是否存在，防止路径变更导致报错
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Skipf("Skip: 未找到配置文件 %s", exampleFile)
	}

	cfg, err := Load(exampleFile)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	fmt.Println()
	fmt.Println("========== Config (来自 examples/sync.yaml) ==========")

	fmt.Printf(
		"Destination: %s/%s (Flatten: %v)\n",
		cfg.Dest.Registry,
		cfg.Dest.Namespace,
		cfg.Dest.Flatten,
	)

	fmt.Printf(
		"Sources Count: %d\n",
		len(cfg.Sources),
	)

	for registry, source := range cfg.Sources {
		for image, tags := range source.Images {
			fmt.Printf(
				"  %s/%s => %v\n",
				registry,
				image,
				tags,
			)
		}
	}

	fmt.Println("====================================================")

	// 校验配置加载结果
	if cfg.Dest.Registry == "" {
		t.Fatal("unexpected empty dest registry")
	}

	if len(cfg.Sources) == 0 {
		t.Fatal("expected at least one source in config")
	}
}

func TestNormalizeRulesEnv(t *testing.T) {
	os.Setenv("TEST_TAG", "v1.2.3")
	defer os.Unsetenv("TEST_TAG")

	cfg := &Config{
		Sources: map[string]SourceConfig{
			"docker.io": {
				Images: map[string]any{
					"test/image": "$TEST_TAG",
				},
			},
		},
	}

	rules := cfg.NormalizeRules()

	if len(rules) == 0 || len(rules[0].Tags) == 0 {
		t.Fatal("normalize rules failed or tags empty")
	}

	if rules[0].Tags[0] != "v1.2.3" {
		t.Fatalf("env expand failed got %s", rules[0].Tags[0])
	}
}

func TestNormalizeRulesDefaultLatest(t *testing.T) {
	cfg := &Config{
		Sources: map[string]SourceConfig{
			"docker.io": {
				Images: map[string]any{
					"test/image": "",
				},
			},
		},
	}

	rules := cfg.NormalizeRules()

	if len(rules) == 0 || len(rules[0].Tags) == 0 {
		t.Fatal("normalize rules failed or tags empty")
	}

	if rules[0].Tags[0] != "latest" {
		t.Fatalf("default latest failed got %s", rules[0].Tags[0])
	}
}

func TestNormalizeRulesWithDestination(t *testing.T) {
	exampleFile := getExampleConfigPath()

	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Skipf("Skip: 未找到配置文件 %s", exampleFile)
	}

	cfg, err := Load(exampleFile)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	rules := cfg.NormalizeRules()

	fmt.Println()
	fmt.Println("========== 镜像映射 (来自 examples/sync.yaml) ==========")

	for _, rule := range rules {
		for _, tag := range rule.Tags {
			source := fmt.Sprintf(
				"%s/%s:%s",
				rule.Registry,
				rule.Repository,
				tag,
			)

			// 适配 cfg.Dest.Flatten 控制位
			repoName := rule.Repository
			if cfg.Dest.Flatten {
				repoName = filepath.Base(rule.Repository)
			}

			var destination string
			if cfg.Dest.Namespace != "" {
				destination = fmt.Sprintf(
					"%s/%s/%s:%s",
					cfg.Dest.Registry,
					cfg.Dest.Namespace,
					repoName,
					tag,
				)
			} else {
				destination = fmt.Sprintf(
					"%s/%s:%s",
					cfg.Dest.Registry,
					repoName,
					tag,
				)
			}

			fmt.Printf(
				"%s => %s\n",
				source,
				destination,
			)
		}
	}

	fmt.Println("=======================================================")
}
