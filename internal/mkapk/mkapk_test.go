package mkapk

import (
	"os"
	"testing"

	"github.com/TecharoHQ/yeet/internal/yeettest"
	"github.com/chainguard-dev/go-apk/pkg/apk"
)

func TestBuild(t *testing.T) {
	path := yeettest.BuildHello(t, Build, "1.0.0", true)

	fin, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fin.Close()

	pkg, err := apk.ParsePackage(t.Context(), fin)
	if err != nil {
		t.Fatal(err)
	}

	if pkg.Version != "1.0.0" {
		t.Errorf("got wrong version %s, wanted 1.0.0", pkg.Version)
	}
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, ".0.0", false)
}
