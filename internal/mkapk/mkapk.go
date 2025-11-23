package mkapk

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/apk"
	"github.com/goreleaser/nfpm/v2/files"
)

// build an alpine-style apk
func Build(p pkgmeta.Package) (foutpath string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				slog.Error("mkapk: error while building", "err", err)
			} else {
				err = fmt.Errorf("%v", r)
				slog.Error("mkapk: error while building", "err", err)
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

	dir, err := os.MkdirTemp("", "yeet-mkapk")
	if err != nil {
		return "", fmt.Errorf("mkapk: can't make temporary directory")
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
	os.Setenv("GOOS", "linux")
	os.Setenv("CGO_ENABLED", "0")

	p.Build(pkgmeta.BuildInput{
		Output: dir,
		Bin:    filepath.Join(dir, "usr", "bin"),
		Doc:    filepath.Join(dir, "usr", "share", "doc", p.Name),
		Etc:    filepath.Join(dir, "etc", p.Name),
		Man:    filepath.Join(dir, "usr", "share", "man"),
		// some APK-based distributions support systemd
		Systemd: filepath.Join(dir, "usr", "lib", "systemd", "system"),
		Openrc: &pkgmeta.Openrc{
			InitDir: filepath.Join(dir, "etc", "init.d"),
			ConfDir: filepath.Join(dir, "etc", "conf.d"),
		},
	})

	var contents files.Contents

	contents = slices.Concat(contents, p.CopyEmptyDirs(0000)) // default mode
	contents = slices.Concat(contents, p.CopyConfigFiles())
	contents = slices.Concat(contents, p.CopyDocumentation())
	contents = slices.Concat(contents, p.CopyFiles())

	cs, err := p.CopyTree(dir)
	if err != nil {
		return "", fmt.Errorf("mkapk: can't walk output directory: %w", err)
	}
	contents = slices.Concat(contents, cs)

	contents, err = files.PrepareForPackager(contents, 0o002, "apk", true, internal.SourceEpoch())
	if err != nil {
		return "", fmt.Errorf("mkapk: can't prepare for packager: %w", err)
	}

	for _, content := range contents {
		content.FileInfo.MTime = internal.SourceEpoch()
	}

	info := nfpm.WithDefaults(&nfpm.Info{
		Name:        p.Name,
		Version:     p.Version,
		Arch:        p.Goarch,
		Platform:    "linux",
		Description: p.Description,
		Maintainer:  fmt.Sprintf("%s <%s>", *internal.UserName, *internal.UserEmail),
		Homepage:    p.Homepage,
		License:     p.License,
		MTime:       internal.SourceEpoch(),
		Overridables: nfpm.Overridables{
			Contents:   contents,
			Depends:    p.Depends,
			Recommends: p.Recommends,
			Replaces:   p.Replaces,
			Conflicts:  p.Replaces,
		},
	})

	if *internal.GPGKeyID != "" {
		return "", fmt.Errorf("cannot specify GPG key for APK package")
	}

	if *internal.APKKeyFile != "" {
		slog.Debug("using APK signing key", "file", *internal.APKKeyFile, "name", *internal.APKKeyName)
		info.Overridables.APK.Signature.KeyFile = *internal.APKKeyFile
		info.Overridables.APK.Signature.KeyPassphrase = *internal.APKKeyPassword
		info.Overridables.APK.Signature.KeyName = *internal.APKKeyName
	}

	pkg, err := nfpm.Get("apk")
	if err != nil {
		return "", fmt.Errorf("mkapk: can't get apk packager: %w", err)
	}

	foutpath = pkg.ConventionalFileName(info)
	fout, err := os.Create(filepath.Join(*internal.PackageDestDir, foutpath))
	if err != nil {
		return "", fmt.Errorf("mkapk: can't create output file: %w", err)
	}
	defer fout.Close()

	if err := pkg.Package(info, fout); err != nil {
		return "", fmt.Errorf("mkapk: can't build package: %w", err)
	}

	slog.Info("built package", "name", p.Name, "arch", p.Goarch, "version", p.Version, "path", fout.Name())

	return fout.Name(), err
}
