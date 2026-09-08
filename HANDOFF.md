# f2md Handoff

**Date**: 2026-09-09  
**Repository**: https://github.com/Arcioth/f2md (Public, Apache-2.0)  
**Binary Location**: `~/.local/bin/f2md` (Symlinked: `~/.local/bin/folder2md`)

---

## 1. Project Overview

`f2md` is a zero-dependency, ultra-fast CLI utility in pure Go (standard library only) designed to inspect, tree-index, and pack any codebase or directory into a structured Markdown document optimized for LLM context injection.

---

## 2. Key Architecture & Improvements

### Anti-Ouroboros Protection
- Resolves the absolute path of the output file upfront (`absOutput`).
- Automatically ignores the destination file in both tree generation (`buildTree`) and directory walking (`filepath.WalkDir`).
- Multiple consecutive runs will never re-ingest or recursively duplicate the context output.

### Markdown Fence Integrity
- Dynamic fence generator (`getCodeFence`) scans byte content for runs of backticks.
- If a target file contains ```` ``` ```` (e.g. `README.md`), the enclosing fence scales to `4+` backticks (```` ```` ````). Markdown renderers and LLMs will never experience premature fence closure.

### Hybrid Filetype Detection & Broad Coverage
- **Fast-Reject Set**: Common binary formats (`.png`, `.jpg`, `.pdf`, `.zip`, `.so`, `.exe`, `.tar`, `.wasm`, etc.) are rejected immediately by extension with zero disk I/O.
- **Native Extension Map**: 100+ languages and formats mapped directly to their markdown syntax tags.
- **Custom Rule Formats**: Explicitly supports `.acv`, `.ctr`, and `.rule` (Daetron / Arclinkdae configs) mapped to `ini` syntax highlighting.
- **Content Sniffing Fallback**: Unrecognized or extensionless files (`Dockerfile`, `Makefile`, `Justfile`, `LICENSE`) are sniffed via the first 512 bytes for null bytes (`0x00`) and UTF-8 validity.

### Noise & Lockfile Suppression
- Automatically skips huge auto-generated lockfiles (`package-lock.json`, `Cargo.lock`, `yarn.lock`, `pnpm-lock.yaml`, `flake.lock`, `poetry.lock`, etc.) via `--skip-locks` (enabled by default).
- Excludes sourcemaps (`*.map`), minified files (`*.min.js`, `*.min.css`), `.env*` secrets, and standard cache/vendor directories (`.git`, `node_modules`, `target`, `dist`, `.venv`, `.cache`, `.gradle`, etc.).
- Excludes binary artifacts (`f2md`, `folder2md`).

### Flexible CLI & Dynamic Naming
- **Flag Positioning**: Uses `reorderArgs` to allow flags to appear anywhere on the command line (e.g. `f2md ~/myproject -fn -d`).
- **Dynamic Filenames**:
  - `-fn`, `--folder-name`: Include directory name in output.
  - `-d`, `--date`: Include date (`YYYY-MM-DD`).
  - `-nd`, `--nodate`: Explicitly disable date.
  - `-t`, `--time`: Include time (`HH-MM-SS`).
- **Short & Long Aliases**: Every option supports both formats (`-s` / `--stdout`, `-m` / `--max-size`, `-sl` / `--skip-locks`, `-h` / `--help`).

---

## 3. Git Status & Releases

- **`v0.1.0`**: Initial clean rewrite, zero dependencies, bug fixes (ouroboros, fence collision, binary detection), Apache-2.0 license, and Makefile.
- **`v0.2.0`**: Added `--date` / `-d`, `--nodate` / `-nd`, `--time` / `-t`, `--folder-name` / `-fn`, argument reordering, and updated `~/.local/bin/f2md`.

---

## 4. Maintenance & Build Commands

```bash
# Rebuild and install locally to ~/.local/bin
make install

# Clean binary artifacts
make clean

# Run tests/help
f2md -h
```

Signed-off-by: Antigravity (Google DeepMind) & Arcioth
