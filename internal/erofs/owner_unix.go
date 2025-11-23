//go:build !windows
// +build !windows

// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package erofs

import (
	"archive/tar"
	"io/fs"
	"syscall"
)

func getOwner(fi fs.FileInfo) (uid, gid int) {
	switch fi.Sys().(type) {
	case *syscall.Stat_t:
		stat := fi.Sys().(*syscall.Stat_t)

		uid = int(stat.Uid)
		gid = int(stat.Gid)

	case *tar.Header:
		hdr := fi.Sys().(*tar.Header)

		uid = hdr.Uid
		gid = hdr.Gid
	}

	return
}
