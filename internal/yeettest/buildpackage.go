package yeettest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/TecharoHQ/yeet/internal/yeet"
)

var Arches = []string{"386", "amd64", "arm64", "ppc64le", "riscv64"}

type Impl func(p pkgmeta.Package) (string, error)

type BuildHelloInput struct {
	Version  string
	KeyFname string
	KeyID    string
	Fatal    bool

	GOOS, GOARCH string
}

func (bi BuildHelloInput) Platform() (os, cpu string) {
	if bi.GOOS != "" && bi.GOARCH != "" {
		return bi.GOOS, bi.GOARCH
	}

	return runtime.GOOS, runtime.GOARCH
}

func BuildHello(t *testing.T, build Impl, inp BuildHelloInput) string {
	t.Helper()

	goos, goarch := inp.Platform()

	version := inp.Version
	keyFname := inp.KeyFname
	keyID := inp.KeyID
	fatal := inp.Fatal

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
		Platform:    goos,
		Goarch:      goarch,
		Build: func(p pkgmeta.BuildInput) {
			yeet.ShouldWork(t.Context(), nil, yeet.WD, "go", "build", "-o", filepath.Join(p.Bin, "yeet-hello"), "../testdata/hello")
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

func RunScript(t *testing.T, ctx context.Context, args ...string) {
	t.Helper()

	var stdout, stderr []byte
	var err error
	backoff := 250 * time.Millisecond

	for attempt := 0; attempt < 5; attempt++ {
		t.Logf("Running command: %s", strings.Join(args, " "))
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)

		stdout, err = cmd.Output()
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
		}

		t.Logf("stdout: %s", stdout)
		t.Logf("stderr: %s", stderr)

		if err == nil {
			return
		}

		t.Logf("Attempt %d failed: %v", attempt+1, err)
		t.Logf("Retrying in %v...", backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	t.Fatalf("script failed after 5 attempts: %v", err)
}
