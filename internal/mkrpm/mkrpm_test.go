package mkrpm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	"github.com/cavaliergopher/rpm"
)

func TestBuild(t *testing.T) {
	keyFname := filepath.Join(t.TempDir(), "foo.gpg")
	keyID, err := gpgtest.MakeKey(t.Context(), keyFname)
	if err != nil {
		t.Fatal(err)
	}

	fname := yeettest.BuildHello(t, Build, "1.0.0", keyFname, keyID, true)

	pkg, err := rpm.Open(fname)
	if err != nil {
		t.Fatalf("failed to open rpm file: %v", err)
	}

	version, err := semver.NewVersion(pkg.Version())
	if err != nil {
		t.Fatalf("failed to parse version: %v", err)
	}
	if version == nil {
		t.Error("version is nil")
	}

	fin, err := os.Open(fname)
	if err != nil {
		t.Fatalf("failed to open rpm file: %v", err)
	}
	defer fin.Close()
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, ".0.0", "", "", false)
}
