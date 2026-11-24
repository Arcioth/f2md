package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testZipFile struct {
	name    string
	body    []byte
	modTime time.Time
}

func makeTestZip(t *testing.T, path string, files []testZipFile) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for _, tf := range files {
		hdr := &zip.FileHeader{
			Name:     tf.name,
			Method:   zip.Deflate,
			Modified: tf.modTime,
		}
		if strings.HasSuffix(tf.name, "/") {
			hdr.Method = zip.Store
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			t.Fatal(err)
		}
		if len(tf.body) > 0 {
			if _, err := w.Write(tf.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPackZipUnwrapsSingleWrapperAndIgnoresMacOSX(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "download.zip")
	makeTestZip(t, zipPath, []testZipFile{
		{name: "myrepo-main/"},
		{name: "myrepo-main/zzz.go", body: []byte("package zzz\n")},
		{name: "myrepo-main/README.md", body: []byte("# My Repo\n")},
		{name: "myrepo-main/src/app.py", body: []byte("print('hi')\n")},
		{name: "myrepo-main/.git/config", body: []byte("secret\n")},
		{name: "myrepo-main/node_modules/pkg/index.js", body: []byte("x\n")},
		{name: "myrepo-main/package-lock.json", body: []byte("{}\n")},
		{name: "myrepo-main/logo.png", body: []byte{0x89, 0x50, 0x4E, 0x47, 0x00}},
		{name: "__MACOSX/myrepo-main/._README.md", body: []byte{0x00, 0x05}},
	})

	rootName, err := zipRootName(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	if rootName != "myrepo-main" {
		t.Fatalf("zipRootName = %q, want myrepo-main", rootName)
	}

	var buf strings.Builder
	err = packZip(zipPath, packOptions{
		skipLocks: true,
		maxSizeKB: 500,
		pinReadme: true,
		format:    formatMarkdown,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "myrepo-main/\n") {
		t.Fatalf("missing unwrapped root dir in tree:\n%s", out)
	}
	if strings.Contains(out, "__MACOSX") || strings.Contains(out, ".git") || strings.Contains(out, "node_modules") || strings.Contains(out, "package-lock.json") {
		t.Fatalf("ignored entries leaked into output:\n%s", out)
	}
	if strings.Contains(out, "## File: `logo.png`") {
		t.Fatalf("binary file content should not be packed:\n%s", out)
	}
	readmeAt := strings.Index(out, "## File: `README.md`")
	appAt := strings.Index(out, "## File: `src/app.py`")
	zzzAt := strings.Index(out, "## File: `zzz.go`")
	if readmeAt < 0 || appAt < 0 || zzzAt < 0 {
		t.Fatalf("missing expected files:\n%s", out)
	}
	if !(readmeAt < appAt && readmeAt < zzzAt) {
		t.Fatalf("README.md should be pinned first:\n%s", out)
	}
}

func TestPackZipFlatArchiveSortAndTextFormat(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "flatproj.zip")
	oldT := time.Date(2022, 1, 1, 10, 0, 0, 0, time.UTC)
	newT := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	makeTestZip(t, zipPath, []testZipFile{
		{name: "old.go", body: []byte("package old\n"), modTime: oldT},
		{name: "new.go", body: []byte("package new\n"), modTime: newT},
		{name: "huge.txt", body: []byte(strings.Repeat("a", 3000)), modTime: oldT},
	})

	var buf strings.Builder
	err := packZip(zipPath, packOptions{
		skipLocks: true,
		maxSizeKB: 1,
		sort:      sortSpec{mode: sortByModified, ascending: false},
		format:    formatText,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "flatproj/\n") {
		t.Fatalf("missing flatproj root:\n%s", out)
	}
	if strings.Contains(out, "```") {
		t.Fatalf("txt output must not contain fences:\n%s", out)
	}
	newAt := strings.Index(out, "----- new.go -----")
	oldAt := strings.Index(out, "----- old.go -----")
	if newAt < 0 || oldAt < 0 || newAt > oldAt {
		t.Fatalf("expected new.go before old.go:\n%s", out)
	}
	if !strings.Contains(out, "[skipped: file size (2 KB) exceeds limit (1 KB)]") {
		t.Fatalf("expected huge.txt to be skipped:\n%s", out)
	}
}

func TestPackZipDynamicFenceAndBinarySniff(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "notes.zip")
	makeTestZip(t, zipPath, []testZipFile{
		{name: "doc.md", body: []byte("```go\nfmt.Println()\n```\n")},
		{name: "rawdata.custom", body: []byte{0x41, 0x42, 0x00, 0x43}},
		{name: "textdata.custom", body: []byte("hello world\n")},
	})

	var buf strings.Builder
	err := packZip(zipPath, packOptions{
		skipLocks: true,
		maxSizeKB: 500,
		format:    formatMarkdown,
	}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "````markdown\n") {
		t.Fatalf("expected 4-backtick dynamic fence:\n%s", out)
	}
	if strings.Contains(out, "## File: `rawdata.custom`") {
		t.Fatalf("binary sniffed file should be skipped:\n%s", out)
	}
	if !strings.Contains(out, "## File: `textdata.custom`") {
		t.Fatalf("text sniffed file should be included:\n%s", out)
	}
}
