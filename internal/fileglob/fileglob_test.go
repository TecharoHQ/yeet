package fileglob

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatch(t *testing.T) {
	for _, tt := range []struct {
		pattern string
		name    string
		want    bool
	}{
		{"**/*.go", "main.go", true},
		{"**/*.go", "cmd/yeet/main.go", true},
		{"**/*.go", "README.md", false},
		{"cmd/**/*.go", "cmd/yeet/main.go", true},
		{"cmd/**/*.go", "internal/foo.go", false},
		{"*.go", "main.go", true},
		{"*.go", "cmd/main.go", false},
		{"**/*", "a/b/c", true},
		{"**/b/*", "a/b/c", true},
		{"a/**/c", "a/c", true},
		{"a/**/c", "a/b/c", true},
		{"a/**/c", "a/b/d/c", true},
		{"[invalid", "x", false}, // filepath.Match returns error
	} {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			got, err := Match(tt.pattern, tt.name)
			if tt.pattern == "[invalid" {
				if err == nil {
					t.Fatal("expected error for invalid pattern")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestGlob(t *testing.T) {
	// Build a temp tree to glob against.
	tmp := t.TempDir()

	files := []string{
		"a.go",
		"b.txt",
		"src/main.go",
		"src/util/helpers.go",
		"src/util/helpers_test.go",
		"docs/readme.md",
	}
	for _, f := range files {
		p := filepath.Join(tmp, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// chdir so relative patterns work against our temp tree
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restoring working directory: %v", err)
		}
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%s): %v", tmp, err)
	}

	for _, tt := range []struct {
		name    string
		pattern string
		want    int // expected match count
	}{
		{"recursive go files", "**/*.go", 4},
		{"scoped recursive", "src/**/*.go", 3},
		{"single level", "*.go", 1},
		{"no matches", "**/*.rs", 0},
		{"non-recursive subdir", "src/*.go", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			matches, err := Glob(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) != tt.want {
				t.Errorf("Glob(%q): got %d matches %v, want %d", tt.pattern, len(matches), matches, tt.want)
			}
		})
	}
}

func TestGlobInvalidPattern(t *testing.T) {
	_, err := Glob("[invalid")
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestStaticPrefix(t *testing.T) {
	for _, tt := range []struct {
		pattern string
		want    string
	}{
		{"**/*.go", "."},
		{"src/**/*.go", "src"},
		{"a/b/c/*.go", "a/b/c"},
		{"*.go", "."},
		{"a/b/**/c", "a/b"},
	} {
		t.Run(tt.pattern, func(t *testing.T) {
			got := staticPrefix(tt.pattern)
			if got != tt.want {
				t.Errorf("staticPrefix(%q) = %q, want %q", tt.pattern, got, tt.want)
			}
		})
	}
}
