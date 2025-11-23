package mkapk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/TecharoHQ/yeet/internal/pkgmeta"
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

	if len(apkPackage.Version) == 0 {
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

	goos := "linux"
	for _, cpu := range yeettest.Arches {
		if cpu == "riscv64" {
			t.Skip("linux/riscv64 is not supported by this test")
		}

		platform := fmt.Sprintf("%s/%s", goos, cpu)
		t.Run(platform, func(t *testing.T) {
			// yeettest manipulates the internal.PackageDestDir global, t.Parallel will race
			fname := yeettest.BuildCustomHello(t, Build, yeettest.BuildHelloInput{
				Version: "1.0.0",
				Fatal:   true,
				GOOS:    goos,
				GOARCH:  cpu,
			}, func(bi pkgmeta.BuildInput) {
				err := errors.Join(
					os.MkdirAll(bi.Bin, 0755),
					os.MkdirAll(bi.Openrc.InitDir, 0755),
					os.WriteFile(filepath.Join(bi.Bin, "yeet-0666"), nil, 0666),
					os.WriteFile(filepath.Join(bi.Bin, "yeet-0777"), nil, 0777),
					os.WriteFile(filepath.Join(bi.Openrc.InitDir, "yeet-0666"), nil, 0666),
					os.WriteFile(filepath.Join(bi.Openrc.InitDir, "yeet-0777"), nil, 0777),
				)
				if err != nil {
					t.Errorf("failed to build test package: %s", err)
				}
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

			type testcase struct {
				name string
				args []string
			}

			testcases := []testcase{
				// built binaries run
				{name: "bin=yeet-hello", args: []string{"/usr/bin/yeet-hello"}},
				// other executables run
				{name: "bin=yeet-0777", args: []string{"test", "-x", "/usr/bin/yeet-0777"}},
				// executable permission fix
				{name: "bin=yeet-0666", args: []string{"test", "-x", "/usr/bin/yeet-0666"}},
				// service script
				{name: "initd=yeet-0777", args: []string{"test", "-x", "/etc/init.d/yeet-0777"}},
				// service script permission fix
				{name: "initd=yeet-0666", args: []string{"test", "-x", "/etc/init.d/yeet-0666"}},
			}

			for _, tc := range testcases {
				t.Run(tc.name, func(t *testing.T) {
					args := []string{"docker", "exec", "-t", containerID}
					yeettest.RunScript(t, t.Context(), slices.Concat(args, tc.args)...)
				})
			}
		})
	}
}
