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

// Package erofs provides the ability to access the contents in an EROFS [1] image.
//
// The design principle of this package is that, it will just provide the ability
// to access the contents in the image, and it will never cache any objects internally.
//
// [1] https://docs.kernel.org/filesystems/erofs.html
package erofs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/fs"
)

const (
	// Definitions for superblock.
	SuperBlockMagicV1 = 0xe0f5e1e2
	SuperBlockOffset  = 1024

	// Inode slot size in bit shift.
	InodeSlotBits = 5

	// Max file name length.
	MaxNameLen = 255
)

// Bit definitions for Inode*::Format.
const (
	InodeLayoutBit  = 0
	InodeLayoutBits = 1

	InodeDataLayoutBit  = 1
	InodeDataLayoutBits = 3
)

// Inode layouts.
const (
	InodeLayoutCompact  = 0
	InodeLayoutExtended = 1
)

// Inode data layouts.
const (
	InodeDataLayoutFlatPlain = iota
	InodeDataLayoutFlatCompressionLegacy
	InodeDataLayoutFlatInline
	InodeDataLayoutFlatCompression
	InodeDataLayoutChunkBased
	InodeDataLayoutMax
)

// Features w/ backward compatibility.
// This is not exhaustive, unused features are not listed.
const (
	FeatureCompatSuperBlockChecksum = 0x00000001
)

// Features w/o backward compatibility.
//
// Any features that aren't in FeatureIncompatSupported are incompatible
// with this implementation.
//
// This is not exhaustive, unused features are not listed.
const (
	FeatureIncompatSupported = 0x0
)

// SuperBlock represents on-disk superblock.
type SuperBlock struct {
	Magic           uint32    // Filesystem magic number
	Checksum        uint32    // CRC32C checksum of the superblock
	FeatureCompat   uint32    // Compatible feature flags
	BlockSizeBits   uint8     // Filesystem block size in bit shift
	ExtSlots        uint8     // Superblock extension slots
	RootNid         uint16    // Root directory inode number
	Inodes          uint64    // Total valid inodes
	BuildTime       uint64    // Build time of the filesystem
	BuildTimeNsec   uint32    // Nanoseconds part of build time
	Blocks          uint32    // Total number of blocks
	MetaBlockAddr   uint32    // Start block address of metadata area
	XattrBlockAddr  uint32    // Start block address of shared xattr area
	UUID            [16]uint8 // UUID for volume
	VolumeName      [16]uint8 // Volume name
	FeatureIncompat uint32    // Incompatible feature flags
	Union1          uint16    // Union for additional features
	ExtraDevices    uint16    // Number of extra devices
	DevTableSlotOff uint16    // Device table slot offset
	Reserved        [38]uint8 // Reserved for future use
}

// BlockSize returns the block size.
func (sb *SuperBlock) BlockSize() uint32 {
	return 1 << sb.BlockSizeBits
}

// BlockAddrToOffset converts block addr to the offset in image file.
func (sb *SuperBlock) BlockAddrToOffset(addr uint32) int64 {
	return int64(addr) << sb.BlockSizeBits
}

// MetaOffset returns the offset of metadata area in image file.
func (sb *SuperBlock) MetaOffset() int64 {
	return sb.BlockAddrToOffset(sb.MetaBlockAddr)
}

// NidToOffset converts inode number to the offset in image file.
func (sb *SuperBlock) NidToOffset(nid uint64) int64 {
	return int64(sb.MetaOffset()) + (int64(nid) << InodeSlotBits)
}

// InodeCompact represents 32-byte reduced form of on-disk inode.
type InodeCompact struct {
	Format       uint16 // Inode format hints
	XattrCount   uint16 // Xattr entry count
	Mode         uint16 // File mode
	Nlink        uint16 // Number of hard links
	Size         uint32 // File size in bytes
	Reserved     uint32 // Reserved for future use
	RawBlockAddr uint32 // Raw block address
	Ino          uint32 // Inode number
	UID          uint16 // User ID of owner
	GID          uint16 // Group ID of owner
	Reserved2    uint32 // Reserved for future use
}

