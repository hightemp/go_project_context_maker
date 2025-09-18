package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	cfg "go_project_context_maker/internal/config"
	"go_project_context_maker/internal/generator"
)

const defaultConfigPath = "config.yaml"

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", defaultConfigPath, "path to config.yaml (used for both init and generate)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s [flags] <command>\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  init       Create a default config.yaml (use -config to choose path)\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  generate   Run generation according to config.yaml\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	cmd := args[0]
	switch cmd {
	case "init":
		if err := runInit(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "init error: %v\n", err)
			os.Exit(1)
		}
	case "generate":
		if err := runGenerate(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "generate error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", cmd)
		flag.Usage()
		os.Exit(2)
	}
}

func runInit(path string) error {
	if path == "" {
		path = defaultConfigPath
	}
	// Avoid overwriting existing config to be safe by default
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot stat %s: %w", path, err)
	}

	def := cfg.Default()
	if err := cfg.Save(path, def); err != nil {
		return err
	}

	fmt.Printf("Default config created at %s\n", path)
	return nil
}

func runGenerate(path string) error {
	if path == "" {
		path = defaultConfigPath
	}
	conf, err := cfg.Load(path)
	if err != nil {
		return err
	}

	if err := generator.Generate(conf, "."); err != nil {
		return err
	}

	fmt.Println("Generation completed")
	return nil
}
