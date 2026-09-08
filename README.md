# f2md

> Fast, zero-dependency CLI tool that packages any directory or repository into a single, clean, LLM-ready Markdown document.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Dependencies](https://img.shields.io/badge/dependencies-0-brightgreen)
![License](https://img.shields.io/badge/License-Apache%202.0-blue)

`f2md` inspects your project, prints an ASCII directory tree, and concatenates all relevant source and config files into structured Markdown code blocks. It is engineered specifically for passing clean context into LLMs (Gemini, Claude, GPT) without context bloat, syntax breakage, or dependency overhead.

---

## Features

- ⚡ **Zero External Dependencies**: Built strictly using the Go standard library. Instant startup (<2ms) and trivial compilation to a single static binary.
- 🌳 **Clean Directory Tree**: Generates a clean, sorted ASCII tree structure (`├── `, `└── `) with directories grouped before files.
- 🛡️ **Markdown-Safe Fencing**: Calculates the longest run of backticks in each file and expands the code fence (e.g. ```` ```` ````) dynamically. Markdown docs containing backticks will never break outer code blocks or corrupt LLM parsing.
- 🚫 **Self-Inclusion Safe (Anti-Ouroboros)**: Resolves absolute output paths and excludes the output target from both the tree and the file walk, preventing recursive duplication across runs.
- 🧹 **Noise & Lockfile Filtering**: By default, excludes noise directories (`.git`, `node_modules`, `dist`, `target`, `build`, `.venv`, `.cache`, etc.) and huge auto-generated lockfiles (`package-lock.json`, `Cargo.lock`, `yarn.lock`, etc.).
- 🔒 **Secrets Protection**: Automatically ignores `.env` and `.env.*` files to prevent leaking API keys or credentials.
- 🔍 **Hybrid Text & Binary Detection**:
  - Instant rejection for known binary assets (`.png`, `.jpg`, `.pdf`, `.zip`, `.so`, `.exe`, etc.) with zero disk I/O.
  - Native syntax highlighting tags for 100+ languages and configs (including C, C++, Rust, Go, Zig, TypeScript, Python, and custom rule types like `.acv`, `.ctr`, `.rule`).
  - Automatic null-byte (`0x00`) content sniffing for extensionless text files (`Dockerfile`, `Makefile`, `Justfile`, `LICENSE`).
- 📋 **Clipboard & Pipeline Ready**: Supports `-stdout` / `-o -` for direct clipboard integration on Wayland, X11, or macOS.

---

## Installation

### From Source

```bash
git clone https://github.com/Arcioth/f2md.git
cd f2md
make install
```

This builds the binary and installs it into `~/.local/bin/f2md` (with a `folder2md` symlink). Ensure `~/.local/bin` is in your `$PATH`.

### Using `go install`

```bash
go install github.com/Arcioth/f2md@latest
```

---

## Usage

```text
f2md [options] [directory]
```

### Options

| Flag | Short | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--folder-name` | `-fn` | `bool` | `false` | Include target directory name in output filename |
| `--date` | `-d` | `bool` | `false` | Include current date (`YYYY-MM-DD`) in output filename |
| `--nodate` | `-nd` | `bool` | `false` | Force exclude date from output filename |
| `--time` | `-t` | `bool` | `false` | Include current time (`HH-MM-SS`) in output filename |
| `--output` | `-o` | `string` | `""` | Explicit output file path (use `"-"` for stdout) |
| `--stdout` | `-s` | `bool` | `false` | Stream Markdown directly to standard output |
| `--max-size` | `-m` | `int` | `500` | Maximum file size in KB to include (0 for unlimited) |
| `--skip-locks` | `-sl` | `bool` | `true` | Skip package-lock, Cargo.lock, yarn.lock, etc. |
| `--noskip-locks`| `-nsl`| `bool` | `false` | Retain package lockfiles in export |
| `--help` | `-h` | `bool` | `false` | Show help message and exit |

---

## Examples

### 1. Basic Export
Pack the current directory into `project_context.md`:
```bash
f2md
```

### 2. Include Folder Name & Date in Filename
Pack with folder name and date (e.g. `daetron_2026-09-08_context.md`):
```bash
f2md -fn -d
```

### 3. Include Folder Name, Date & Time
Pack with folder name, date, and timestamp (e.g. `daetron_2026-09-08_22-30-15_context.md`):
```bash
f2md -fn -d -t
```

### 4. Export Another Directory
Pack a specific project into a custom file:
```bash
f2md -o ~/Desktop/daetron_context.md ~/Documents/daetron
```

### 5. Direct to Clipboard
Pipe project context straight to your system clipboard:

**Wayland:**
```bash
f2md -s | wl-copy
```

**X11:**
```bash
f2md -s | xclip -selection clipboard
```

**macOS:**
```bash
f2md -s | pbcopy
```

### 6. Custom Limits
Include files up to 2MB and retain package lockfiles:
```bash
f2md -m 2048 -nsl
```

---

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
