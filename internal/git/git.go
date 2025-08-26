package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func ChangedFiles(repoRoot, from, to string) ([]string, error) {
	diffRef := fmt.Sprintf("%s..%s", from, to)
	cmd := exec.Command("git", "-C", repoRoot, "diff", "--name-only", diffRef)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff error: %w: %s", err, out.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		result = append(result, filepath.Clean(l))
	}
	return result, nil
}
