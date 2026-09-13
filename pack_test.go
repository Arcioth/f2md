package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPackDirectoryPinsReadmeWhenRequested(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "zzz.go"), "package zzz\n")
	mustWrite(t, filepath.Join(dir, "README.md"), "# hi\n")
	mustWrite(t, filepath.Join(dir, "aaa.go"), "package aaa\n")

	var buf strings.Builder
	err := packDirectory(dir, packOptions{skipLocks: true, maxSizeKB: 500, pinReadme: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	readmeAt := strings.Index(out, "## File: `README.md`")
	zzzAt := strings.Index(out, "## File: `zzz.go`")
	aaaAt := strings.Index(out, "## File: `aaa.go`")
	if readmeAt < 0 || zzzAt < 0 || aaaAt < 0 {
		t.Fatalf("missing file headers in output:\n%s", out)
	}
	if !(readmeAt < zzzAt && readmeAt < aaaAt) {
		t.Fatalf("README.md not first in file contents, offsets r=%d z=%d a=%d\n%s", readmeAt, zzzAt, aaaAt, out)
	}
}

func TestPackDirectoryTreePinsReadmeAmongFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "zzz.go"), "package zzz\n")
	mustWrite(t, filepath.Join(dir, "README.md"), "# hi\n")
	mustWrite(t, filepath.Join(dir, "aaa.go"), "package aaa\n")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	err := packDirectory(dir, packOptions{skipLocks: true, maxSizeKB: 500, pinReadme: true}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	treeEnd := strings.Index(out, "# File Contents")
	if treeEnd < 0 {
		t.Fatalf("no file contents heading:\n%s", out)
	}
	tree := out[:treeEnd]
	readmeAt := strings.Index(tree, "README.md")
	aaaAt := strings.Index(tree, "aaa.go")
	zzzAt := strings.Index(tree, "zzz.go")
	if readmeAt < 0 || aaaAt < 0 || zzzAt < 0 {
		t.Fatalf("missing names in tree:\n%s", tree)
	}
	if !(readmeAt < aaaAt && readmeAt < zzzAt) {
		t.Fatalf("README.md not first among files in tree:\n%s", tree)
	}
}

func TestPackDirectorySortsByModifiedNewestFirst(t *testing.T) {
	dir := t.TempDir()
	oldp := filepath.Join(dir, "old.go")
	newp := filepath.Join(dir, "new.go")
	mustWrite(t, oldp, "package old\n")
	mustWrite(t, newp, "package new\n")
	oldT := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	newT := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	if err := os.Chtimes(oldp, oldT, oldT); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newp, newT, newT); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	err := packDirectory(dir, packOptions{skipLocks: true, maxSizeKB: 500, sort: sortSpec{mode: sortByModified, ascending: false}}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	newAt := strings.Index(out, "## File: `new.go`")
	oldAt := strings.Index(out, "## File: `old.go`")
	if newAt < 0 || oldAt < 0 {
		t.Fatalf("missing file headers:\n%s", out)
	}
	if newAt > oldAt {
		t.Fatalf("expected new.go before old.go, offsets new=%d old=%d\n%s", newAt, oldAt, out)
	}
}

func TestPackDirectorySortsByCreatedNewestFirst(t *testing.T) {
	dir := t.TempDir()
	oldp := filepath.Join(dir, "old.go")
	mustWrite(t, oldp, "package old\n")
	time.Sleep(25 * time.Millisecond)
	newp := filepath.Join(dir, "new.go")
	mustWrite(t, newp, "package new\n")

	oldInfo, err := os.Stat(oldp)
	if err != nil {
		t.Fatal(err)
	}
	newInfo, err := os.Stat(newp)
	if err != nil {
		t.Fatal(err)
	}
	oldCreated, _ := fileTimes(oldp, oldInfo)
	newCreated, _ := fileTimes(newp, newInfo)
	if !newCreated.After(oldCreated) {
		t.Skip("filesystem birth times are not distinct")
	}

	var buf strings.Builder
	err = packDirectory(dir, packOptions{skipLocks: true, maxSizeKB: 500, sort: sortSpec{mode: sortByCreated, ascending: false}}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	newAt := strings.Index(out, "## File: `new.go`")
	oldAt := strings.Index(out, "## File: `old.go`")
	if newAt < 0 || oldAt < 0 {
		t.Fatalf("missing file headers:\n%s", out)
	}
	if newAt > oldAt {
		t.Fatalf("expected new.go before old.go by created time, offsets new=%d old=%d\n%s", newAt, oldAt, out)
	}
}

func TestPackDirectorySortsBySizeDescending(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "small.go"), "package s\n")
	mustWrite(t, filepath.Join(dir, "big.go"), "package b\n"+strings.Repeat("x", 4000)+"\n")

	var buf strings.Builder
	err := packDirectory(dir, packOptions{skipLocks: true, maxSizeKB: 500, sort: sortSpec{mode: sortBySize, ascending: false}}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	bigAt := strings.Index(out, "## File: `big.go`")
	smallAt := strings.Index(out, "## File: `small.go`")
	if bigAt < 0 || smallAt < 0 {
		t.Fatalf("missing file headers:\n%s", out)
	}
	if bigAt > smallAt {
		t.Fatalf("expected big.go before small.go, offsets big=%d small=%d\n%s", bigAt, smallAt, out)
	}
}

func TestPackDirectoryPlainTextHasNoFences(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "README.md"), "# hi\n")
	mustWrite(t, filepath.Join(dir, "main.go"), "package main\n")

	var buf strings.Builder
	err := packDirectory(dir, packOptions{
		skipLocks: true,
		maxSizeKB: 500,
		pinReadme: true,
		format:    formatText,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "```") {
		t.Fatalf("plain text output still has markdown fences:\n%s", out)
	}
	if !strings.Contains(out, "----- README.md -----") {
		t.Fatalf("missing text file banner:\n%s", out)
	}
	readmeAt := strings.Index(out, "----- README.md -----")
	mainAt := strings.Index(out, "----- main.go -----")
	if readmeAt < 0 || mainAt < 0 || readmeAt > mainAt {
		t.Fatalf("README should come first in txt output:\n%s", out)
	}
}
