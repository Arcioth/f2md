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

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-o` | `string` | `"project_context.md"` | Output file path (use `"-"` for stdout) |
| `-stdout` | `bool` | `false` | Stream Markdown directly to standard output |
| `-max-size` | `int` | `500` | Maximum file size in KB to include (0 for unlimited) |
| `-skip-locks` | `bool` | `true` | Skip package-lock, Cargo.lock, yarn.lock, etc. |
| `-h`, `--help` | `bool` | `false` | Show help and exit |

---

## Examples

### 1. Basic Export
Pack the current directory into `project_context.md`:
```bash
f2md
```

### 2. Export Another Directory
Pack a specific project into a custom file:
```bash
f2md -o ~/Desktop/daetron_context.md ~/Documents/daetron
```

### 3. Direct to Clipboard
Pipe project context straight to your system clipboard:

**Wayland:**
```bash
f2md -stdout | wl-copy
```

**X11:**
```bash
f2md -stdout | xclip -selection clipboard
```

**macOS:**
```bash
f2md -stdout | pbcopy
```

### 4. Custom Limits
Include files up to 2MB and retain package lockfiles:
```bash
f2md -max-size 2048 -skip-locks=false
```

---

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