// InodeExtended represents 64-byte complete form of on-disk inode.
type InodeExtended struct {
	Format       uint16    // Inode format hints
	XattrCount   uint16    // Xattr entry count
	Mode         uint16    // File mode
	Reserved     uint16    // Reserved for future use
	Size         uint64    // File size in bytes
	RawBlockAddr uint32    // Raw block address
	Ino          uint32    // Inode number
	UID          uint32    // User ID of owner
	GID          uint32    // Group ID of owner
	Mtime        uint64    // Last modification time
	MtimeNsec    uint32    // Nanoseconds part of Mtime
	Nlink        uint32    // Number of hard links
	Reserved2    [16]uint8 // Reserved for future use
}

// Dirent represents on-disk directory entry.
type Dirent struct {
	Nid      uint64 // Inode number
	NameOff  uint16 // Offset of file name
	FileType uint8  // File type
	Reserved uint8  // Reserved for future use
}

var DirentSize = int64(binary.Size(Dirent{}))

// Image represents an open EROFS image.
type Image struct {
	src io.ReaderAt
	sb  SuperBlock
}

// OpenImage returns an Image providing access to the contents in the image file src.
//
// On success, the ownership of src is transferred to Image.
func OpenImage(src io.ReaderAt) (*Image, error) {
	i := &Image{src: src}

	if err := i.initSuperBlock(); err != nil {
		return nil, err
	}

	return i, nil
}

// SuperBlock returns a copy of the image's superblock.
func (i *Image) SuperBlock() SuperBlock {
	return i.sb
}

// BlockSize returns the block size of this image.
func (i *Image) BlockSize() uint32 {
	return i.sb.BlockSize()
}

// Blocks returns the total blocks of this image.
func (i *Image) Blocks() uint32 {
	return i.sb.Blocks
}

// RootNid returns the root inode number of this image.
func (i *Image) RootNid() uint64 {
	return uint64(i.sb.RootNid)
}

// initSuperBlock initializes the superblock of this image.
func (i *Image) initSuperBlock() error {
	if err := i.unmarshalFrom(SuperBlockOffset, &i.sb); err != nil {
		return err
	}

	if i.sb.Magic != SuperBlockMagicV1 {
		return fmt.Errorf("unknown magic: 0x%x", i.sb.Magic)
	}

	if err := i.verifyChecksum(); err != nil {
		return err
	}

	if featureIncompat := i.sb.FeatureIncompat & ^uint32(FeatureIncompatSupported); featureIncompat != 0 {
		return fmt.Errorf("unsupported incompatible features detected: 0x%x", featureIncompat)
	}

	return nil
}

// verifyChecksum verifies the checksum of the superblock.
func (i *Image) verifyChecksum() error {
	if i.sb.FeatureCompat&FeatureCompatSuperBlockChecksum == 0 {
		return nil
	}

	sb := i.sb
	sb.Checksum = 0

	var marshalledSb bytes.Buffer
	if err := binary.Write(&marshalledSb, binary.LittleEndian, sb); err != nil {
		return err
	}

	table := crc32.MakeTable(crc32.Castagnoli)
	checksum := crc32.Checksum(marshalledSb.Bytes(), table)

	off := SuperBlockOffset + int64(binary.Size(i.sb))
	if buf, err := i.bytesAt(off, int64(i.BlockSize())-off); err != nil {
		return errors.New("image size is too small")
	} else {
		checksum = ^crc32.Update(checksum, table, buf)
	}
	if checksum != i.sb.Checksum {
		return fmt.Errorf("invalid checksum: 0x%x, expected: 0x%x", checksum, i.sb.Checksum)
	}

	return nil
}

