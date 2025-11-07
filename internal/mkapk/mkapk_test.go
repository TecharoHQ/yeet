package mkapk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TecharoHQ/yeet/internal/yeet"
	"github.com/TecharoHQ/yeet/internal/yeettest"
	apk "gitlab.alpinelinux.org/alpine/go/repository"
)

func TestBuild(t *testing.T) {
	keyFname := filepath.Join(t.TempDir(), "foo.key")
	err := yeettest.GenerateRSAKey(keyFname)
	if err != nil {
		t.Fatal(err)
	}

	fname := yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
		Version:  "1.0.0",
		KeyFname: keyFname,
		Fatal:    true,
	})
	apkFile, err := os.Open(fname)
	if err != nil {
		t.Fatalf("failed to open apk file: %v", err)
	}

	apkPackage, err := apk.ParsePackage(apkFile)
	if err != nil {
		t.Fatalf("failed to load apk package: %v", err)
	}
	defer apkFile.Close()

	if 0 == len(apkPackage.Version) {
		t.Error("version is empty")
	}
}

// // apk fails, unlike deb/rpm
// func TestBuildError(t *testing.T) {
// 	yeettest.BuildHello(t, Build, yeettest.BuildHelloInput{
// 		Version: ".0.0",
// 		Fatal:   false,
// 	})
// }

func TestEndToEndInstall(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not installed: %v", err)
	}

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

			yeettest.RunScript(t, t.Context(), "docker", "pull", "--platform", platform, "alpine:edge")

			containerID, err := yeet.Output(t.Context(), "docker", "run", "-dit", "--platform", platform, "alpine:edge", "sleep", "inf")
			if err != nil {
				t.Fatal(err)
			}

			containerID = strings.TrimSpace(containerID)
			t.Cleanup(func() {
				yeettest.RunScript(t, context.Background(), "docker", "rm", "-f", containerID)
			})

			yeettest.RunScript(t, t.Context(), "docker", "cp", fname, fmt.Sprintf("%s:/tmp/%s", containerID, pkgName))
			yeettest.RunScript(t, t.Context(), "docker", "exec", "-t", containerID, "apk", "add", "--allow-untrusted", fmt.Sprintf("/tmp/%[1]s", pkgName))
			yeettest.RunScript(t, t.Context(), "docker", "exec", "-t", containerID, "/usr/bin/yeet-hello")
		})
	}
}
