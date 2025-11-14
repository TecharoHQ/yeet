package pkgmeta

import (
	"os"
	"path/filepath"

	"github.com/TecharoHQ/yeet/internal"
	"github.com/goreleaser/nfpm/v2/files"
)

type Package struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Homepage    string   `json:"homepage"`
	Group       string   `json:"group"`
	License     string   `json:"license"`
	Platform    string   `json:"platform"` // if not set, default to linux
	Goarch      string   `json:"goarch"`
	Replaces    []string `json:"replaces"`
	Depends     []string `json:"depends"`
	Recommends  []string `json:"recommends"`

	EmptyDirs     []string          `json:"emptyDirs"`     // rpm destination path
	ConfigFiles   map[string]string `json:"configFiles"`   // pwd-relative source path, rpm destination path
	Documentation map[string]string `json:"documentation"` // pwd-relative source path, file in /usr/share/doc/$Name
	Files         map[string]string `json:"files"`         // pwd-relative source path, rpm destination path

	Build    func(BuildInput)     `json:"build"`
	Filename func(Package) string `json:"mkFilename"`
}

type InstalledFileCallback func(files.Content) (*files.Content, error)

func (p Package) CopyConfigFiles() files.Contents {
	contents := make([]*files.Content, 0, len(p.ConfigFiles))

	for repoPath, pkgPath := range p.ConfigFiles {
		defaultContent := files.Content{
			Type:        files.TypeConfig,
			Source:      repoPath,
			Destination: pkgPath,
			FileInfo: &files.ContentFileInfo{
				Mode:  os.FileMode(0600),
				MTime: internal.SourceEpoch(),
			},
		}

		contents = append(contents, &defaultContent)
	}

	return contents
}

func (p Package) CopyDocumentation() files.Contents {
	contents := make([]*files.Content, 0, len(p.Documentation))

	for repoPath, pkgPath := range p.Documentation {
		defaultContent := files.Content{
			Type:        files.TypeFile,
			Source:      repoPath,
			Destination: filepath.Join("/usr/share/doc", p.Name, pkgPath),
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
			},
		}

		contents = append(contents, &defaultContent)
	}

	return contents
}

func (p Package) CopyEmptyDirs(mode os.FileMode) files.Contents {
	contents := make([]*files.Content, 0, len(p.Documentation))

	for _, d := range p.EmptyDirs {
		if d == "" {
			continue
		}

		contents = append(contents, &files.Content{
			Type:        files.TypeDir,
			Destination: d,
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
				Mode:  mode,
			},
		})
	}

	return contents
}

func (p Package) CopyFiles() files.Contents {
	contents := make([]*files.Content, 0, len(p.ConfigFiles))

	for repoPath, pkgPath := range p.Files {
		defaultContent := files.Content{
			Type:        files.TypeFile,
			Source:      repoPath,
			Destination: pkgPath,
			FileInfo: &files.ContentFileInfo{
				MTime: internal.SourceEpoch(),
			},
		}

		contents = append(contents, &defaultContent)
	}

	return contents
}

func (p Package) CopyTree(dir string) (files.Contents, error) {
	contents := make([]*files.Content, 0, len(p.ConfigFiles))

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
		return nil, err
	}

	return contents, nil
}

type BuildInput struct {
	Output string `json:"out"`
	Bin    string `json:"bin"`
	Doc    string `json:"doc"`
	Etc    string `json:"etc"`
	Man    string `json:"man"`

	Systemd string  `json:"systemd,omitempty"`
	Openrc  *Openrc `json:"openrc,omitempty"`
}

type Openrc struct {
	InitDir string `json:"init"`
	ConfDir string `json:"conf"`
}

func (b BuildInput) String() string {
	return b.Output
}
