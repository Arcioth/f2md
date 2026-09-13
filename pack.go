package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type packOptions struct {
	skipLocks bool
	maxSizeKB int64
	absOutput string
	sort      sortSpec
	pinReadme bool
	format    outputFormat
}

func packDirectory(absRoot string, opt packOptions, w io.Writer) error {
	rootDirName := filepath.Base(absRoot)
	var treeBuilder strings.Builder
	if err := buildTree(absRoot, "", opt, &treeBuilder); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to build full tree: %v\n", err)
	}

	if opt.format == formatText {
		fmt.Fprintf(w, "Project Directory Structure\n\n%s/\n", rootDirName)
		io.WriteString(w, treeBuilder.String())
		io.WriteString(w, "\nFile Contents\n\n")
	} else {
		fmt.Fprintf(w, "# Project Directory Structure\n\n```text\n%s/\n", rootDirName)
		io.WriteString(w, treeBuilder.String())
		io.WriteString(w, "```\n\n---\n\n# File Contents\n\n")
	}

	maxSizeBytes := opt.maxSizeKB * 1024
	var files []packedFile

	err := filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if opt.absOutput != "" && path == opt.absOutput {
			return nil
		}

		name := d.Name()
		if d.IsDir() {
			if path != absRoot && shouldIgnore(name, true, opt.skipLocks) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldIgnore(name, false, opt.skipLocks) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		base := strings.ToLower(name)
		isText, lang := detectTextFile(path, base, ext)
		if !isText {
			return nil
		}

		relPath, relErr := filepath.Rel(absRoot, path)
		if relErr != nil {
			relPath = path
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}

		created, modified := fileTimes(path, info)
		files = append(files, packedFile{
			path:     path,
			relPath:  relPath,
			name:     name,
			lang:     lang,
			created:  created,
			modified: modified,
			size:     info.Size(),
		})
		return nil
	})
	if err != nil {
		return err
	}

	sortPackedFiles(files, opt.sort, opt.pinReadme)

	for _, f := range files {
		if maxSizeBytes > 0 && f.size > maxSizeBytes {
			writeSkipped(w, opt.format, f.relPath, f.size, opt.maxSizeKB)
			continue
		}

		content, readErr := os.ReadFile(f.path)
		if readErr != nil {
			writeReadError(w, opt.format, f.relPath, readErr)
			continue
		}

		writeFileBody(w, opt.format, f, content)
	}

	return nil
}

func writeSkipped(w io.Writer, format outputFormat, relPath string, size, maxKB int64) {
	if format == formatText {
		fmt.Fprintf(w, "----- %s -----\n", relPath)
		fmt.Fprintf(w, "[skipped: file size (%d KB) exceeds limit (%d KB)]\n\n", size/1024, maxKB)
		return
	}
	fmt.Fprintf(w, "## File: `%s`\n\n", relPath)
	fmt.Fprintf(w, "<!-- Skipped: file size (%d KB) exceeds limit (%d KB) -->\n\n---\n\n", size/1024, maxKB)
}

func writeReadError(w io.Writer, format outputFormat, relPath string, err error) {
	if format == formatText {
		fmt.Fprintf(w, "----- %s -----\n", relPath)
		fmt.Fprintf(w, "Error reading file: %v\n\n", err)
		return
	}
	fmt.Fprintf(w, "## File: `%s`\n\n", relPath)
	fmt.Fprintf(w, "Error reading file: %v\n\n---\n\n", err)
}

func writeFileBody(w io.Writer, format outputFormat, f packedFile, content []byte) {
	if format == formatText {
		fmt.Fprintf(w, "----- %s -----\n", f.relPath)
		w.Write(content)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			io.WriteString(w, "\n")
		}
		io.WriteString(w, "\n")
		return
	}

	fence := getCodeFence(content)
	fmt.Fprintf(w, "## File: `%s`\n\n", f.relPath)
	fmt.Fprintf(w, "%s%s\n", fence, f.lang)
	w.Write(content)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		io.WriteString(w, "\n")
	}
	fmt.Fprintf(w, "%s\n\n---\n\n", fence)
}
