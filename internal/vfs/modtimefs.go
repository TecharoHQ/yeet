package vfs

import (
	"io/fs"
	"time"
)

// ModTimeFS wraps an fs.FS and overrides all file mtimes with a fixed time.
type ModTimeFS struct {
	fs.FS
	Time time.Time
}

// Open overrides the FS.Open method to wrap returned files.
func (m ModTimeFS) Open(name string) (fs.File, error) {
	f, err := m.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return &modTimeFile{File: f, Time: m.Time}, nil
}

// ReadDir implements fs.ReadDirFS if the underlying FS supports it.
func (m ModTimeFS) ReadDir(name string) ([]fs.DirEntry, error) {
	readDirFS, ok := m.FS.(fs.ReadDirFS)
	if !ok {
		return nil, &fs.PathError{Op: "ReadDir", Path: name, Err: fs.ErrInvalid}
	}

	entries, err := readDirFS.ReadDir(name)
	if err != nil {
		return nil, err
	}

	wrapped := make([]fs.DirEntry, len(entries))
	for i, entry := range entries {
		wrapped[i] = modTimeDirEntry{DirEntry: entry, Time: m.Time}
	}
	return wrapped, nil
}

// modTimeFile wraps fs.File to override Stat().ModTime().
type modTimeFile struct {
	fs.File
	Time time.Time
}

func (f *modTimeFile) Stat() (fs.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	return modTimeFileInfo{FileInfo: info, Time: f.Time}, nil
}

// modTimeFileInfo overrides ModTime to return a fixed time.
type modTimeFileInfo struct {
	fs.FileInfo
	Time time.Time
}

func (fi modTimeFileInfo) ModTime() time.Time {
	return fi.Time
}

// modTimeDirEntry wraps fs.DirEntry to override Info().ModTime().
type modTimeDirEntry struct {
	fs.DirEntry
	Time time.Time
}

func (d modTimeDirEntry) Info() (fs.FileInfo, error) {
	info, err := d.DirEntry.Info()
	if err != nil {
		return nil, err
	}
	return modTimeFileInfo{FileInfo: info, Time: d.Time}, nil
}
