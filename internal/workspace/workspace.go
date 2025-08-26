package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Package struct {
	Name         string            `json:"name"`
	Path         string            `json:"path"` // relative to repo root
	Dependencies map[string]string `json:"dependencies"`
}

func DiscoverPackages(root string) (map[string]*Package, error) {
	var pkgJsonPaths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
		}

		if !d.IsDir() && strings.EqualFold(filepath.Base(path), "package.json") {

			rel, _ := filepath.Rel(root, path)
			pkgJsonPaths = append(pkgJsonPaths, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := map[string]*Package{}
	var mu sync.Mutex

	maxWorkers := runtime.GOMAXPROCS(0)
	if maxWorkers <= 0 {
		maxWorkers = 4
	}

	const maxCap = 12
	if maxWorkers > maxCap {
		maxWorkers = maxCap
	}

	g, _ := errgroup.WithContext(context.Background())
	sem := semaphore.NewWeighted(int64(maxWorkers))

	for _, relPath := range pkgJsonPaths {
		relPath := relPath // capture
		if err := sem.Acquire(context.Background(), 1); err != nil {
			return nil, err
		}
		g.Go(func() error {
			defer sem.Release(1)
			full := filepath.Join(root, relPath)
			data, err := os.ReadFile(full)
			if err != nil {
				return fmt.Errorf("read package.json %s: %w", full, err)
			}
			var pj struct {
				Name         string            `json:"name"`
				Dependencies map[string]string `json:"dependencies"`
				DevDeps      map[string]string `json:"devDependencies"`
				PeerDeps     map[string]string `json:"peerDependencies"`
			}
			if err := json.Unmarshal(data, &pj); err != nil {
				return fmt.Errorf("invalid package.json at %s: %w", full, err)
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

			dir := filepath.Dir(relPath)
			if dir == "." {
				dir = ""
			}
			name := pj.Name
			if name == "" {

				if dir == "" {
					name = filepath.Base(root)
				} else {
					name = filepath.Base(dir)
				}
			}
			mu.Lock()
			out[name] = &Package{
				Name:         name,
				Path:         dir,
				Dependencies: deps,
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func GraphAdjacency(pkgs map[string]*Package) map[string][]string {
	adj := map[string][]string{}

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
