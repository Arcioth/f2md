package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)


var defaultIgnoreDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"target":       true,
	"dist":         true,
	"build":        true,
	".next":        true,
	"__pycache__":  true,
	".turbo":       true,
	".nuxt":        true,
	"out":          true,
	"bin":          true,
	"obj":          true,
	".venv":        true,
	"venv":         true,
	"env":          true,
	"vendor":       true,
	".idea":        true,
	".vscode":      true,
	".cache":       true,
	".gradle":      true,
	"coverage":     true,
	".nyc_output":  true,
	"tmp":          true,
}

var commonTextFilenames = map[string]string{
	"dockerfile":    "dockerfile",
	"containerfile": "dockerfile",
	"makefile":      "makefile",
	"justfile":      "makefile",
	"gemfile":       "ruby",
	"procfile":      "yaml",
	"vagrantfile":   "ruby",
	"rakefile":      "ruby",
	"cmakelists.txt":"cmake",
	"license":       "text",
	"licence":       "text",
	"copying":       "text",
	"readme":        "markdown",
	"caddyfile":     "caddyfile",
	"earthfile":     "earthfile",
	"authors":       "text",
	"contributors":  "text",
}

var lockFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"cargo.lock":        true,
	"composer.lock":     true,
	"poetry.lock":       true,
	"gemfile.lock":      true,
	"mix.lock":          true,
	"flake.lock":        true,
	"bun.lockb":         true,
}

var binaryExtensions = map[string]bool{
	// Images
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".ico": true,
	".bmp": true, ".tiff": true, ".tif": true, ".avif": true, ".heic": true, ".psd": true,
	// Audio / Video
	".mp3": true, ".wav": true, ".ogg": true, ".flac": true, ".aac": true, ".m4a": true,
	".mp4": true, ".mkv": true, ".webm": true, ".mov": true, ".avi": true, ".flv": true,
	// Archives & Packages
	".zip": true, ".tar": true, ".gz": true, ".tgz": true, ".bz2": true, ".xz": true,
	".7z": true, ".rar": true, ".iso": true, ".zst": true, ".deb": true, ".rpm": true,
	// Executables, Libraries & Bytecode
	".exe": true, ".bin": true, ".dll": true, ".so": true, ".dylib": true, ".a": true,
	".lib": true, ".o": true, ".obj": true, ".class": true, ".pyc": true, ".pyo": true,
	".wasm": true,
	// Documents & Fonts
	".pdf": true, ".docx": true, ".xlsx": true, ".pptx": true, ".odt": true, ".ods": true,
	".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
	// Databases
	".db": true, ".sqlite": true, ".sqlite3": true, ".parquet": true,
}

var textExtensions = map[string]string{
	// User custom / Daetron / Arclinkdae rule formats
	".acv":          "ini",
	".ctr":          "ini",
	".rule":         "ini",

	// Systems & Low-level
	".c":            "c",
	".h":            "c",
	".cpp":          "cpp",
	".cc":           "cpp",
	".cxx":          "cpp",
	".hpp":          "cpp",
	".hxx":          "cpp",
	".hh":           "cpp",
	".inl":          "cpp",
	".rs":           "rust",
	".go":           "go",
	".zig":          "zig",
	".d":            "d",
	".nim":          "nim",
	".v":            "verilog",
	".sv":           "verilog",
	".svh":          "verilog",
	".vhd":          "vhdl",
	".vhdl":         "vhdl",
	".s":            "assembly",
	".asm":          "assembly",
	".nasm":         "assembly",

	// Web & Scripting
	".js":           "javascript",
	".mjs":          "javascript",
	".cjs":          "javascript",
	".jsx":          "jsx",
	".ts":           "typescript",
	".mts":          "typescript",
	".cts":          "typescript",
	".tsx":          "tsx",
	".vue":          "vue",
	".svelte":       "svelte",
	".astro":        "astro",
	".html":         "html",
	".htm":          "html",
	".xhtml":        "html",
	".css":          "css",
	".scss":         "scss",
	".sass":         "sass",
	".less":         "less",
	".py":           "python",
	".pyi":          "python",
	".pyw":          "python",
	".rb":           "ruby",
	".rake":         "ruby",
	".gemspec":      "ruby",
	".php":          "php",
	".phtml":        "php",
	".lua":          "lua",
	".pl":           "perl",
	".pm":           "perl",
	".tcl":          "tcl",
	".awk":          "awk",
	".sed":          "sed",
	".sh":           "bash",
	".bash":         "bash",
	".zsh":          "zsh",
	".fish":         "fish",
	".ps1":          "powershell",
	".psm1":         "powershell",
	".bat":          "batch",
	".cmd":          "batch",
	".r":            "r",
	".jl":           "julia",
	".dart":         "dart",

	// Functional & BEAM
	".ex":           "elixir",
	".exs":          "elixir",
	".erl":          "erlang",
	".hrl":          "erlang",
	".clj":          "clojure",
	".cljs":         "clojure",
	".cljc":         "clojure",
	".edn":          "clojure",
	".hs":           "haskell",
	".lhs":          "haskell",
	".ml":           "ocaml",
	".mli":          "ocaml",
	".lisp":         "lisp",
	".lsp":          "lisp",
	".scm":          "lisp",
	".rkt":          "racket",

	// JVM & .NET
	".java":         "java",
	".kt":           "kotlin",
	".kts":          "kotlin",
	".scala":        "scala",
	".sc":           "scala",
	".groovy":       "groovy",
	".gradle":       "groovy",
	".cs":           "csharp",
	".fs":           "fsharp",
	".fsx":          "fsharp",
	".vb":           "vb",
	".swift":        "swift",

	// Config, Data & Schemas
	".json":         "json",
	".json5":        "json",
	".jsonc":        "json",
	".yaml":         "yaml",
	".yml":          "yaml",
	".toml":         "toml",
	".xml":          "xml",
	".xsl":          "xml",
	".xslt":         "xml",
	".svg":          "xml",
	".sql":          "sql",
	".graphql":      "graphql",
	".gql":          "graphql",
	".proto":        "protobuf",
	".prisma":       "prisma",
	".nix":          "nix",
	".ini":          "ini",
	".cfg":          "ini",
	".conf":         "ini",
	".properties":   "ini",
	".desktop":      "ini",
	".env.example":  "dotenv",
	".env.template": "dotenv",
	".csv":          "csv",
	".tsv":          "tsv",
	".diff":         "diff",
	".patch":        "diff",

	// Documentation & Markup
	".md":           "markdown",
	".markdown":     "markdown",
	".mdown":        "markdown",
	".mdx":          "markdown",
	".txt":          "text",
	".rst":          "rst",
	".adoc":         "asciidoc",
	".asciidoc":     "asciidoc",
	".tex":          "latex",
	".latex":        "latex",
	".org":          "org",
}