// checksum populates the checksum of the superblock, it assumes the remainder
// of the block contains all zeroes.
func (sb *SuperBlock) checksum() error {
	sbCopy := *sb
	sbCopy.Checksum = 0

	// Marshal the superblock into a byte buffer.
	var marshalled bytes.Buffer
	if err := binary.Write(&marshalled, binary.LittleEndian, sbCopy); err != nil {
		return err
	}

	// Create a CRC32C table and calculate the checksum over the marshalled superblock.
	table := crc32.MakeTable(crc32.Castagnoli)
	checksum := crc32.Checksum(marshalled.Bytes(), table)

	// Calculate the remaining bytes in the block after the superblock
	// (assume all zeroes for now).
	off := SuperBlockOffset + int64(binary.Size(sb))
	remainingBytes := make([]byte, BlockSize-off)
	checksum = ^crc32.Update(checksum, table, remainingBytes)

	sb.Checksum = checksum

	return nil
}

// inodeFormatAt returns the format of the inode at offset off within the
// image.
func (i *Image) inodeFormatAt(off int64) (uint16, error) {
	if !checkInodeAlignment(off) {
		return 0, fmt.Errorf("invalid inode alignment at offset %d", off)
	}
	buf, err := i.bytesAt(off, 2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf), nil
}

// inodeCompactAt returns a pointer to the compact inode at offset off within
// the image.
func (i *Image) inodeCompactAt(off int64) (*InodeCompact, error) {
	if !checkInodeAlignment(off) {
		return nil, fmt.Errorf("invalid inode alignment at offset %d", off)
	}
	var inode InodeCompact
	if err := i.unmarshalFrom(int64(off), &inode); err != nil {
		return nil, err
	}

	return &inode, nil
}

// inodeExtendedAt returns a pointer to the extended inode at offset off within
// the image.
func (i *Image) inodeExtendedAt(off int64) (*InodeExtended, error) {
	if !checkInodeAlignment(off) {
		return nil, fmt.Errorf("invalid inode alignment at offset %d", off)
	}

	var inode InodeExtended
	if err := i.unmarshalFrom(int64(off), &inode); err != nil {
		return nil, err
	}
	return &inode, nil
}

// direntAt returns a pointer to the dirent at offset off within the image.
func (i *Image) direntAt(off int64) (*Dirent, error) {
	// Each valid dirent should be aligned to 4 bytes.
	if off&3 != 0 {
		return nil, fmt.Errorf("invalid dirent alignment at offset %d", off)
	}

	var dirent Dirent
	if err := i.unmarshalFrom(off, &dirent); err != nil {
		return nil, err
	}

	return &dirent, nil
}

