package yeet

import (
	"os"
	"testing"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

func TestBuildOwnPackages(t *testing.T) {
	if os.Getenv("CI") == "" {
		t.Skip("Skipping test in non-CI environment")
	}

	dir := t.TempDir()
	internal.PackageDestDir = &dir
	yeet.ShouldWork(t.Context(), nil, yeet.WD, "go", "run", "./cmd/yeet")
}
