// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package erofs_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/TecharoHQ/yeet/internal/erofs"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
)

func TestEROFS(t *testing.T) {
	f, err := os.Open("testdata/toybox.img")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	fsys, err := erofs.Open(f)
	require.NoError(t, err)

	t.Run("Open", func(t *testing.T) {
		t.Run("File", func(t *testing.T) {
			f, err := fsys.Open("/usr/bin/toybox")
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, f.Close())
			})

			info, err := f.Stat()
			require.NoError(t, err)

			require.Equal(t, "toybox", info.Name())
			require.Equal(t, 849544, int(info.Size()))
			require.Equal(t, os.FileMode(0o555), info.Mode()&fs.ModePerm)
			require.False(t, info.IsDir())

			h := sha256.New()

			n, err := io.Copy(h, f)
			require.NoError(t, err)

			require.Equal(t, int64(849544), n)

			require.Equal(t, "31aa01d6d46f63edcadc00fd5c40f3474f0df6c22a39ed0c5751ba3efa2855ac", hex.EncodeToString(h.Sum(nil)))
		})

		t.Run("Symlink", func(t *testing.T) {
			f, err := fsys.Open("bin/sh")
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, f.Close())
			})

			info, err := f.Stat()
			require.NoError(t, err)

			require.Equal(t, "toybox", info.Name())
			require.Equal(t, os.FileMode(0o555), info.Mode()&fs.ModePerm)
			require.False(t, info.IsDir())

			h := sha256.New()

			n, err := io.Copy(h, f)
			require.NoError(t, err)

			require.Equal(t, int64(849544), n)

			require.Equal(t, "31aa01d6d46f63edcadc00fd5c40f3474f0df6c22a39ed0c5751ba3efa2855ac", hex.EncodeToString(h.Sum(nil)))
		})
	})

	t.Run("ReadDir", func(t *testing.T) {
		entries, err := fsys.ReadDir("/etc")
		require.NoError(t, err)

		require.Len(t, entries, 5)

		require.Equal(t, "group", entries[0].Name())
		require.False(t, entries[0].IsDir())
		require.Equal(t, 0o644, int(entries[0].Type()))

		require.Equal(t, "os-release", entries[1].Name())
		require.False(t, entries[1].IsDir())
		require.Equal(t, 0o644, int(entries[1].Type()))

		require.Equal(t, "passwd", entries[2].Name())
		require.False(t, entries[2].IsDir())
		require.Equal(t, 0o644, int(entries[2].Type()))

		require.Equal(t, "rc", entries[3].Name())
		require.True(t, entries[3].IsDir())
		require.True(t, entries[3].Type()&fs.ModeDir > 0)

		require.Equal(t, "resolv.conf", entries[4].Name())
		require.False(t, entries[4].IsDir())
		require.Equal(t, 0o644, int(entries[4].Type()))
	})

	t.Run("Stat", func(t *testing.T) {
		t.Run("File", func(t *testing.T) {
			info, err := fsys.Stat("/usr/bin/toybox")
			require.NoError(t, err)

			require.Equal(t, "toybox", info.Name())
			require.Equal(t, 849544, int(info.Size()))
			require.Equal(t, os.FileMode(0o555), info.Mode()&fs.ModePerm)
			require.False(t, info.IsDir())

			ino, ok := info.Sys().(*erofs.Inode)
			require.True(t, ok)

			require.Equal(t, uint64(0x327), ino.Nid())
			require.Zero(t, ino.UID())
			require.Zero(t, ino.GID())
		})

		t.Run("Dir", func(t *testing.T) {
			info, err := fsys.Stat("usr")
			require.NoError(t, err)

			require.Equal(t, "usr", info.Name())
			require.Equal(t, os.FileMode(0o755), info.Mode()&fs.ModePerm)
			require.True(t, info.IsDir())

			ino, ok := info.Sys().(*erofs.Inode)
			require.True(t, ok)

			require.Equal(t, uint64(0x9c), ino.Nid())
			require.Zero(t, ino.UID())
			require.Zero(t, ino.GID())
		})

		t.Run("Symlink", func(t *testing.T) {
			info, err := fsys.Stat("bin/sh")
			require.NoError(t, err)

			require.Equal(t, "toybox", info.Name())
			require.Equal(t, os.FileMode(0o555), info.Mode()&fs.ModePerm)
			require.False(t, info.IsDir())

			ino, ok := info.Sys().(*erofs.Inode)
			require.True(t, ok)

			require.Equal(t, uint64(0x327), ino.Nid())
			require.Zero(t, ino.UID())
			require.Zero(t, ino.GID())
		})
	})

	t.Run("Readlink", func(t *testing.T) {
		target, err := fsys.ReadLink("bin")
		require.NoError(t, err)

		require.Equal(t, "usr/bin", target)
	})

	t.Run("StatLink", func(t *testing.T) {
		info, err := fsys.StatLink("bin")
		require.NoError(t, err)

		require.Equal(t, "bin", info.Name())
		require.Equal(t, os.FileMode(0o777), info.Mode()&fs.ModePerm)
		require.False(t, info.IsDir())

		ino, ok := info.Sys().(*erofs.Inode)
		require.True(t, ok)

		require.Equal(t, uint64(0x2f), ino.Nid())
		require.Zero(t, ino.UID())
		require.Zero(t, ino.GID())
	})

	t.Run("WalkDir", func(t *testing.T) {
		var paths []string
		err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			paths = append(paths, path)

			return nil
		})
		require.NoError(t, err)

		require.Len(t, paths, 266)

		require.Equal(t, []string{
			".",
			"bin",
			"dev",
			"etc",
			"etc/group",
		}, paths[:5])
	})

	t.Run("DirHash", func(t *testing.T) {
		var files []string
		err = fs.WalkDir(fsys, ".", func(file string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || d.Type()&fs.ModeSymlink != 0 {
				return nil
			}

			files = append(files, filepath.ToSlash(file))
			return nil
		})
		require.NoError(t, err)

		h, err := dirhash.Hash1(files, func(name string) (io.ReadCloser, error) {
			return fsys.Open(name)
		})
		require.NoError(t, err)

		require.Equal(t, "h1:adgxkqVceeKMyJdMZMvcUIbg94TthnXUmOeufCPuzQI=", h)
	})
}

func TestEROFSCreate(t *testing.T) {
	srcFile, err := os.Open("testdata/toybox.img")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, srcFile.Close())
	})

	srcFS, err := erofs.Open(srcFile)
	require.NoError(t, err)

	dstFile, err := os.OpenFile(filepath.Join(t.TempDir()+"/toybox.img"), os.O_RDWR|os.O_CREATE, 0o644)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dstFile.Close())
	})

	require.NoError(t, erofs.Create(dstFile, srcFS))

	dstFS, err := erofs.Open(dstFile)
	require.NoError(t, err)

	var files []string
	err = fs.WalkDir(dstFS, ".", func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		files = append(files, filepath.ToSlash(file))
		return nil
	})
	require.NoError(t, err)

	h, err := dirhash.Hash1(files, func(name string) (io.ReadCloser, error) {
		return srcFS.Open(name)
	})
	require.NoError(t, err)

	require.Equal(t, "h1:adgxkqVceeKMyJdMZMvcUIbg94TthnXUmOeufCPuzQI=", h)
}
