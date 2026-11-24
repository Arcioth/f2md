package main

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func isZipPath(p string) bool {
	return strings.EqualFold(filepath.Ext(p), ".zip")
}

func cleanZipEntryPath(raw string) (string, bool, bool) {
	norm := strings.ReplaceAll(raw, "\\", "/")
	isDir := strings.HasSuffix(norm, "/")
	norm = strings.TrimPrefix(norm, "/")
	if norm == "" {
		return "", false, false
	}

	rawParts := strings.Split(norm, "/")
	parts := make([]string, 0, len(rawParts))
	for _, seg := range rawParts {
		if seg == "" || seg == "." {
			continue
		}
		if seg == ".." {
			return "", false, false
		}
		parts = append(parts, seg)
	}
	if len(parts) == 0 {
		return "", false, false
	}
	return strings.Join(parts, "/"), isDir, true
}

func shouldIgnoreZipPath(relPath string, isDir bool, skipLocks bool) bool {
	parts := strings.Split(relPath, "/")
	if len(parts) == 0 {
		return true
	}
	dirLimit := len(parts) - 1
	if isDir {
		dirLimit = len(parts)
	}
	for i := 0; i < dirLimit; i++ {
		if shouldIgnore(parts[i], true, skipLocks) {
			return true
		}
	}
	if !isDir {
		if shouldIgnore(parts[len(parts)-1], false, skipLocks) {
			return true
		}
	}
	return false
}

// detectZipWrapper returns the single common top-level directory name if all
// non-ignored entries in the archive live inside one wrapper folder (such as
// GitHub's "repo-main/..."), or "" otherwise.
func detectZipWrapper(files []*zip.File, skipLocks bool) string {
	var topDir string
	hasNested := false

	for _, zf := range files {
		cleaned, isDir, ok := cleanZipEntryPath(zf.Name)
		if !ok {
			continue
		}
		if zf.FileInfo().IsDir() {
			isDir = true
		}

		parts := strings.Split(cleaned, "/")
		first := parts[0]
		firstIsDir := isDir || len(parts) > 1
		if shouldIgnore(first, firstIsDir, skipLocks) {
			continue
		}

		if !firstIsDir {
			// A non-ignored file exists directly at the zip root.
			return ""
		}
		if topDir == "" {
			topDir = first
		} else if topDir != first {
			return ""
		}
		if len(parts) > 1 {
			hasNested = true
		}
	}

	if topDir != "" && hasNested {
		return topDir
	}
	return ""
}