// Inode returns the inode identified by nid.
func (i *Image) Inode(nid uint64) (Inode, error) {
	inode := Inode{
		image: i,
		nid:   nid,
	}

	off := i.sb.NidToOffset(nid)
	if format, err := i.inodeFormatAt(off); err != nil {
		return Inode{}, err
	} else {
		inode.format = format
	}

	var (
		rawBlockAddr uint32
		inodeSize    int64
	)

	switch layout := inode.Layout(); layout {
	case InodeLayoutCompact:
		ino, err := i.inodeCompactAt(off)
		if err != nil {
			return Inode{}, err
		}

		if ino.XattrCount != 0 {
			return Inode{}, fmt.Errorf("unsupported xattr at inode %d", nid)
		}

		rawBlockAddr = ino.RawBlockAddr
		inodeSize = int64(binary.Size(*ino))

		inode.size = uint64(ino.Size)
		inode.nlink = uint32(ino.Nlink)
		inode.mode = ino.Mode
		inode.uid = uint32(ino.UID)
		inode.gid = uint32(ino.GID)
		inode.mtime = i.sb.BuildTime
		inode.mtimeNsec = i.sb.BuildTimeNsec

	case InodeLayoutExtended:
		ino, err := i.inodeExtendedAt(off)
		if err != nil {
			return Inode{}, err
		}

		if ino.XattrCount != 0 {
			return Inode{}, fmt.Errorf("unsupported xattr at inode %d", nid)
		}

		rawBlockAddr = ino.RawBlockAddr
		inodeSize = int64(binary.Size(*ino))

		inode.size = ino.Size
		inode.nlink = ino.Nlink
		inode.mode = ino.Mode
		inode.uid = ino.UID
		inode.gid = ino.GID
		inode.mtime = ino.Mtime
		inode.mtimeNsec = ino.MtimeNsec

	default:
		return Inode{}, fmt.Errorf("unsupported layout at inode %d", nid)
	}

	blockSize := int64(i.BlockSize())
	inode.blocks = (int64(inode.size) + (blockSize - 1)) / blockSize

	switch dataLayout := inode.DataLayout(); dataLayout {
	case InodeDataLayoutFlatInline:
		// Check that whether the file data in the last block fits into
		// the remaining room of the metadata block.
		tailSize := int64(inode.size) & (blockSize - 1)
		if tailSize == 0 || tailSize > blockSize-inodeSize {
			return Inode{}, fmt.Errorf("inline data not found or cross block boundary at inode %d, tail size: %d",
				nid, tailSize)
		}
		inode.idataOff = off + inodeSize
		fallthrough

	case InodeDataLayoutFlatPlain:
		inode.dataOff = i.sb.BlockAddrToOffset(rawBlockAddr)

	default:
		return Inode{}, fmt.Errorf("unsupported data layout at inode %d", nid)
	}

	return inode, nil
}

// bytesAt returns the bytes at [off, off+n) of the image.
func (i *Image) bytesAt(off, n int64) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := i.src.ReadAt(buf, off); err != nil {
		return nil, err
	}
	return buf, nil
}

func (i *Image) unmarshalFrom(off int64, data any) error {
	if err := binary.Read(io.NewSectionReader(i.src, off, int64(binary.Size(data))),
		binary.LittleEndian, data); err != nil {
		return err
	}

	return nil
}

// checkInodeAlignment checks whether off matches inode's alignment requirement.
func checkInodeAlignment(off int64) bool {
	// Each valid inode should be aligned with an inode slot, which is
	// a fixed value (32 bytes).
	return off&((1<<InodeSlotBits)-1) == 0
}

// Inode represents an inode object.
type Inode struct {
	// image is the underlying image. Inode should not perform writable
	// operations (e.g. Close()) on the image.
	image *Image

	// dataOff points to the data of this inode in the data blocks.
	dataOff int64

	// idataOff points to the tail packing inline data of this inode
	// if it's not zero in the metadata block.
	idataOff int64

	// blocks indicates the count of blocks that store the data associated
	// with this inode. It will count in the metadata block that includes
	// the inline data as well.
	blocks int64

	// format is the format of this inode.
	format uint16

	// Metadata.
	mode      uint16
	nid       uint64
	size      uint64
	mtime     uint64
	mtimeNsec uint32
	uid       uint32
	gid       uint32
	nlink     uint32
}

// bitRange returns the bits within the range [bit, bit+bits) in value.
func bitRange(value, bit, bits uint16) uint16 {
	return (value >> bit) & ((1 << bits) - 1)
}

// Layout returns the inode layout.
func (ino *Inode) Layout() uint16 {
	return bitRange(ino.format, InodeLayoutBit, InodeLayoutBits)
}

// DataLayout returns the inode data layout.
func (ino *Inode) DataLayout() uint16 {
	return bitRange(ino.format, InodeDataLayoutBit, InodeDataLayoutBits)
}

// IsRegular indicates whether i represents a regular file.
func (ino *Inode) IsRegular() bool {
	return ino.mode&S_IFMT == S_IFREG
}

// IsDir indicates whether i represents a directory.
func (ino *Inode) IsDir() bool {
	return ino.mode&S_IFMT == S_IFDIR
}

