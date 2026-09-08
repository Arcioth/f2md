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

func printHelp() {
	helpText := `f2md - Fast, zero-dependency codebase to Markdown context packer

Usage:
  f2md [options] [directory]

Options:
  -o string
        Output file path (default "project_context.md", use '-' for stdout)
  -stdout
        Write directly to stdout instead of a file
  -max-size int
        Maximum file size in KB to include (default 500, 0 for unlimited)
  -skip-locks
        Skip dependency lockfiles like package-lock.json, Cargo.lock, etc. (default true)
  -h, --help
        Show this help message

Examples:
  f2md                          # Pack current directory into project_context.md
  f2md path/to/project          # Pack specified project directory
  f2md -o context.md .          # Save to custom output file
  f2md -stdout . | wl-copy      # Pipe directly to Wayland clipboard
  f2md -stdout . | xclip        # Pipe directly to X11 clipboard
  f2md -max-size 1000 .         # Allow files up to 1 MB
`
	fmt.Print(helpText)
}

func main() {
	helpFlag := flag.Bool("help", false, "Show help message")
	hFlag := flag.Bool("h", false, "Show help message")
	outputFlag := flag.String("o", "project_context.md", "Output file path (use '-' for stdout)")
	stdoutFlag := flag.Bool("stdout", false, "Write directly to stdout instead of a file")
	maxSizeKB := flag.Int64("max-size", 500, "Maximum file size in KB to include (0 for unlimited)")
	skipLocksFlag := flag.Bool("skip-locks", true, "Skip package-lock, yarn.lock, Cargo.lock and other generated lockfiles")

	flag.Usage = printHelp
	flag.Parse()

	if *helpFlag || *hFlag {
		printHelp()
		os.Exit(0)
	}

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

	toStdout := *stdoutFlag || *outputFlag == "-"

	var absOutput string
	var out io.Writer

	if toStdout {
		out = os.Stdout
	} else {
		resolvedOutput, err := filepath.Abs(*outputFlag)
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

	rootDirName := filepath.Base(absRoot)

	// 1. Write Directory Tree
	fmt.Fprintf(writer, "# Project Directory Structure\n\n```text\n%s/\n", rootDirName)
	var treeBuilder strings.Builder
	if err := buildTree(absRoot, "", absOutput, *skipLocksFlag, &treeBuilder); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to build full tree: %v\n", err)
	}
	writer.WriteString(treeBuilder.String())
	writer.WriteString("```\n\n---\n\n# File Contents\n\n")

	maxSizeBytes := *maxSizeKB * 1024

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
			if path != absRoot && shouldIgnore(name, true, *skipLocksFlag) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnore(name, false, *skipLocksFlag) {
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
			fmt.Fprintf(writer, "<!-- Skipped: file size (%d KB) exceeds limit (%d KB) -->\n\n---\n\n", info.Size()/1024, *maxSizeKB)
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

	if !toStdout {
		fmt.Printf("Successfully generated %s\n", *outputFlag)
	}
}
