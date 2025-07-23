package mkrpm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/yeet"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	"github.com/cavaliergopher/rpm"
)

func TestBuild(t *testing.T) {
	keyFname := filepath.Join(t.TempDir(), "foo.gpg")
	keyID, err := gpgtest.MakeKey(t.Context(), keyFname)
	if err != nil {
		t.Fatal(err)
	}

	fname := yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version:  "1.0.0",
		KeyFname: keyFname,
		KeyID:    keyID,
		Fatal:    true,
	})

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
	yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version: ".0.0",
		Fatal:   false,
	})
}

func TestEndToEndInstall(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not installed: %v", err)
	}

	os := "linux"
	for _, cpu := range yeettest.Arches {
		platform := fmt.Sprintf("%s/%s", os, cpu)
		t.Run(platform, func(t *testing.T) {
			if cpu == "386" {
				t.Skip("linux/386 is not supported by this test")
			}

			fname := yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
				Version: "1.0.0",
				Fatal:   true,
				GOOS:    os,
				GOARCH:  cpu,
			})
			pkgName := filepath.Base(fname)

			t.Log(filepath.Base(fname), t.Name())

			yeettest.RunScript(t, t.Context(), "docker", "pull", "--platform", platform, "rockylinux/rockylinux:10-ubi")

			containerID, err := yeet.Output(t.Context(), "docker", "run", "-dit", "--platform", platform, "rockylinux/rockylinux:10-ubi", "sleep", "inf")
			if err != nil {
				t.Fatal(err)
			}

			containerID = strings.TrimSpace(containerID)
			t.Cleanup(func() {
				yeettest.RunScript(t, context.Background(), "docker", "rm", "-f", containerID)
			})

			yeettest.RunScript(t, t.Context(), "docker", "cp", fname, fmt.Sprintf("%s:/tmp/%s", containerID, pkgName))
			yeettest.RunScript(t, t.Context(), "docker", "exec", "-t", containerID, "rpm", "-i", fmt.Sprintf("/tmp/%[1]s", pkgName))
			yeettest.RunScript(t, t.Context(), "docker", "exec", "-t", containerID, "/usr/bin/yeet-hello")
		})
	}
}
