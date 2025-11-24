package mkportable

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/TecharoHQ/yeet"
	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/erofs"
	"github.com/TecharoHQ/yeet/internal/mktarball"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
)

const confextReleaseTemplate = `ID=_any
VERSION_ID=_any
EXTENSION_RELOAD_MANAGER=1
`

const sysextReleaseTemplate = `ID=_any
VERSION_ID=_any
EXTENSION_RELOAD_MANAGER=1
ARCHITECTURE={{.Arch}}
`

func Confext(p pkgmeta.Package) (foutpath string, err error) {
	oldBuild := p.Build
	buildFunc := func(bi pkgmeta.BuildInput) {
		oldBuild(bi)

		os.MkdirAll(filepath.Join(bi.Output, "etc", "extension-release.d"), 0755)
		os.WriteFile(filepath.Join(bi.Output, "etc", "extension-release.d", "extension-release."+p.Name), []byte(confextReleaseTemplate), 0666)
	}

	p.Build = buildFunc

	return build(p)
}

// Maps GOARCH (key) to Systemd Architecture (value)
var goToSystemd = map[string]string{
	"amd64":    "x86-64",
	"386":      "x86",
	"arm":      "arm",
	"arm64":    "arm64",
	"loong64":  "loongarch64",
	"mips64":   "mips64",
	"mips64le": "mips64-le",
	"ppc64":    "ppc64",
	"ppc64le":  "ppc64-le",
	"riscv64":  "riscv64",
	"s390x":    "s390x",
}

func Sysext(p pkgmeta.Package) (foutpath string, err error) {
	oldBuild := p.Build
	buildFunc := func(bi pkgmeta.BuildInput) {
		oldBuild(bi)

		os.MkdirAll(filepath.Join(bi.Output, "usr", "lib", "extension-release.d"), 0755)
		fout, err := os.Create(filepath.Join(bi.Output, "usr", "lib", "extension-release.d", "extension-release."+p.Name))
		if err != nil {
			panic(err) // caught by upstream
		}
		defer fout.Close()

		if err := fout.Chmod(0666); err != nil {
			panic(err) // caught by upstream
		}

		t := template.Must(template.New("extension-release").Parse(sysextReleaseTemplate))
		if err := t.Execute(fout, map[string]string{
			"Arch": goToSystemd[p.Goarch],
		}); err != nil {
			panic(err) // caught by upstream
		}

		for _, d := range p.EmptyDirs {
			if err := os.MkdirAll(filepath.Join(bi.Output, d), 0755); err != nil {
				panic(err)
			}
		}

		for src, dst := range p.ConfigFiles {
			if err := mktarball.Copy(src, filepath.Join(bi.Output, dst)); err != nil {
				panic(err)
			}
		}

		for src, dst := range p.Documentation {
			if err := mktarball.Copy(src, filepath.Join(bi.Doc, dst)); err != nil {
				panic(err)
			}
		}

		for src, dst := range p.Files {
			if err := mktarball.Copy(src, filepath.Join(bi.Output, dst)); err != nil {
				panic(err)
			}
		}
	}

	p.Build = buildFunc

	return build(p)
}