// IsCharDev indicates whether i represents a character device.
func (ino *Inode) IsCharDev() bool {
	return ino.mode&S_IFMT == S_IFCHR
}

// IsBlockDev indicates whether i represents a block device.
func (ino *Inode) IsBlockDev() bool {
	return ino.mode&S_IFMT == S_IFBLK
}

// IsFIFO indicates whether i represents a named pipe.
func (ino *Inode) IsFIFO() bool {
	return ino.mode&S_IFMT == S_IFIFO
}

// IsSocket indicates whether i represents a socket.
func (ino *Inode) IsSocket() bool {
	return ino.mode&S_IFMT == S_IFSOCK
}

// IsSymlink indicates whether i represents a symbolic link.
func (ino *Inode) IsSymlink() bool {
	return ino.mode&S_IFMT == S_IFLNK
}

// Nid returns the inode number.
func (ino *Inode) Nid() uint64 {
	return ino.nid
}

// Size returns the data size.
func (ino *Inode) Size() uint64 {
	return ino.size
}

// Nlink returns the number of hard links.
func (ino *Inode) Nlink() uint32 {
	return ino.nlink
}

// Mtime returns the time of last modification.
func (ino *Inode) Mtime() uint64 {
	return ino.mtime
}

// MtimeNsec returns the nano second part of Mtime.
func (ino *Inode) MtimeNsec() uint32 {
	return ino.mtimeNsec
}

// Mode returns the file type and permissions.
func (ino *Inode) Mode() fs.FileMode {
	mode := fs.FileMode(ino.mode) & fs.ModePerm

	if ino.IsDir() {
		mode |= fs.ModeDir
	}
	if ino.IsCharDev() {
		mode |= fs.ModeCharDevice
	}
	if ino.IsBlockDev() {
		mode |= fs.ModeDevice
	}
	if ino.IsFIFO() {
		mode |= fs.ModeNamedPipe
	}
	if ino.IsSocket() {
		mode |= fs.ModeSocket
	}
	if ino.IsSymlink() {
		mode |= fs.ModeSymlink
	}

	return mode
}

// UID returns the user ID of the owner.
func (ino *Inode) UID() uint32 {
	return ino.uid
}

// GID returns the group ID of the owner.
func (ino *Inode) GID() uint32 {
	return ino.gid
}

// Data returns the read-only file data of this inode.
func (ino *Inode) Data() (io.Reader, error) {
	switch dataLayout := ino.DataLayout(); dataLayout {
	case InodeDataLayoutFlatPlain:
		return io.NewSectionReader(ino.image.src, int64(ino.dataOff), int64(ino.size)), nil

	case InodeDataLayoutFlatInline:
		readers := make([]io.Reader, 0, 2)
		idataSize := ino.size & (uint64(ino.image.BlockSize()) - 1)
		if ino.size > idataSize {
			readers = append(readers, io.NewSectionReader(ino.image.src, int64(ino.dataOff), int64(ino.size-idataSize)))
		}
		readers = append(readers, io.NewSectionReader(ino.image.src, int64(ino.idataOff), int64(idataSize)))
		return io.MultiReader(readers...), nil

	default:
		return nil, errors.New("unsupported data layout")
	}
}

// blockData represents the information of the data in a block.
type blockData struct {
	// base indicates the data offset within the image.
	base int64
	// size indicates the data size.
	size uint32
}

// getBlockDataInfo returns the information of the data in the block identified by
// blockIdx of this inode.
//
// Precondition: blockIdx < i.blocks.
func (ino *Inode) getBlockDataInfo(blockIdx uint64) blockData {
	blockSize := ino.image.BlockSize()
	lastBlock := int64(blockIdx) == ino.blocks-1
	base := ino.idataOff
	if !lastBlock || base == 0 {
		base = ino.dataOff + int64(blockIdx)*int64(blockSize)
	}
	size := blockSize
	if lastBlock {
		if tailSize := uint32(ino.size) & (blockSize - 1); tailSize != 0 {
			size = tailSize
		}
	}
	return blockData{base, size}
}

