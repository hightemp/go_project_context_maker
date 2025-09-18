# Go Project Context Maker

Small Go CLI to generate llm context markdown file with files contents.

## Quick usage

- Create default config:
```bash
./gpcm -config config.yaml init
```

- Generate output:
```bash
./gpcm -config config.yaml generate
```

## Config (YAML)

Minimal example matching the requested behavior:
```yaml
projectPath: "."
documents:
  - description: "Project structure overview"
    outputPath: project-structure.md
    sources:
      - type: tree
        sourcePaths:
          - src
          - migrations
          - templates
        filePattern: "*.php,*.twig"
        excludePaths:
          - vendor
          - node_modules
          - .git        
      - type: file
        sourcePaths:
          - src
          - migrations
          - templates
        filePattern: "*.php,*.twig"
        excludePaths:
          - vendor
          - node_modules
          - .git        
```