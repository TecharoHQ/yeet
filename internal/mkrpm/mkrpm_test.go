package mkrpm

import (
	"os"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	"github.com/cavaliergopher/rpm"
)

func TestBuild(t *testing.T) {
	fname := yeettest.BuildHello(t, Build)

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
