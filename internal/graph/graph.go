package graph

// Simple directed graph representation using adjacency lists
type Graph struct {
	Adj map[string][]string
}

func New() *Graph {
	return &Graph{Adj: map[string][]string{}}
}

func (g *Graph) AddEdge(from, to string) {
	g.Adj[from] = append(g.Adj[from], to)
}

func (g *Graph) Reverse() *Graph {
	r := New()
	for u, outs := range g.Adj {
		for _, v := range outs {
			r.AddEdge(v, u)
		}
	}
	return r
}
