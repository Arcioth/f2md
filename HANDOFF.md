# f2md Handoff

**Date**: 2026-09-13  
**Repository**: https://github.com/Arcioth/f2md (Public, Apache-2.0)  
**Binary Location**: `~/.local/bin/f2md` (Symlinked: `~/.local/bin/folder2md`)  
**Version**: 0.3.0

---

## 1. Project Overview

`f2md` is a zero-dependency CLI in pure Go (standard library only) that packs a directory into one Markdown or plain-text document for LLM context.

README is the public landing page. This file is session state.

---

## 2. Current CLI (v0.3.0)

- `-S` / `--sort KEY[:asc|:desc]` — keys `size`, `created`, `modified`. Default desc.
- `-F` / `--format md|txt` — also inferred from `-o` extension.
- `-fn` pins README variants to the top and names the output after the folder.
- Created time: Linux `statx` btime, else ctime. Darwin birthtime. Windows CreationTime.

Tests: `sort_test.go`, `pack_test.go`, `format_test.go`.

---

## 3. Git

- `v0.2.0` already on origin.
- `v0.3.0`: `--sort`, `--format txt`, README pin, README rewrite.

## 4. Next moves

- More export types (json, html): extend `parseFormat` + `writeFileBody`.
- `make` is not on NixOS default PATH; `go build` + `install` works.

## 5. Build

```bash
# needs Go on PATH
make install
f2md -h
```

Signed-off-by: Grok & Arcioth
