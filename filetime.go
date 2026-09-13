package main

import (
	"os"
	"time"
)

func fileTimes(path string, info os.FileInfo) (created, modified time.Time) {
	modified = info.ModTime()
	created = fileBirthTime(path, info)
	if created.IsZero() {
		created = modified
	}
	return created, modified
}
