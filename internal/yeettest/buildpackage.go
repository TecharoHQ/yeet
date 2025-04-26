package yeettest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

type Impl func(p pkgmeta.Package) (string, error)

func BuildHello(t *testing.T, build Impl) string {
	t.Helper()

	p := pkgmeta.Package{
		Name:        "hello",
		Version:     "1.0.0",
		Description: "Hello world",
		Homepage:    "https://example.com",
		License:     "MIT",
		Platform:    runtime.GOOS,
		Goarch:      runtime.GOARCH,
		Build: func(p pkgmeta.BuildInput) {
			yeet.ShouldWork(t.Context(), nil, yeet.WD, "go", "build", "-o", filepath.Join(p.Bin, "hello"), "../testdata/hello")
		},
	}

	foutpath, err := build(p)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if foutpath == "" {
		t.Fatal("Build() returned empty path")
	}

	t.Cleanup(func() {
		os.RemoveAll(filepath.Dir(foutpath))
	})

	return foutpath
}
