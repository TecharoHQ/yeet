package yeettest

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

func HelloFixture(t *testing.T, version, goos, goarch string, yeetBuild *func(pkgmeta.BuildInput)) pkgmeta.Package {
	t.Helper()

	if version == "" {
		version = "0.0.1-test"
	}

	if goos == "" && goarch == "" {
		goos, goarch = runtime.GOOS, runtime.GOARCH
	}

	defaultYeetBuild := func(p pkgmeta.BuildInput) {
		yeet.ShouldWork(t.Context(), nil, yeet.WD, "go", "build", "-o", filepath.Join(p.Bin, "yeet-hello"), "../testdata/hello")
	}

	var yeetBuildFn *func(pkgmeta.BuildInput)
	if yeetBuild == nil {
		yeetBuildFn = &defaultYeetBuild
	} else {
		wrappedYeetBuild := func(p pkgmeta.BuildInput) {
			defaultYeetBuild(p)
			(*yeetBuild)(p)
		}
		yeetBuildFn = &wrappedYeetBuild
	}

	return pkgmeta.Package{
		Name:        "hello",
		Version:     version,
		Description: "Hello world",
		Homepage:    "https://example.com",
		License:     "MIT",
		Platform:    goos,
		Goarch:      goarch,
		Build:       *yeetBuildFn,
	}
}
