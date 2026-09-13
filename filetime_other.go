//go:build !linux && !darwin && !windows

package main

import (
	"os"
	"time"
)

func fileBirthTime(path string, info os.FileInfo) time.Time {
	return info.ModTime()
}
