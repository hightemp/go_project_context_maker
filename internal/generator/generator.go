package generator

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cfg "go_project_context_maker/internal/config"
)

func Generate(c cfg.Config, projectRoot string) error {
	for _, doc := range c.Documents {
		var b strings.Builder

		if doc.Description != "" {
			fmt.Fprintf(&b, "# %s\n\n", doc.Description)
		}

		for _, src := range doc.Sources {
			files, err := collectFiles(projectRoot, src.SourcePaths, src.FilePattern)
			if err != nil {
				return fmt.Errorf("collect files for %q: %w", src.Type, err)
			}

			switch strings.ToLower(src.Type) {
			case "tree":
				if len(files) == 0 {
					fmt.Fprintf(&b, "```\n(no matches for %q in %v)\n```\n\n", src.FilePattern, src.SourcePaths)
					continue
				}
				tree := renderTree(files)
				// Put tree into code block for readability
				fmt.Fprintf(&b, "```\n%s\n```\n\n", tree)

			case "file":
				if len(files) == 0 {
					fmt.Fprintf(&b, "_No files matched %q under %v_\n\n", src.FilePattern, src.SourcePaths)
					continue
				}
				for _, rel := range files {
					abs := filepath.Join(projectRoot, rel)
					data, err := os.ReadFile(abs)
					if err != nil {
						return fmt.Errorf("read %s: %w", rel, err)
					}
					// Show path and content as markdown code block
					// Heading with the path for clarity
					fmt.Fprintf(&b, "### %s\n\n", rel)
					lang := detectLang(rel)
					if lang != "" {
						fmt.Fprintf(&b, "```%s\n", lang)
					} else {
						fmt.Fprintf(&b, "```\n")
					}
					b.Write(data)
					if len(data) > 0 && data[len(data)-1] != '\n' {
						b.WriteByte('\n')
					}
					fmt.Fprintf(&b, "```\n\n")
				}

			default:
				return fmt.Errorf("unknown source type: %q", src.Type)
			}
		}

		if err := ensureDir(filepath.Dir(doc.OutputPath)); err != nil {
			return err
		}
		if err := os.WriteFile(doc.OutputPath, []byte(b.String()), 0o644); err != nil {
			return fmt.Errorf("write output %s: %w", doc.OutputPath, err)
		}
	}

	return nil
}

func collectFiles(root string, dirs []string, patternCSV string) ([]string, error) {
	patterns := splitPatterns(patternCSV)
	seen := make(map[string]struct{})

	for _, d := range dirs {
		start := filepath.Join(root, d)
		info, err := os.Stat(start)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// silently skip non-existent source path
				continue
			}
			return nil, fmt.Errorf("stat %s: %w", start, err)
		}
		if !info.IsDir() {
			// if it's a file, optional: include if matches
			if matchAny(patterns, filepath.Base(d)) || len(patterns) == 0 {
				rel, err := filepath.Rel(root, start)
				if err != nil {
					return nil, err
				}
				seen[rel] = struct{}{}
			}
			continue
		}

		err = filepath.WalkDir(start, func(path string, de fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if de.IsDir() {
				return nil
			}
			name := de.Name()
			if len(patterns) == 0 || matchAny(patterns, name) {
				rel, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				// Normalize to OS-specific separators already returned by Rel
				seen[rel] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", start, err)
		}
	}

	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

func splitPatterns(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func matchAny(patterns []string, name string) bool {
	for _, p := range patterns {
		if ok, _ := filepath.Match(p, name); ok {
			return true
		}
	}
	return false
}

type tnode struct {
	name     string
	children map[string]*tnode
	isFile   bool
}

func newNode(name string) *tnode {
	return &tnode{
		name:     name,
		children: make(map[string]*tnode),
	}
}

func insertPath(root *tnode, rel string) {
	parts := splitPath(rel)
	cur := root
	for i, part := range parts {
		n, ok := cur.children[part]
		if !ok {
			n = newNode(part)
			cur.children[part] = n
		}
		if i == len(parts)-1 {
			n.isFile = true
		}
		cur = n
	}
}

func splitPath(p string) []string {
	// Ensure we split using OS separator
	p = filepath.Clean(p)
	return strings.Split(p, string(filepath.Separator))
}

func renderTree(paths []string) string {
	root := newNode("")
	for _, p := range paths {
		insertPath(root, p)
	}

	var b strings.Builder
	// top-level entries
	names := sortedKeys(root.children, true)
	for i, name := range names {
		child := root.children[name]
		last := i == len(names)-1
		renderNode(&b, child, "", last)
	}
	return b.String()
}

func renderNode(b *strings.Builder, n *tnode, prefix string, isLast bool) {
	branch := "├── "
	nextPrefix := prefix + "│   "
	if isLast {
		branch = "└── "
		nextPrefix = prefix + "    "
	}
	if isDir(n) {
		fmt.Fprintf(b, "%s%s%s/\n", prefix, branch, n.name)
		// sort children: directories first, then files, each alphabetical
		names := sortedKeys(n.children, true)
		for i, name := range names {
			child := n.children[name]
			last := i == len(names)-1
			renderNode(b, child, nextPrefix, last)
		}
	} else {
		fmt.Fprintf(b, "%s%s%s\n", prefix, branch, n.name)
	}
}

func isDir(n *tnode) bool {
	// a node is a directory if it has children; leaf nodes are files
	return len(n.children) > 0 && !n.isFile
}

func sortedKeys(m map[string]*tnode, dirsFirst bool) []string {
	if !dirsFirst {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	}
	var dirs, files []string
	for k, v := range m {
		if isDir(v) {
			dirs = append(dirs, k)
		} else {
			files = append(files, k)
		}
	}
	sort.Strings(dirs)
	sort.Strings(files)
	return append(dirs, files...)
}

func detectLang(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".php":
		return "php"
	case ".twig":
		return "twig"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "md"
	default:
		return ""
	}
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
