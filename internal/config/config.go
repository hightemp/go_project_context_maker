package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// ProjectPath — корневая директория проекта, относительно которой выполняется поиск.
	// Если пусто, используется ".".
	ProjectPath string `yaml:"projectPath"`

	Documents []Document `yaml:"documents"`
}

type Document struct {
	Description string   `yaml:"description"`
	OutputPath  string   `yaml:"outputPath"`
	Sources     []Source `yaml:"sources"`
}

type Source struct {
	Type         string   `yaml:"type"`         // "tree" or "file"
	SourcePaths  []string `yaml:"sourcePaths"`  // directories to scan
	ExcludePaths []string `yaml:"excludePaths"` // path globs (relative to project root) to exclude; supports simple * and ? globs
	FilePattern  string   `yaml:"filePattern"`  // comma-separated globs for file names, e.g. "*.php,*.twig"
}

// Default returns the default configuration matching the task description.
func Default() Config {
	return Config{
		ProjectPath: ".",
		Documents: []Document{
			{
				Description: "Project structure overview",
				OutputPath:  "project-structure.md",
				Sources: []Source{
					{
						Type:         "tree",
						SourcePaths:  []string{"src", "migrations", "templates"},
						FilePattern:  "*.php,*.twig",
						ExcludePaths: []string{"vendor", "node_modules", ".git"},
					},
					{
						Type:         "file",
						SourcePaths:  []string{"src", "migrations", "templates"},
						FilePattern:  "*.php,*.twig",
						ExcludePaths: []string{"vendor", "node_modules", ".git"},
					},
				},
			},
		},
	}
}

// Load reads configuration from a YAML file.
func Load(path string) (Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return c, err
	}
	return c, nil
}

// Save writes configuration to a YAML file, creating parent directories if needed.
func Save(path string, c Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func ensureDir(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		info, statErr := os.Stat(dir)
		if statErr == nil && !info.IsDir() {
			return errors.New("path exists and is not a directory: " + dir)
		}
		return err
	}
	return nil
}
