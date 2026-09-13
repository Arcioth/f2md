//go:build windows

package main

import (
	"os"
	"syscall"
	"time"
)

func fileBirthTime(path string, info os.FileInfo) time.Time {
	if st, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return time.Unix(0, st.CreationTime.Nanoseconds())
	}
	return info.ModTime()
}
