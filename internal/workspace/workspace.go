package workspace

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Package represents a workspace package
type Package struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"` // relative to repo root
	Dependencies map[string]string `json:"dependencies"`
}

// DiscoverPackages scans the repo root for package.json files and returns a map[name]*Package
func DiscoverPackages(root string) (map[string]*Package, error) {
	pkgs := map[string]*Package{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// skip node_modules and .git
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
		}

		if strings.EqualFold(filepath.Base(path), "package.json") {
			rel, _ := filepath.Rel(root, path)
			dir := filepath.Dir(rel)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var pj struct {
				Name         string            `json:"name"`
				Dependencies map[string]string `json:"dependencies"`
				DevDeps      map[string]string `json:"devDependencies"`
				PeerDeps     map[string]string `json:"peerDependencies"`
			}
			if err := json.Unmarshal(data, &pj); err != nil {
				return fmt.Errorf("invalid package.json at %s: %w", path, err)
			}
			deps := map[string]string{}
			for k, v := range pj.Dependencies {
				deps[k] = v
			}
			for k, v := range pj.DevDeps {
				deps[k] = v
			}
			for k, v := range pj.PeerDeps {
				deps[k] = v
			}
			name := pj.Name
			if name == "" {
				// fallback to path name
				name = dir
			}
			pkgs[name] = &Package{
				Name:         name,
				Path:         dir,
				Dependencies: deps,
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}

// GraphAdjacency returns adjacency map for JSON output: package -> local deps (names)
func GraphAdjacency(pkgs map[string]*Package) map[string][]string {
	adj := map[string][]string{}
	// build reverse map from dependency name -> isLocal
	localNames := map[string]bool{}
	for n := range pkgs {
		localNames[n] = true
	}
	for name, p := range pkgs {
		for dep := range p.Dependencies {
			if localNames[dep] {
				adj[name] = append(adj[name], dep)
			}
		}
	}
	return adj
}
