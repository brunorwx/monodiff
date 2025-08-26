package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/brunorwx/monodiff/internal/analyzer"
	"github.com/brunorwx/monodiff/internal/git"
	"github.com/brunorwx/monodiff/internal/workspace"
)

func main() {
	from := flag.String("from", "", "git ref to compare from (required)")
	to := flag.String("to", "HEAD", "git ref to compare to")
	format := flag.String("format", "text", "output format: text|json")
	root := flag.String("root", ".", "repo root (default .)")

	flag.Parse()

	if *from == "" {
		log.Fatalf("--from is required (example: --from=main)")
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		log.Fatalf("failed to resolve root: %v", err)
	}

	files, err := git.ChangedFiles(absRoot, *from, *to)
	if err != nil {
		log.Fatalf("git diff failed: %v", err)
	}

	pkgs, err := workspace.DiscoverPackages(absRoot)
	if err != nil {
		log.Fatalf("discover packages failed: %v", err)
	}

	changedPkgs := analyzer.MapFilesToPackages(files, pkgs)
	impacted := analyzer.ComputeImpact(pkgs, changedPkgs)

	out := struct {
		From     string              `json:"from"`
		To       string              `json:"to"`
		Changed  []string            `json:"changed"`
		Impacted []string            `json:"impacted"`
		Graph    map[string][]string `json:"graph"`
	}{
		From:     *from,
		To:       *to,
		Changed:  changedPkgs.List(),
		Impacted: impacted.List(),
		Graph:    workspace.GraphAdjacency(pkgs),
	}

	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}

	// text output
	fmt.Printf("Changed packages:\n")
	for _, p := range out.Changed {
		fmt.Printf(" - %s\n", p)
	}
	fmt.Printf("\nImpacted packages (transitive):\n")
	for _, p := range out.Impacted {
		fmt.Printf(" - %s\n", p)
	}
}
