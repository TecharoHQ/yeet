package mkdeb

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TecharoHQ/yeet/internal/gpgtest"
	"github.com/TecharoHQ/yeet/internal/yeet"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	"pault.ag/go/debian/deb"
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
	debFile, close, err := deb.LoadFile(fname)
	if err != nil {
		t.Fatalf("failed to load deb file: %v", err)
	}
	defer close()

	if debFile.Control.Version.Empty() {
		t.Error("version is empty")
	}
}

func TestBuildError(t *testing.T) {
	yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version: ".0.0",
		Fatal:   false,
	})
}

func TestEndToEndInstall(t *testing.T) {
	os := "linux"
	for _, cpu := range yeettest.Arches {
		if cpu == "riscv64" {
			t.Skip("linux/riscv64 is not supported by this test")
		}

		platform := fmt.Sprintf("%s/%s", os, cpu)
		t.Run(platform, func(t *testing.T) {
			fname := yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
				Version: "1.0.0",
				Fatal:   true,
				GOOS:    os,
				GOARCH:  cpu,
			})
			pkgName := filepath.Base(fname)

			t.Log(filepath.Base(fname), t.Name())

			yeettest.RunScript(t, t.Context(), "docker", "pull", "--platform", platform, "debian:bookworm")

			containerID, err := yeet.Output(t.Context(), "docker", "run", "-dit", "--platform", platform, "debian:bookworm", "sleep", "inf")
			if err != nil {
				t.Fatal(err)
			}

			containerID = strings.TrimSpace(containerID)
			t.Cleanup(func() {
				yeettest.RunScript(t, context.Background(), "docker", "rm", "-f", containerID)
			})

			yeettest.RunScript(t, t.Context(), "docker", "cp", fname, fmt.Sprintf("%s:/tmp/%s", containerID, pkgName))
			yeettest.RunScript(t, t.Context(), "docker", "exec", "-t", containerID, "dpkg", "-i", fmt.Sprintf("/tmp/%[1]s", pkgName))
			yeettest.RunScript(t, t.Context(), "docker", "exec", "-t", containerID, "/usr/bin/yeet-hello")
		})
	}
}