// shouldIgnore filters out hidden files, ignored directories, lockfiles, and minified noise.
func shouldIgnore(name string, isDir bool, skipLocks bool) bool {
	if isDir {
		return defaultIgnoreDirs[name] || strings.HasPrefix(name, ".")
	}
	// Skip hidden files (.DS_Store, .gitconfig, etc.)
	if strings.HasPrefix(name, ".") {
		return true
	}
	lower := strings.ToLower(name)
	// Skip self binary artifacts
	if lower == "f2md" || lower == "folder2md" {
		return true
	}
	// Safeguard: skip sensitive secret files
	if lower == ".env" || strings.HasPrefix(lower, ".env.") {
		return true
	}
	// Skip sourcemaps and minified files
	if strings.HasSuffix(lower, ".map") || strings.HasSuffix(lower, ".min.js") || strings.HasSuffix(lower, ".min.css") {
		return true
	}
	// Skip huge dependency lockfiles if requested
	if skipLocks && lockFiles[lower] {
		return true
	}
	return false
}

// detectTextFile determines if a file is text (checking extensions or sniffing first 512 bytes)
// and returns the markdown language tag.
func detectTextFile(path string, base string, ext string) (bool, string) {
	if binaryExtensions[ext] {
		return false, ""
	}
	if lang, ok := commonTextFilenames[base]; ok {
		return true, lang
	}
	if lang, ok := textExtensions[ext]; ok {
		return true, lang
	}

	// Sniff unknown file: read first 512 bytes to test for null byte
	f, err := os.Open(path)
	if err != nil {
		return false, ""
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, ""
	}
	if n == 0 {
		return true, "text"
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return false, ""
		}
	}
	if !utf8.Valid(buf[:n]) {
		return false, ""
	}

	trimmed := strings.TrimPrefix(ext, ".")
	if trimmed != "" {
		return true, trimmed
	}
	return true, "text"
}

// getCodeFence calculates the closing/opening fence (e.g. ```, ````, `````)
// ensuring it has more backticks than the longest run inside the file to prevent markdown corruption.
func getCodeFence(content []byte) string {
	maxRun := 0
	currentRun := 0
	for _, b := range content {
		if b == '`' {
			currentRun++
			if currentRun > maxRun {
				maxRun = currentRun
			}
		} else {
			currentRun = 0
		}
	}

	fenceLen := 3
	if maxRun >= 3 {
		fenceLen = maxRun + 1
	}
	return strings.Repeat("`", fenceLen)
}

