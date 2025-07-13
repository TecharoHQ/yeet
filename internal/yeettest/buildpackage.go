package yeettest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

type Impl func(p pkgmeta.Package) (string, error)

func BuildHello(t *testing.T, build Impl, version, keyFname, keyID string, fatal bool) string {
	t.Helper()

	dir := t.TempDir()
	internal.GPGKeyFile = &keyFname
	internal.GPGKeyID = &keyID
	internal.PackageDestDir = &dir

	p := pkgmeta.Package{
		Name:        "hello",
		Version:     version,
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
	switch fatal {
	case true:
		if err != nil {
			t.Fatalf("Build() error = %v", err)
		}
	case false:
		if err != nil {
			t.Logf("Build() error = %v", err)
		}
		return ""
	}

	if foutpath == "" {
		t.Fatal("Build() returned empty path")
	}

	t.Cleanup(func() {
		os.RemoveAll(filepath.Dir(foutpath))
	})

	return foutpath
}