func Portable(p pkgmeta.Package) (foutpath string, err error) {
	oldBuild := p.Build
	buildFunc := func(bi pkgmeta.BuildInput) {
		oldBuild(bi)

		os.MkdirAll(filepath.Join(bi.Output, "usr", "lib"), 0755)
		os.WriteFile(filepath.Join(bi.Output, "usr", "lib", "os-release"), []byte(fmt.Sprintf("BUILD_ID=%s\nID=yeet-minimal\nPORTABLE_ID=%s\nPORTABLE_PRETTY_NAME=%s\nPRETTY_NAME=Yeet Minimal\n", yeet.Version, p.Name, p.Name)), 0444)

		os.MkdirAll(filepath.Join(bi.Output, "etc"), 0755)
		os.WriteFile(filepath.Join(bi.Output, "etc", "resolv.conf"), nil, 0444)
		os.WriteFile(filepath.Join(bi.Output, "etc", "machine-id"), nil, 0444)
		os.MkdirAll(filepath.Join(bi.Output, "etc"), 0755)
		os.MkdirAll(filepath.Join(bi.Output, "proc"), 0755)
		os.MkdirAll(filepath.Join(bi.Output, "sys"), 0755)
		os.MkdirAll(filepath.Join(bi.Output, "dev"), 0755)
		os.MkdirAll(filepath.Join(bi.Output, "run"), 0755)
		os.MkdirAll(filepath.Join(bi.Output, "tmp"), 0755)
		os.MkdirAll(filepath.Join(bi.Output, "var", "tmp"), 0755)

		for _, d := range p.EmptyDirs {
			if err := os.MkdirAll(filepath.Join(bi.Output, d), 0755); err != nil {
				panic(err)
			}
		}

		for src, dst := range p.ConfigFiles {
			if err := mktarball.Copy(src, filepath.Join(bi.Output, dst)); err != nil {
				panic(err)
			}
		}

		for src, dst := range p.Documentation {
			if err := mktarball.Copy(src, filepath.Join(bi.Doc, dst)); err != nil {
				panic(err)
			}
		}

		for src, dst := range p.Files {
			if err := mktarball.Copy(src, filepath.Join(bi.Output, dst)); err != nil {
				panic(err)
			}
		}
	}

	p.Build = buildFunc

	return build(p)
}

func build(p pkgmeta.Package) (foutpath string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				slog.Error("mkportable: error while building", "err", err)
			} else {
				err = fmt.Errorf("%v", r)
				slog.Error("mkportable: error while building", "err", err)
			}
		}
	}()

	os.MkdirAll(*internal.PackageDestDir, 0755)
	os.WriteFile(filepath.Join(*internal.PackageDestDir, ".gitignore"), []byte("*\n!.gitignore"), 0644)

	if p.Version == "" {
		p.Version = internal.GitVersion()
	}

	if _, err := semver.NewVersion(p.Version); err != nil {
		return "", fmt.Errorf("invalid version %q: %w", p.Version, err)
	}

	if p.Platform == "" {
		p.Platform = "linux"
	}

	dir, err := os.MkdirTemp("", "yeet-mkportable")
	if err != nil {
		return "", fmt.Errorf("mkportable: can't make temporary directory")
	}
	defer os.RemoveAll(dir)
	os.MkdirAll(dir, 0755)

	cgoEnabled := os.Getenv("CGO_ENABLED")
	defer func() {
		os.Setenv("GOARCH", runtime.GOARCH)
		os.Setenv("GOOS", runtime.GOOS)
		os.Setenv("CGO_ENABLED", cgoEnabled)
	}()
	os.Setenv("GOARCH", p.Goarch)
	os.Setenv("GOOS", p.Platform)
	os.Setenv("CGO_ENABLED", "0")

	bi := pkgmeta.BuildInput{
		Output:  dir,
		Bin:     filepath.Join(dir, "usr", "bin"),
		Doc:     filepath.Join(dir, "usr", "share", "doc", p.Name),
		Etc:     filepath.Join(dir, "etc", p.Name),
		Man:     filepath.Join(dir, "usr", "share", "man"),
		Systemd: filepath.Join(dir, "usr", "lib", "systemd", "system"),
	}

	p.Build(bi)

	foutpath = filepath.Join(*internal.PackageDestDir, fmt.Sprintf("%s_%s_%s.raw", p.Name, p.Version, p.Goarch))
	fout, err := os.Create(foutpath)
	if err != nil {
		return "", err
	}

	if err := erofs.Create(fout, os.DirFS(dir)); err != nil {
		return "", err
	}

	slog.Info("built package", "name", p.Name, "arch", p.Goarch, "version", p.Version, "path", fout.Name())

	return fout.Name(), err
}
