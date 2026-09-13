package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type sortMode int

const (
	sortByName sortMode = iota
	sortByCreated
	sortByModified
	sortBySize
)

type sortSpec struct {
	mode      sortMode
	ascending bool
}

func (s sortSpec) active() bool {
	return s.mode != sortByName
}

func parseSortSpec(s string) (sortSpec, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return sortSpec{}, nil
	}

	key := s
	ascending := false
	if i := strings.LastIndex(s, ":"); i >= 0 {
		key = s[:i]
		switch s[i+1:] {
		case "asc", "a", "ascending":
			ascending = true
		case "desc", "d", "descending":
			ascending = false
		default:
			return sortSpec{}, fmt.Errorf("unknown sort direction %q (use asc or desc)", s[i+1:])
		}
	}

	var mode sortMode
	switch key {
	case "size", "s":
		mode = sortBySize
	case "created", "date-created", "dc", "birth":
		mode = sortByCreated
	case "modified", "date-modified", "dm", "mtime", "m":
		mode = sortByModified
	default:
		return sortSpec{}, fmt.Errorf("unknown sort key %q (size, created, modified)", key)
	}
	return sortSpec{mode: mode, ascending: ascending}, nil
}

type packedFile struct {
	path     string
	relPath  string
	name     string
	lang     string
	created  time.Time
	modified time.Time
	size     int64
}

func isReadmeLike(name string) bool {
	switch strings.ToLower(name) {
	case "readme", "readme.md", "readme.markdown", "readme.mdown",
		"readme.mdx", "readme.rst", "readme.txt", "readme.org":
		return true
	default:
		return false
	}
}

func pathDepth(p string) int {
	p = filepath.ToSlash(p)
	if p == "" || p == "." {
		return 0
	}
	return strings.Count(strings.TrimSuffix(p, "/"), "/")
}

func lessPacked(a, b packedFile, spec sortSpec, pinReadme bool) bool {
	if pinReadme {
		ar, br := isReadmeLike(a.name), isReadmeLike(b.name)
		if ar != br {
			return ar
		}
		if ar && br {
			da, db := pathDepth(a.relPath), pathDepth(b.relPath)
			if da != db {
				return da < db
			}
		}
	}

	switch spec.mode {
	case sortByCreated:
		if !a.created.Equal(b.created) {
			if spec.ascending {
				return a.created.Before(b.created)
			}
			return a.created.After(b.created)
		}
	case sortByModified:
		if !a.modified.Equal(b.modified) {
			if spec.ascending {
				return a.modified.Before(b.modified)
			}
			return a.modified.After(b.modified)
		}
	case sortBySize:
		if a.size != b.size {
			if spec.ascending {
				return a.size < b.size
			}
			return a.size > b.size
		}
	}

	return false
}

func sortPackedFiles(files []packedFile, spec sortSpec, pinReadme bool) {
	if !spec.active() && !pinReadme {
		return
	}
	sort.SliceStable(files, func(i, j int) bool {
		return lessPacked(files[i], files[j], spec, pinReadme)
	})
}

func sortDirEntries(dir string, entries []os.DirEntry, spec sortSpec, pinReadme bool) {
	type wrap struct {
		e        os.DirEntry
		created  time.Time
		modified time.Time
		size     int64
	}
	wraps := make([]wrap, len(entries))
	for i, e := range entries {
		wraps[i].e = e
		info, err := e.Info()
		if err == nil {
			p := filepath.Join(dir, e.Name())
			wraps[i].created, wraps[i].modified = fileTimes(p, info)
			wraps[i].size = info.Size()
		}
	}
	sort.SliceStable(wraps, func(i, j int) bool {
		a, b := wraps[i], wraps[j]
		if a.e.IsDir() != b.e.IsDir() {
			return a.e.IsDir()
		}
		pin := pinReadme && !a.e.IsDir()
		pa := packedFile{relPath: a.e.Name(), name: a.e.Name(), created: a.created, modified: a.modified, size: a.size}
		pb := packedFile{relPath: b.e.Name(), name: b.e.Name(), created: b.created, modified: b.modified, size: b.size}
		if lessPacked(pa, pb, spec, pin) {
			return true
		}
		if lessPacked(pb, pa, spec, pin) {
			return false
		}
		return a.e.Name() < b.e.Name()
	})
	for i := range wraps {
		entries[i] = wraps[i].e
	}
}
