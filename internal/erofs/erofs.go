// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Portions of this file are based on code originally from: github.com/google/gvisor
 *
 * Copyright 2023 The gVisor Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package erofs

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	_ fs.FS        = (*Filesystem)(nil)
	_ fs.ReadDirFS = (*Filesystem)(nil)
	_ fs.StatFS    = (*Filesystem)(nil)
)

type Filesystem struct {
	image *Image
	root  *dirEntry
}

func Open(src io.ReaderAt) (*Filesystem, error) {
	image := &Image{src: src}

	if err := image.initSuperBlock(); err != nil {
		return nil, err
	}

	return &Filesystem{
		image: image,
		root: &dirEntry{
			image: image,
			nid:   image.RootNid(),
			typ:   FT_DIR,
		},
	}, nil
}

func (fsys *Filesystem) Open(name string) (fs.File, error) {
	de, err := fsys.resolve(name, false)
	if err != nil {
		return nil, err
	}

	return &file{
		image: fsys.image,
		de:    de,
	}, nil
}

func (fsys *Filesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	de, err := fsys.resolve(name, false)
	if err != nil {
		return nil, err
	}

	if !de.IsDir() {
		return nil, errors.New("not a directory")
	}

	ino, err := de.getInode()
	if err != nil {
		return nil, err
	}

	var dirents []fs.DirEntry
	err = ino.IterDirents(func(name string, typ uint8, nid uint64) error {
		// Skip "." and ".." entries.
		if name == "." || name == ".." {
			return nil
		}

		dirents = append(dirents, &dirEntry{
			image: de.image,
			name:  name,
			nid:   nid,
			typ:   typ,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return dirents, nil
}

func (fsys *Filesystem) Stat(name string) (fs.FileInfo, error) {
	de, err := fsys.resolve(name, false)
	if err != nil {
		return nil, err
	}

	ino, err := de.getInode()
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		image: de.image,
		name:  de.name,
		inode: ino,
	}, nil
}

// ReadLink returns the destination of the named symbolic link.
// Experimental implementation of: https://github.com/golang/go/issues/49580
func (fsys *Filesystem) ReadLink(name string) (string, error) {
	de, err := fsys.resolve(name, true)
	if err != nil {
		return "", err
	}

	ino, err := de.getInode()
	if err != nil {
		return "", err
	}

	return ino.Readlink()
}

// StatLink returns a FileInfo describing the file without following any symbolic links.
// Experimental implementation of: https://github.com/golang/go/issues/49580
func (fsys *Filesystem) StatLink(name string) (fs.FileInfo, error) {
	de, err := fsys.resolve(name, true)
	if err != nil {
		return nil, err
	}

	ino, err := de.getInode()
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		name:  de.name,
		inode: ino,
	}, nil
}

func (fsys *Filesystem) resolve(name string, noResolveLastSymlink bool) (*dirEntry, error) {
	de := fsys.root

	components := splitPath(name)
	for i, comp := range components {
		child, err := de.lookup(comp)
		if err != nil {
			return nil, err
		}

		ino, err := child.getInode()
		if err != nil {
			return nil, err
		}

		if ino.IsSymlink() && !(noResolveLastSymlink && i == len(components)-1) {
			link, err := ino.Readlink()
			if err != nil {
				return nil, err
			}
			link = filepath.Clean(link)

			if strings.HasPrefix(link, "/") {
				link = strings.TrimPrefix(link, "/")
			} else {
				link = filepath.Join(strings.Join(components[:i], "/"), link)
			}

			child, err = fsys.resolve(link, noResolveLastSymlink)
			if err != nil {
				return nil, err
			}
		}

		de = child
	}
	return de, nil
}

type file struct {
	image *Image
	de    *dirEntry
	r     io.Reader
}

func (f *file) Read(p []byte) (int, error) {
	if f.r == nil {
		ino, err := f.de.getInode()
		if err != nil {
			return 0, err
		}

		f.r, err = ino.Data()
		if err != nil {
			return 0, err
		}
	}

	return f.r.Read(p)
}

func (f *file) Close() error {
	return nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f.de.Info()
}

type dirEntry struct {
	image         *Image
	name          string
	typ           uint8
	nid           uint64
	readInodeOnce sync.Once
	inode         *Inode
}

func (de *dirEntry) Name() string {
	return de.name
}

func (de *dirEntry) IsDir() bool {
	return de.typ == FT_DIR
}

func (de *dirEntry) Type() fs.FileMode {
	ino, err := de.getInode()
	if err != nil {
		return 0
	}

	return ino.Mode()
}

func (de *dirEntry) Info() (fs.FileInfo, error) {
	ino, err := de.getInode()
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		image: de.image,
		name:  de.name,
		inode: ino,
	}, nil
}

func (de *dirEntry) lookup(name string) (*dirEntry, error) {
	ino, err := de.getInode()
	if err != nil {
		return nil, err
	}

	d, err := ino.Lookup(name)
	if err != nil {
		return nil, err
	}

	return &dirEntry{
		image: de.image,
		name:  name,
		nid:   d.Nid,
		typ:   d.FileType,
	}, nil
}

func (de *dirEntry) getInode() (Inode, error) {
	de.readInodeOnce.Do(func() {
		ino, err := de.image.Inode(de.nid)
		if err != nil {
			panic(err)
		}
		de.inode = &ino
	})

	return *de.inode, nil
}

type fileInfo struct {
	image *Image
	name  string
	inode Inode
}

func (fi *fileInfo) Name() string {
	return fi.name
}

func (fi *fileInfo) Size() int64 {
	return int64(fi.inode.Size())
}

func (fi *fileInfo) Mode() fs.FileMode {
	return fi.inode.Mode()
}

func (fi *fileInfo) ModTime() time.Time {
	return time.Unix(int64(fi.inode.Mtime()), 0)
}

func (fi *fileInfo) IsDir() bool {
	return fi.inode.IsDir()
}

func (fi *fileInfo) Sys() any {
	return &fi.inode
}

func splitPath(path string) []string {
	var components []string
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part != "" {
			components = append(components, part)
		}
	}
	return components
}
