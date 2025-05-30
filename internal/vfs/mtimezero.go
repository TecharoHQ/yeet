package vfs

import (
	"io/fs"
	"time"
)

// MtimeZeroFS wraps an fs.FS and overrides all file mtimes to time.Unix(0, 0).
type MtimeZeroFS struct {
	fs.FS
}

// Open overrides the FS.Open method to wrap returned files.
func (m MtimeZeroFS) Open(name string) (fs.File, error) {
	f, err := m.FS.Open(name)
	if err != nil {
		return nil, err
	}
	return &mtimeZeroFile{File: f}, nil
}

// ReadDir implements fs.ReadDirFS if the underlying FS supports it.
func (m MtimeZeroFS) ReadDir(name string) ([]fs.DirEntry, error) {
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
		wrapped[i] = mtimeZeroDirEntry{DirEntry: entry}
	}
	return wrapped, nil
}

// mtimeZeroFile wraps fs.File to override Stat().ModTime().
type mtimeZeroFile struct {
	fs.File
}

func (f *mtimeZeroFile) Stat() (fs.FileInfo, error) {
	info, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	return mtimeZeroFileInfo{FileInfo: info}, nil
}

// mtimeZeroFileInfo overrides ModTime to return time.Unix(0, 0).
type mtimeZeroFileInfo struct {
	fs.FileInfo
}

func (fi mtimeZeroFileInfo) ModTime() time.Time {
	return time.Unix(0, 0)
}

// mtimeZeroDirEntry wraps fs.DirEntry to override Info().ModTime().
type mtimeZeroDirEntry struct {
	fs.DirEntry
}

func (d mtimeZeroDirEntry) Info() (fs.FileInfo, error) {
	info, err := d.DirEntry.Info()
	if err != nil {
		return nil, err
	}
	return mtimeZeroFileInfo{FileInfo: info}, nil
}
