package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type outputFormat int

const (
	formatMarkdown outputFormat = iota
	formatText
)

func (f outputFormat) ext() string {
	if f == formatText {
		return "txt"
	}
	return "md"
}

func parseFormat(s string) (outputFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "md", "markdown", "mdown":
		return formatMarkdown, nil
	case "txt", "text", "plain", "plaintext":
		return formatText, nil
	default:
		return 0, fmt.Errorf("unknown format %q (md, txt)", s)
	}
}

func inferFormatFromPath(path string) (outputFormat, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".mdown":
		return formatMarkdown, true
	case ".txt", ".text":
		return formatText, true
	default:
		return formatMarkdown, false
	}
}
