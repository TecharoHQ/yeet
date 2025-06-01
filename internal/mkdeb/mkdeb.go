package mkdeb

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/TecharoHQ/yeet/internal"
	"github.com/TecharoHQ/yeet/internal/pkgmeta"
	"github.com/goreleaser/nfpm/v2"
	_ "github.com/goreleaser/nfpm/v2/deb"
	"github.com/goreleaser/nfpm/v2/files"
)

func Build(p pkgmeta.Package) (foutpath string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				slog.Error("mkrpm: error while building", "err", err)
			} else {
				err = fmt.Errorf("%v", r)
				slog.Error("mkrpm: error while building", "err", err)
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

	dir, err := os.MkdirTemp("", "yeet-mkdeb")
	if err != nil {
		return "", fmt.Errorf("mkrpm: can't make temporary directory")
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
		Output:  dir,
		Bin:     filepath.Join(dir, "usr", "bin"),
		Doc:     filepath.Join(dir, "usr", "share", "doc", p.Name),
		Etc:     filepath.Join(dir, "etc", p.Name),
		Man:     filepath.Join(dir, "usr", "share", "man"),
		Systemd: filepath.Join(dir, "usr", "lib", "systemd", "system"),
	})

	var contents files.Contents

	for _, d := range p.EmptyDirs {
		if d == "" {
			continue
		}

		contents = append(contents, &files.Content{
			Type:        files.TypeDir,
			Destination: d,
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
				Mode:  os.FileMode(0600),
			},
		})
	}

	for repoPath, osPath := range p.ConfigFiles {
		contents = append(contents, &files.Content{
			Type:        files.TypeConfig,
			Source:      repoPath,
			Destination: osPath,
			FileInfo: &files.ContentFileInfo{
				Mode:  os.FileMode(0600),
				MTime: internal.SourceEpoch(),
			},
		})
	}

	for repoPath, rpmPath := range p.Documentation {
		contents = append(contents, &files.Content{
			Type:        files.TypeFile,
			Source:      repoPath,
			Destination: filepath.Join("/usr/share/doc", p.Name, rpmPath),
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
			},
		})
	}

	for repoPath, rpmPath := range p.Files {
		contents = append(contents, &files.Content{
			Type:        files.TypeFile,
			Source:      repoPath,
			Destination: rpmPath,
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
			},
		})
	}

	if err := filepath.Walk(dir, func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if stat.IsDir() {
			return nil
		}

		contents = append(contents, &files.Content{
			Type:        files.TypeFile,
			Source:      path,
			Destination: path[len(dir)+1:],
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
			},
		})

		return nil
	}); err != nil {
		return "", fmt.Errorf("mkdeb: can't walk output directory: %w", err)
	}

	contents, err = files.PrepareForPackager(contents, 0o002, "deb", true, internal.SourceEpoch())
	if err != nil {
		return "", fmt.Errorf("mkdeb: can't prepare for packager: %w", err)
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

	info.Overridables.RPM.Group = p.Group

	if *internal.GPGKeyID != "" {
		slog.Debug("using GPG key", "file", *internal.GPGKeyFile, "id", *internal.GPGKeyID)
		info.Overridables.Deb.Signature.KeyFile = *internal.GPGKeyFile
		info.Overridables.Deb.Signature.KeyID = internal.GPGKeyID
		info.Overridables.Deb.Signature.KeyPassphrase = *internal.GPGKeyPassword
	}

	pkg, err := nfpm.Get("deb")
	if err != nil {
		return "", fmt.Errorf("mkdeb: can't get RPM packager: %w", err)
	}

	foutpath = pkg.ConventionalFileName(info)
	fout, err := os.Create(filepath.Join(*internal.PackageDestDir, foutpath))
	if err != nil {
		return "", fmt.Errorf("mkdeb: can't create output file: %w", err)
	}
	defer fout.Close()

	if err := pkg.Package(info, fout); err != nil {
		return "", fmt.Errorf("mkdeb: can't build package: %w", err)
	}

	slog.Info("built package", "name", p.Name, "arch", p.Goarch, "version", p.Version, "path", fout.Name())

	return fout.Name(), err
}