// buildTree creates an ASCII tree visualization of the project.
func buildTree(root string, prefix string, absOutput string, skipLocks bool, sb *strings.Builder) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	var validEntries []os.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		if shouldIgnore(name, entry.IsDir(), skipLocks) {
			continue
		}

		entryAbsPath := filepath.Join(root, name)
		if absOutput != "" && entryAbsPath == absOutput {
			continue
		}

		validEntries = append(validEntries, entry)
	}

	sort.Slice(validEntries, func(i, j int) bool {
		if validEntries[i].IsDir() != validEntries[j].IsDir() {
			return validEntries[i].IsDir()
		}
		return validEntries[i].Name() < validEntries[j].Name()
	})

	for i, entry := range validEntries {
		connector := "├── "
		extension := "│   "
		if i == len(validEntries)-1 {
			connector = "└── "
			extension = "    "
		}

		sb.WriteString(prefix + connector + entry.Name() + "\n")

		if entry.IsDir() {
			subPath := filepath.Join(root, entry.Name())
			if err := buildTree(subPath, prefix+extension, absOutput, skipLocks, sb); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateOutputFilename(rootDirName string, useFolderName bool, useDate bool, useTime bool) string {
	now := time.Now()
	var parts []string

	if useFolderName && rootDirName != "" && rootDirName != "." && rootDirName != "/" {
		parts = append(parts, rootDirName)
	} else {
		parts = append(parts, "project")
	}

	if useDate {
		parts = append(parts, now.Format("2006-01-02"))
	}

	if useTime {
		parts = append(parts, now.Format("15-04-05"))
	}

	parts = append(parts, "context.md")
	return strings.Join(parts, "_")
}

func printHelp() {
	helpText := `f2md - Fast, zero-dependency codebase to Markdown context packer

Usage:
  f2md [options] [directory]

Output Naming Options:
  -fn,  --folder-name     Include folder name in output filename (e.g. <folder>_context.md)
  -d,   --date            Include current date in output filename (YYYY-MM-DD)
  -nd,  --nodate          Do not include date in output filename
  -t,   --time            Include current time in output filename (HH-MM-SS)
  -o,   --output string   Explicit output file path (use '-' for stdout)

Processing Options:
  -s,   --stdout          Stream Markdown directly to stdout instead of a file
  -m,   --max-size int    Maximum file size in KB to include (default 500, 0 for unlimited)
  -sl,  --skip-locks      Skip dependency lockfiles (default true)
  -nsl, --noskip-locks    Do not skip dependency lockfiles
  -h,   --help            Show this help message

Examples:
  f2md                             # Pack current directory -> project_context.md
  f2md -fn                         # Pack with folder name  -> <folder>_context.md
  f2md -fn -d                      # Pack with date         -> <folder>_YYYY-MM-DD_context.md
  f2md -fn -d -t                   # Pack with date & time  -> <folder>_YYYY-MM-DD_HH-MM-SS_context.md
  f2md -d path/to/project          # Pack specified directory with date
  f2md -o custom.md                # Write to custom output file
  f2md -s | wl-copy                # Stream to Wayland clipboard
  f2md -s | xclip -sel clip        # Stream to X11 clipboard
  f2md -m 1000                     # Allow files up to 1 MB
`
	fmt.Print(helpText)
}

func reorderArgs(raw []string) []string {
	var flags []string
	var positional []string

	valueFlags := map[string]bool{
		"-o": true, "--output": true, "-output": true,
		"-m": true, "-ms": true, "--max-size": true, "-max-size": true,
	}

	for i := 0; i < len(raw); i++ {
		arg := raw[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if valueFlags[arg] && i+1 < len(raw) && !strings.HasPrefix(raw[i+1], "-") {
				flags = append(flags, raw[i+1])
				i++
			}
		} else {
			positional = append(positional, arg)
		}
	}
	return append(flags, positional...)
}

func main() {
	var (
		useDate       bool
		noDate        bool
		useTime       bool
		useFolderName bool
		outputFile    string
		toStdout      bool
		maxSizeKB     int64 = 500
		skipLocks     bool  = true
		noSkipLocks   bool
		showHelp      bool
	)

	flag.BoolVar(&useDate, "date", false, "Include current date (YYYY-MM-DD) in output filename")
	flag.BoolVar(&useDate, "d", false, "Include current date (YYYY-MM-DD) in output filename")

	flag.BoolVar(&noDate, "nodate", false, "Do not include date in output filename")
	flag.BoolVar(&noDate, "nd", false, "Do not include date in output filename")

	flag.BoolVar(&useTime, "time", false, "Include current time (HH-MM-SS) in output filename")
	flag.BoolVar(&useTime, "t", false, "Include current time (HH-MM-SS) in output filename")

	flag.BoolVar(&useFolderName, "folder-name", false, "Include folder name in output filename")
	flag.BoolVar(&useFolderName, "fn", false, "Include folder name in output filename")

	flag.StringVar(&outputFile, "output", "", "Output file path (use '-' for stdout)")
	flag.StringVar(&outputFile, "o", "", "Output file path (use '-' for stdout)")

	flag.BoolVar(&toStdout, "stdout", false, "Write directly to stdout instead of a file")
	flag.BoolVar(&toStdout, "s", false, "Write directly to stdout instead of a file")

	flag.Int64Var(&maxSizeKB, "max-size", 500, "Maximum file size in KB to include (0 for unlimited)")
	flag.Int64Var(&maxSizeKB, "m", 500, "Maximum file size in KB to include (0 for unlimited)")
	flag.Int64Var(&maxSizeKB, "ms", 500, "Maximum file size in KB to include (0 for unlimited)")

	flag.BoolVar(&skipLocks, "skip-locks", true, "Skip dependency lockfiles like package-lock.json, Cargo.lock, etc.")
	flag.BoolVar(&skipLocks, "sl", true, "Skip dependency lockfiles like package-lock.json, Cargo.lock, etc.")

	flag.BoolVar(&noSkipLocks, "noskip-locks", false, "Do not skip lockfiles")
	flag.BoolVar(&noSkipLocks, "nsl", false, "Do not skip lockfiles")

	flag.BoolVar(&showHelp, "help", false, "Show this help message")
	flag.BoolVar(&showHelp, "h", false, "Show this help message")

	flag.Usage = printHelp
	flag.CommandLine.Parse(reorderArgs(os.Args[1:]))

	if showHelp {
		printHelp()
		os.Exit(0)
	}

	finalDate := useDate && !noDate
	finalSkipLocks := skipLocks && !noSkipLocks

	targetDir := "."
	args := flag.Args()
	if len(args) > 0 {
		targetDir = args[0]
	}

	absRoot, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	stat, err := os.Stat(absRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing path: %v\n", err)
		os.Exit(1)
	}
	if !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "Target path is not a directory: %s\n", absRoot)
		os.Exit(1)
	}

	rootDirName := filepath.Base(absRoot)

	isStdout := toStdout || outputFile == "-"

	var absOutput string
	var out io.Writer
	var chosenOutput string

	if isStdout {
		out = os.Stdout
	} else {
		if outputFile == "" {
			chosenOutput = generateOutputFilename(rootDirName, useFolderName, finalDate, useTime)
		} else {
			chosenOutput = outputFile
			if fi, err := os.Stat(chosenOutput); err == nil && fi.IsDir() {
				genName := generateOutputFilename(rootDirName, useFolderName, finalDate, useTime)
				chosenOutput = filepath.Join(chosenOutput, genName)
			}
		}

		resolvedOutput, err := filepath.Abs(chosenOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving output path: %v\n", err)
			os.Exit(1)
		}
		absOutput = resolvedOutput

		if err := os.MkdirAll(filepath.Dir(absOutput), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}

		file, err := os.Create(absOutput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		out = file
	}

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	// 1. Write Directory Tree
	fmt.Fprintf(writer, "# Project Directory Structure\n\n```text\n%s/\n", rootDirName)
	var treeBuilder strings.Builder
	if err := buildTree(absRoot, "", absOutput, finalSkipLocks, &treeBuilder); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to build full tree: %v\n", err)
	}
	writer.WriteString(treeBuilder.String())
	writer.WriteString("```\n\n---\n\n# File Contents\n\n")

	maxSizeBytes := maxSizeKB * 1024

	// 2. Walk and Append File Contents
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Prevent self-inclusion
		if absOutput != "" && path == absOutput {
			return nil
		}

		name := d.Name()

		if d.IsDir() {
			if path != absRoot && shouldIgnore(name, true, finalSkipLocks) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnore(name, false, finalSkipLocks) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		base := strings.ToLower(name)

		// Fast text check without reading entire file
		isText, lang := detectTextFile(path, base, ext)
		if !isText {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			relPath = path
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Skip oversized files to protect memory & LLM context
		if maxSizeBytes > 0 && info.Size() > maxSizeBytes {
			fmt.Fprintf(writer, "## File: `%s`\n\n", relPath)
			fmt.Fprintf(writer, "<!-- Skipped: file size (%d KB) exceeds limit (%d KB) -->\n\n---\n\n", info.Size()/1024, maxSizeKB)
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(writer, "## File: `%s`\n\n", relPath)
			fmt.Fprintf(writer, "Error reading file: %v\n\n---\n\n", err)
			return nil
		}

		fence := getCodeFence(content)

		fmt.Fprintf(writer, "## File: `%s`\n\n", relPath)
		fmt.Fprintf(writer, "%s%s\n", fence, lang)
		writer.Write(content)
		if len(content) > 0 && content[len(content)-1] != '\n' {
			writer.WriteByte('\n')
		}
		fmt.Fprintf(writer, "%s\n\n---\n\n", fence)

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if !isStdout {
		fmt.Printf("Successfully generated %s\n", chosenOutput)
	}
}
