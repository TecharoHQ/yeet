package yeet

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

func TestBuildOwnPackages(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("Skipping test in non-CI environment")
	}

	type packageJSON struct {
		Version string `json:"version"`
	}

	fin, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("can't read package.json: %v", err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(fin, &pkg); err != nil {
		t.Fatalf("can't unmarshal package.json: %v", err)
	}

	dir := t.TempDir()
	internal.PackageDestDir = &dir
	yeet.ShouldWork(t.Context(), nil, yeet.WD, "go", "run", "./cmd/yeet", "--force-git-version", pkg.Version, "--package-dest-dir", t.TempDir())
}
