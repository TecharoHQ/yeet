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
	"io/fs"
)

// Linux values for fs on-disk file types.
const (
	FT_REG_FILE = 1
	FT_DIR      = 2
	FT_CHRDEV   = 3
	FT_BLKDEV   = 4
	FT_FIFO     = 5
	FT_SOCK     = 6
	FT_SYMLINK  = 7
)

func fileTypeFromFileMode(mode fs.FileMode) uint8 {
	switch mode.Type() {
	case fs.ModeDir:
		return FT_DIR
	case fs.ModeSymlink:
		return FT_SYMLINK
	case fs.ModeDevice:
		return FT_BLKDEV
	case fs.ModeCharDevice:
		return FT_CHRDEV
	case fs.ModeNamedPipe:
		return FT_FIFO
	case fs.ModeSocket:
		return FT_SOCK
	default:
		return FT_REG_FILE
	}
}

// Values for mode_t.
const (
	S_IFMT   = 0170000
	S_IFSOCK = 0140000
	S_IFLNK  = 0120000
	S_IFREG  = 0100000
	S_IFBLK  = 060000
	S_IFDIR  = 040000
	S_IFCHR  = 020000
	S_IFIFO  = 010000
	S_ISUID  = 04000
	S_ISGID  = 02000
	S_ISVTX  = 01000
)

func statModeFromFileMode(mode fs.FileMode) uint16 {
	stMode := uint16(mode.Perm())

	switch mode & fs.ModeType {
	case fs.ModeDir:
		stMode |= S_IFDIR
	case fs.ModeSymlink:
		stMode |= S_IFLNK
	case fs.ModeDevice:
		stMode |= S_IFBLK
	case fs.ModeCharDevice:
		stMode |= S_IFCHR
	case fs.ModeNamedPipe:
		stMode |= S_IFIFO
	case fs.ModeSocket:
		stMode |= S_IFSOCK
	default:
		stMode |= S_IFREG
	}

	// Handle setuid, setgid and sticky bits.
	if mode&fs.ModeSetuid != 0 {
		stMode |= S_ISUID
	}
	if mode&fs.ModeSetgid != 0 {
		stMode |= S_ISGID
	}
	if mode&fs.ModeSticky != 0 {
		stMode |= S_ISVTX
	}

	return stMode
}
