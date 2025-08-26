package analyzer

import (
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/brunorwx/monodiff/internal/workspace"
)

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

func MapFilesToPackages(files []string, pkgs map[string]*workspace.Package) StrSet {
	owned := StrSet{}

	type pinfo struct {
		name string
		path string
	}
	pi := make([]pinfo, 0, len(pkgs))
	for n, p := range pkgs {
		pi = append(pi, pinfo{name: n, path: filepath.Clean(p.Path)})
	}

	for _, f := range files {
		fclean := filepath.Clean(f)
		best := ""
		bestLen := -1
		for _, p := range pi {
			pth := p.path
			if pth == "." || pth == "" {

				if 0 > bestLen {
					best = p.name
					bestLen = 0
				}
				continue
			}

			pref := pth + string(filepath.Separator)
			if fclean == pth || strings.HasPrefix(fclean, pref) {
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

func ComputeImpact(pkgs map[string]*workspace.Package, changed StrSet) StrSet {
	n := len(pkgs)
	// map names -> id
	idOf := make(map[string]int, n)
	names := make([]string, 0, n)
	i := 0
	for name := range pkgs {
		idOf[name] = i
		names = append(names, name)
		i++
	}

	adj := make([][]int, n)
	for name, p := range pkgs {
		u := idOf[name]
		for dep := range p.Dependencies {
			if v, ok := idOf[dep]; ok {
				adj[u] = append(adj[u], v)
			}
		}
	}

	rev := make([][]int, n)
	for u := 0; u < n; u++ {
		for _, v := range adj[u] {
			rev[v] = append(rev[v], u)
		}
	}

	visited := make([]uint32, n)
	queue := make([]int, 0, len(changed))
	for name := range changed {
		if id, ok := idOf[name]; ok {
			if atomic.CompareAndSwapUint32(&visited[id], 0, 1) {
				queue = append(queue, id)
			}
		}
	}

	head := 0
	for head < len(queue) {
		cur := queue[head]
		head++
		for _, dep := range rev[cur] {
			if atomic.CompareAndSwapUint32(&visited[dep], 0, 1) {
				queue = append(queue, dep)
			}
		}
	}

	result := StrSet{}
	for id, v := range visited {
		if v == 1 {
			result.Add(names[id])
		}
	}
	return result
}