// getDirentName returns the name of dirent d in the given block of this inode.
//
// The on-disk format of one block looks like this:
//
//	                 ___________________________
//	                /                           |
//	               /              ______________|________________
//	              /              /              | nameoff1       | nameoffN-1
//	 ____________.______________._______________v________________v__________
//	| dirent | dirent | ... | dirent | filename | filename | ... | filename |
//	|___.0___|____1___|_____|___N-1__|____0_____|____1_____|_____|___N-1____|
//	     \                           ^
//	      \                          |                           * could have
//	       \                         |                             trailing '\0'
//	        \________________________| nameoff0
//	                            Directory block
//
// The on-disk format of one directory looks like this:
//
// [ (block 1) dirent 1 | dirent 2 | dirent 3 | name 1 | name 2 | name 3 | optional padding ]
// [ (block 2) dirent 4 | dirent 5 | name 4 | name 5 | optional padding ]
// ...
// [ (block N) dirent M | dirent M+1 | name M | name M+1 | optional padding ]
//
// [ (metadata block) inode | optional fields | dirent M+2 | dirent M+3 | name M+2 | name M+3 | optional padding ]
//
// Refer: https://docs.kernel.org/filesystems/erofs.html#directories
func (ino *Inode) getDirentName(d *Dirent, direntOff int64, block blockData, lastDirent bool) ([]byte, error) {
	var nameLen uint32
	if lastDirent {
		nameLen = block.size - uint32(d.NameOff)
	} else {
		next, err := ino.image.direntAt(direntOff + DirentSize)
		if err != nil {
			return nil, err
		}

		nameLen = uint32(next.NameOff - d.NameOff)
	}
	if uint32(d.NameOff)+nameLen > block.size || nameLen > MaxNameLen || nameLen == 0 {
		return nil, errors.New("corrupted dirent")
	}
	name, err := ino.image.bytesAt(int64(block.base)+int64(d.NameOff), int64(nameLen))
	if err != nil {
		return nil, err
	}
	if lastDirent {
		// Optional padding may exist at the end of a block.
		n := bytes.IndexByte(name, 0)
		if n == 0 {
			return nil, errors.New("corrupted dirent")
		}
		if n != -1 {
			name = name[:n]
		}
	}
	return name, nil
}

// getDirent0 returns a pointer to the first dirent in the given block of this inode.
func (ino *Inode) getDirent0(block blockData) (*Dirent, error) {
	d0, err := ino.image.direntAt(int64(block.base))
	if err != nil {
		return nil, err
	}
	if d0.NameOff < uint16(DirentSize) || uint32(d0.NameOff) >= block.size {
		return nil, fmt.Errorf("invalid nameOff0 %d at inode %d", d0.NameOff, ino.Nid())
	}
	return d0, nil
}

