package analyzer

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/brunorwx/monodiff/internal/graph"
	"github.com/brunorwx/monodiff/internal/workspace"
)

// simple set type for convenience
type StrSet map[string]struct{}

func (s StrSet) Add(v string) { s[v] = struct{}{} }
func (s StrSet) Has(v string) bool {
	_, ok := s[v]
	return ok
}
func (s StrSet) List() []string {
	var out []string
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MapFilesToPackages maps changed files to owning package(s) by longest-prefix path match
func MapFilesToPackages(files []string, pkgs map[string]*workspace.Package) StrSet {
	owned := StrSet{}
	// Build list of package paths
	type pinfo struct {
		name string
		path string
	}
	var pi []pinfo
	for n, p := range pkgs {
		pi = append(pi, pinfo{name: n, path: filepath.Clean(p.Path)})
	}
	// for each file, find the package with the longest path prefix match
	for _, f := range files {
		fclean := filepath.Clean(f)
		best := ""
		bestLen := -1
		for _, p := range pi {
			// if package path is "." or "", it owns everything at root
			pth := p.path
			if pth == "." || pth == "" {
				pth = ""
			}
			if pth == "" {
				// root package, match any file
				if 0 > bestLen {
					best = p.name
					bestLen = 0
				}
				continue
			}
			// check prefix (path separators normalized)
			if strings.HasPrefix(fclean, pth+string(filepath.Separator)) || fclean == pth {
				if len(pth) > bestLen {
					best = p.name
					bestLen = len(pth)
				}
			}
		}
		if best != "" {
			owned.Add(best)
		}
	}
	return owned
}

// ComputeImpact computes transitive dependents (reverse graph BFS)
func ComputeImpact(pkgs map[string]*workspace.Package, changed StrSet) StrSet {
	// build graph adjacency for local deps
	g := graph.New()
	local := map[string]bool{}
	for n := range pkgs {
		local[n] = true
	}
	for name, p := range pkgs {
		for dep := range p.Dependencies {
			if local[dep] {
				g.AddEdge(name, dep) // name -> dep
			}
		}
	}
	rev := g.Reverse() // dep -> dependent
	// BFS from changed nodes on reverse graph
	result := StrSet{}
	queue := []string{}
	for n := range changed {
		queue = append(queue, n)
		result.Add(n)
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, dep := range rev.Adj[cur] {
			if !result.Has(dep) {
				result.Add(dep)
				queue = append(queue, dep)
			}
		}
	}
	return result
}
