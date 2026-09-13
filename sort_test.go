package main

import (
	"testing"
	"time"
)

func TestIsReadmeLike(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"README.md", true},
		{"readme", true},
		{"ReadMe.RST", true},
		{"README.markdown", true},
		{"README.txt", true},
		{"README.org", true},
		{"README.mdx", true},
		{"CHANGELOG.md", false},
		{"main.go", false},
		{"README.md.bak", false},
		{"not-readme.md", false},
	}
	for _, c := range cases {
		if got := isReadmeLike(c.name); got != c.want {
			t.Errorf("isReadmeLike(%q)=%v want %v", c.name, got, c.want)
		}
	}
}

func namesOf(files []packedFile) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.name
	}
	return out
}

func TestSortPinsReadmeFirst(t *testing.T) {
	files := []packedFile{
		{relPath: "zzz.go", name: "zzz.go"},
		{relPath: "README.md", name: "README.md"},
		{relPath: "aaa.go", name: "aaa.go"},
	}
	sortPackedFiles(files, sortSpec{}, true)

	if files[0].name != "README.md" {
		t.Fatalf("first file = %q, want README.md", files[0].name)
	}
	if files[1].name != "zzz.go" || files[2].name != "aaa.go" {
		t.Fatalf("remaining order not stable: %q, %q", files[1].name, files[2].name)
	}
}

func TestSortRootReadmeBeforeNestedReadme(t *testing.T) {
	files := []packedFile{
		{relPath: "docs/README.md", name: "README.md"},
		{relPath: "README.md", name: "README.md"},
	}
	sortPackedFiles(files, sortSpec{}, true)
	if files[0].relPath != "README.md" {
		t.Fatalf("first = %q, want root README.md", files[0].relPath)
	}
}

func TestSortByModifiedDescendingNewestFirst(t *testing.T) {
	old := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	mid := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	new := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	files := []packedFile{
		{relPath: "old.go", name: "old.go", modified: old},
		{relPath: "new.go", name: "new.go", modified: new},
		{relPath: "mid.go", name: "mid.go", modified: mid},
	}
	sortPackedFiles(files, sortSpec{mode: sortByModified, ascending: false}, false)
	got := namesOf(files)
	want := []string{"new.go", "mid.go", "old.go"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestSortByModifiedAscendingOldestFirst(t *testing.T) {
	old := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	new := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	files := []packedFile{
		{relPath: "new.go", name: "new.go", modified: new},
		{relPath: "old.go", name: "old.go", modified: old},
	}
	sortPackedFiles(files, sortSpec{mode: sortByModified, ascending: true}, false)
	if files[0].name != "old.go" {
		t.Fatalf("first = %q, want old.go", files[0].name)
	}
}

func TestSortByCreatedDescendingNewestFirst(t *testing.T) {
	old := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	new := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	files := []packedFile{
		{relPath: "old.go", name: "old.go", created: old},
		{relPath: "new.go", name: "new.go", created: new},
	}
	sortPackedFiles(files, sortSpec{mode: sortByCreated, ascending: false}, false)
	if files[0].name != "new.go" {
		t.Fatalf("first = %q, want new.go", files[0].name)
	}
}

func TestSortByCreatedAscendingOldestFirst(t *testing.T) {
	old := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	new := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	files := []packedFile{
		{relPath: "new.go", name: "new.go", created: new},
		{relPath: "old.go", name: "old.go", created: old},
	}
	sortPackedFiles(files, sortSpec{mode: sortByCreated, ascending: true}, false)
	if files[0].name != "old.go" {
		t.Fatalf("first = %q, want old.go", files[0].name)
	}
}

func TestSortBySizeDescendingLargestFirst(t *testing.T) {
	files := []packedFile{
		{relPath: "small.go", name: "small.go", size: 10},
		{relPath: "big.go", name: "big.go", size: 1000},
		{relPath: "mid.go", name: "mid.go", size: 100},
	}
	sortPackedFiles(files, sortSpec{mode: sortBySize, ascending: false}, false)
	got := namesOf(files)
	want := []string{"big.go", "mid.go", "small.go"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestSortBySizeAscendingSmallestFirst(t *testing.T) {
	files := []packedFile{
		{relPath: "big.go", name: "big.go", size: 1000},
		{relPath: "small.go", name: "small.go", size: 10},
	}
	sortPackedFiles(files, sortSpec{mode: sortBySize, ascending: true}, false)
	if files[0].name != "small.go" {
		t.Fatalf("first = %q, want small.go", files[0].name)
	}
}

func TestSortReadmePinnedAheadOfNewerFiles(t *testing.T) {
	old := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	new := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	files := []packedFile{
		{relPath: "new.go", name: "new.go", created: new, modified: new, size: 999},
		{relPath: "README.md", name: "README.md", created: old, modified: old, size: 1},
	}
	sortPackedFiles(files, sortSpec{mode: sortByModified, ascending: false}, true)
	if files[0].name != "README.md" {
		t.Fatalf("first = %q, want README.md even when older", files[0].name)
	}
}

func TestParseSortSpec(t *testing.T) {
	cases := []struct {
		in        string
		mode      sortMode
		ascending bool
		wantErr   bool
	}{
		{"", sortByName, false, false},
		{"size", sortBySize, false, false},
		{"size:asc", sortBySize, true, false},
		{"size:desc", sortBySize, false, false},
		{"s:a", sortBySize, true, false},
		{"created", sortByCreated, false, false},
		{"date-created:asc", sortByCreated, true, false},
		{"modified:asc", sortByModified, true, false},
		{"mtime", sortByModified, false, false},
		{"nope", 0, false, true},
		{"size:sideways", 0, false, true},
	}
	for _, c := range cases {
		spec, err := parseSortSpec(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseSortSpec(%q): want error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSortSpec(%q): %v", c.in, err)
			continue
		}
		if spec.mode != c.mode || spec.ascending != c.ascending {
			t.Errorf("parseSortSpec(%q)=%+v, want mode=%v asc=%v", c.in, spec, c.mode, c.ascending)
		}
	}
}