// Lookup looks up a child by the name.
func (ino *Inode) Lookup(name string) (Dirent, error) {
	if !ino.IsDir() {
		return Dirent{}, fs.ErrInvalid
	}

	// Currently (Go 1.21), there is no safe and efficient way to do three-way
	// string comparisons, so let's convert the string to a byte slice first.
	nameBytes := []byte(name)

	// In EROFS, all directory entries are _strictly_ recorded in alphabetical
	// order. The lookup is done by directly performing binary search on the
	// disk data similar to what Linux does in fs/erofs/namei.c:erofs_namei().
	var (
		targetBlock      blockData
		targetNumDirents uint16
	)

	// Find the block that may contain the target dirent first.
	bLeft, bRight := int64(0), int64(ino.blocks)-1
	for bLeft <= bRight {
		// Cast to uint64 to avoid overflow.
		mid := uint64(bLeft+bRight) >> 1
		block := ino.getBlockDataInfo(mid)
		d0, err := ino.getDirent0(block)
		if err != nil {
			return Dirent{}, err
		}
		numDirents := d0.NameOff / uint16(DirentSize)
		d0Name, err := ino.getDirentName(d0, int64(block.base), block, numDirents == 1)
		if err != nil {
			return Dirent{}, err
		}
		switch bytes.Compare(nameBytes, d0Name) {
		case 0:
			// Found the target dirent.
			return *d0, nil
		case 1:
			// name > d0Name, this block may contain the target dirent.
			targetBlock = block
			targetNumDirents = numDirents
			bLeft = int64(mid) + 1
		case -1:
			// name < d0Name, this is not the block we're looking for.
			bRight = int64(mid) - 1
		}
	}

	if targetBlock.base == 0 {
		// The target block was not found.
		return Dirent{}, fs.ErrNotExist
	}

	// Find the target dirent in the target block. Note that, as the 0th dirent
	// has already been checked during the block binary search, we don't need to
	// check it again and can define dLeft/dRight as unsigned types.
	dLeft, dRight := uint16(1), targetNumDirents-1
	for dLeft <= dRight {
		// The sum will never lead to a uint16 overflow, as the maximum value of
		// the operands is MaxUint16/DirentSize.
		mid := (dLeft + dRight) >> 1
		direntOff := int64(targetBlock.base) + int64(mid)*DirentSize
		d, err := ino.image.direntAt(direntOff)
		if err != nil {
			return Dirent{}, err
		}
		dName, err := ino.getDirentName(d, direntOff, targetBlock, mid == targetNumDirents-1)
		if err != nil {
			return Dirent{}, err
		}
		switch bytes.Compare(nameBytes, dName) {
		case 0:
			// Found the target dirent.
			return *d, nil
		case 1:
			// name > dName.
			dLeft = mid + 1
		case -1:
			// name < dName.
			dRight = mid - 1
		}
	}

	return Dirent{}, fs.ErrNotExist
}

// IterDirents invokes cb on each entry in the directory represented by this inode.
// The directory entries will be iterated in alphabetical order.
func (ino *Inode) IterDirents(cb func(name string, typ uint8, nid uint64) error) error {
	if !ino.IsDir() {
		return fs.ErrInvalid
	}

	// Iterate all the blocks which contain dirents.
	for blockIdx := uint64(0); blockIdx < uint64(ino.blocks); blockIdx++ {
		block := ino.getBlockDataInfo(blockIdx)
		d, err := ino.getDirent0(block)
		if err != nil {
			return err
		}
		// Iterate all the dirents in this block.
		numDirents := d.NameOff / uint16(DirentSize)
		direntOff := int64(block.base)
		for {
			name, err := ino.getDirentName(d, direntOff, block, numDirents == 1)
			if err != nil {
				return err
			}
			if err := cb(string(name), d.FileType, d.Nid); err != nil {
				return err
			}
			if numDirents--; numDirents == 0 {
				break
			}

			direntOff += DirentSize
			d, err = ino.image.direntAt(direntOff)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Readlink reads the link target.
func (ino *Inode) Readlink() (string, error) {
	if !ino.IsSymlink() {
		return "", fs.ErrInvalid
	}
	off := int64(ino.dataOff)
	size := int64(ino.size)
	if ino.idataOff != 0 {
		// Inline symlink data shouldn't cross block boundary.
		if ino.blocks > 1 {
			return "", fmt.Errorf("inline data cross block boundary at inode %d", ino.Nid())
		}
		off = int64(ino.idataOff)
	} else {
		// This matches Linux's behaviour in fs/namei.c:page_get_link() and
		// include/linux/namei.h:nd_terminate_link().
		if size > int64(ino.image.BlockSize())-1 {
			size = int64(ino.image.BlockSize()) - 1
		}
	}
	target, err := ino.image.bytesAt(off, size)
	if err != nil {
		return "", err
	}
	return string(target), nil
}
