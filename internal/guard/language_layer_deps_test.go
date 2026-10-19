package guard_test

import (
	"path/filepath"
	"strings"
	"testing"
)

// languageLayerRoots are the public entry points of the reusable language
// layer. Their transitive project-local import closures must never contain
// analyzer packages, so the language layer stays consumable without the
// analyzer, its embedded catalogs, or its GPL-licensed NRT data.
var languageLayerRoots = []string{"kustolang", "kustolang/environment"}

var forbiddenForLanguageLayer = []string{"analyzer", "internal/analyzer"}

func TestLanguageLayerTransitiveDeps(t *testing.T) {
	for _, root := range languageLayerRoots {
		visited := map[string]bool{}
		var walk func(dir string, chain []string)
		walk = func(dir string, chain []string) {
			if visited[dir] {
				return
			}
			visited[dir] = true
			chain = append(chain, dir)
			for _, forbidden := range forbiddenForLanguageLayer {
				if dir == forbidden {
					t.Errorf("language layer root %s transitively imports %s (chain: %s)",
						root, forbidden, strings.Join(chain, " -> "))
					return
				}
			}
			for _, imported := range projectImports(t, filepath.Join(repoRoot, dir)) {
				walk(imported, chain)
			}
		}
		walk(root, nil)
	}
}