func resolveZipRoot(zipPath string, files []*zip.File, skipLocks bool) (rootName string, stripPrefix string) {
	if wrapper := detectZipWrapper(files, skipLocks); wrapper != "" {
		return wrapper, wrapper + "/"
	}
	base := filepath.Base(zipPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if stem == "" || stem == "." || stem == "/" {
		stem = "archive"
	}
	return stem, ""
}

func zipRootName(zipPath string) (string, error) {
	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	rootName, _ := resolveZipRoot(zipPath, rc.File, true)
	return rootName, nil
}

type zipTreeNode struct {
	name     string
	isDir    bool
	size     int64
	modified time.Time
	children map[string]*zipTreeNode
}

func newZipDirNode(name string) *zipTreeNode {
	return &zipTreeNode{
		name:     name,
		isDir:    true,
		children: make(map[string]*zipTreeNode),
	}
}

func (n *zipTreeNode) insert(relPath string, isDir bool, size int64, modTime time.Time) {
	parts := strings.Split(relPath, "/")
	cur := n
	for i, seg := range parts {
		last := i == len(parts)-1
		if !last || isDir {
			child, ok := cur.children[seg]
			if !ok {
				child = newZipDirNode(seg)
				cur.children[seg] = child
			}
			child.isDir = true
			if last && !modTime.IsZero() {
				child.modified = modTime
			}
			cur = child
		} else {
			cur.children[seg] = &zipTreeNode{
				name:     seg,
				isDir:    false,
				size:     size,
				modified: modTime,
			}
		}
	}
}

func renderZipTree(node *zipTreeNode, prefix string, opt packOptions, sb *strings.Builder) {
	if len(node.children) == 0 {
		return
	}
	children := make([]*zipTreeNode, 0, len(node.children))
	for _, c := range node.children {
		children = append(children, c)
	}

	sort.SliceStable(children, func(i, j int) bool {
		a, b := children[i], children[j]
		if a.isDir != b.isDir {
			return a.isDir
		}
		pin := opt.pinReadme && !a.isDir
		pa := packedFile{relPath: a.name, name: a.name, created: a.modified, modified: a.modified, size: a.size}
		pb := packedFile{relPath: b.name, name: b.name, created: b.modified, modified: b.modified, size: b.size}
		if lessPacked(pa, pb, opt.sort, pin) {
			return true
		}
		if lessPacked(pb, pa, opt.sort, pin) {
			return false
		}
		return a.name < b.name
	})

	for i, child := range children {
		connector := "├── "
		extension := "│   "
		if i == len(children)-1 {
			connector = "└── "
			extension = "    "
		}
		sb.WriteString(prefix + connector + child.name + "\n")
		if child.isDir {
			renderZipTree(child, prefix+extension, opt, sb)
		}
	}
}

func packZip(zipPath string, opt packOptions, w io.Writer) error {
	rc, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer rc.Close()

	return packZipReader(zipPath, &rc.Reader, opt, w)
}

func packZipReader(zipPath string, zr *zip.Reader, opt packOptions, w io.Writer) error {
	rootDirName, stripPrefix := resolveZipRoot(zipPath, zr.File, opt.skipLocks)

	treeRoot := newZipDirNode(rootDirName)

	type zipEntry struct {
		pf packedFile
		zf *zip.File
	}
	var entries []zipEntry

	for _, zf := range zr.File {
		cleaned, isDir, ok := cleanZipEntryPath(zf.Name)
		if !ok {
			continue
		}
		if zf.FileInfo().IsDir() {
			isDir = true
		}

		relPath := cleaned
		if stripPrefix != "" {
			if cleaned == strings.TrimSuffix(stripPrefix, "/") {
				continue
			}
			if !strings.HasPrefix(cleaned, stripPrefix) {
				// Top-level sibling that was ignored (e.g. __MACOSX)
				continue
			}
			relPath = strings.TrimPrefix(cleaned, stripPrefix)
			if relPath == "" {
				continue
			}
		}

		if shouldIgnoreZipPath(relPath, isDir, opt.skipLocks) {
			continue
		}

		modTime := zf.Modified
		size := int64(zf.UncompressedSize64)
		treeRoot.insert(relPath, isDir, size, modTime)

		if isDir {
			continue
		}

		name := relPath
		if slash := strings.LastIndex(relPath, "/"); slash >= 0 {
			name = relPath[slash+1:]
		}
		ext := strings.ToLower(filepath.Ext(name))
		base := strings.ToLower(name)

		isText, lang := detectTextWithOpener(base, ext, func() (io.ReadCloser, error) {
			return zf.Open()
		})
		if !isText {
			continue
		}

		entries = append(entries, zipEntry{
			pf: packedFile{
				path:     relPath,
				relPath:  relPath,
				name:     name,
				lang:     lang,
				created:  modTime,
				modified: modTime,
				size:     size,
			},
			zf: zf,
		})
	}

	var treeBuilder strings.Builder
	renderZipTree(treeRoot, "", opt, &treeBuilder)

	if opt.format == formatText {
		fmt.Fprintf(w, "Project Directory Structure\n\n%s/\n", rootDirName)
		io.WriteString(w, treeBuilder.String())
		io.WriteString(w, "\nFile Contents\n\n")
	} else {
		fmt.Fprintf(w, "# Project Directory Structure\n\n```text\n%s/\n", rootDirName)
		io.WriteString(w, treeBuilder.String())
		io.WriteString(w, "```\n\n---\n\n# File Contents\n\n")
	}

	// Establish deterministic lexicographical walk order before applying sort/pinning.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].pf.relPath < entries[j].pf.relPath
	})

	if opt.sort.active() || opt.pinReadme {
		sort.SliceStable(entries, func(i, j int) bool {
			return lessPacked(entries[i].pf, entries[j].pf, opt.sort, opt.pinReadme)
		})
	}

	maxSizeBytes := opt.maxSizeKB * 1024

	for _, entry := range entries {
		f := entry.pf
		if maxSizeBytes > 0 && f.size > maxSizeBytes {
			writeSkipped(w, opt.format, f.relPath, f.size, opt.maxSizeKB)
			continue
		}

		rc, openErr := entry.zf.Open()
		if openErr != nil {
			writeReadError(w, opt.format, f.relPath, openErr)
			continue
		}

		var content []byte
		var readErr error
		if maxSizeBytes > 0 {
			content, readErr = io.ReadAll(io.LimitReader(rc, maxSizeBytes+1))
		} else {
			content, readErr = io.ReadAll(rc)
		}
		rc.Close()

		if readErr != nil {
			writeReadError(w, opt.format, f.relPath, readErr)
			continue
		}
		if maxSizeBytes > 0 && int64(len(content)) > maxSizeBytes {
			actualSize := f.size
			if actualSize < int64(len(content)) {
				actualSize = int64(len(content))
			}
			writeSkipped(w, opt.format, f.relPath, actualSize, opt.maxSizeKB)
			continue
		}

		writeFileBody(w, opt.format, f, content)
	}

	return nil
}
