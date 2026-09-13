//go:build linux

package main

import (
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

func sysStatx() uintptr {
	switch runtime.GOARCH {
	case "amd64":
		return 332
	case "arm64", "riscv64", "loong64":
		return 291
	case "386", "ppc64", "ppc64le":
		return 383
	case "arm":
		return 397
	case "s390x":
		return 379
	default:
		return 0
	}
}

const (
	atFDCWD           = ^uintptr(99) // AT_FDCWD = -100
	atSymlinkNoFollow = 0x100
	statxBtime        = 0x800
)

type statxTimestamp struct {
	Sec  int64
	Nsec uint32
	Pad  int32
}

type statxT struct {
	Mask           uint32
	Blksize        uint32
	Attributes     uint64
	Nlink          uint32
	Uid            uint32
	Gid            uint32
	Mode           uint16
	_              uint16
	Ino            uint64
	Size           uint64
	Blocks         uint64
	AttributesMask uint64
	Atime          statxTimestamp
	Btime          statxTimestamp
	Ctime          statxTimestamp
	Mtime          statxTimestamp
	RdevMajor      uint32
	RdevMinor      uint32
	DevMajor       uint32
	DevMinor       uint32
	MntID          uint64
	DioMemAlign    uint32
	DioOffsetAlign uint32
	_              [12]uint64
}

func fileBirthTime(path string, info os.FileInfo) time.Time {
	if t, ok := statxBirth(path); ok {
		return t
	}
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return time.Unix(st.Ctim.Sec, st.Ctim.Nsec)
	}
	return info.ModTime()
}

func statxBirth(path string) (time.Time, bool) {
	p, err := syscall.BytePtrFromString(path)
	if err != nil {
		return time.Time{}, false
	}
	nr := sysStatx()
	if nr == 0 {
		return time.Time{}, false
	}
	var stx statxT
	_, _, errno := syscall.Syscall6(
		nr,
		atFDCWD,
		uintptr(unsafe.Pointer(p)),
		atSymlinkNoFollow,
		statxBtime,
		uintptr(unsafe.Pointer(&stx)),
		0,
	)
	if errno != 0 {
		return time.Time{}, false
	}
	if stx.Mask&statxBtime == 0 || (stx.Btime.Sec == 0 && stx.Btime.Nsec == 0) {
		return time.Time{}, false
	}
	return time.Unix(stx.Btime.Sec, int64(stx.Btime.Nsec)), true
}
