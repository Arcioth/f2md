# f2md Handoff

**Date**: 2026-11-24  
**Repository**: https://github.com/Arcioth/f2md (Public, Apache-2.0)  
**Binary Location**: `~/.local/bin/f2md` (Symlinked: `~/.local/bin/folder2md`)  
**Android App**: `android/` (`com.arcioth.f2md`, signed APK attached to GitHub Releases)  
**Version**: 0.4.0

---

## 1. Project Overview

`f2md` is a zero-dependency CLI in pure Go (standard library only) and a native Kotlin Android app that pack a directory or `.zip` archive into one Markdown or plain-text document for LLM context.

README is the public landing page. This file is session state.

---

## 2. Current Features (v0.4.0)

- **ZIP input support** in both Go CLI (`zip.go`) and Android (`F2mdPacker.packZipFile`): unwraps single wrapper directories (e.g. GitHub `repo-main/`), filters `__MACOSX`, enforces size limits safely, supports all sort/format/README-pinning flags.
- **Android APK (`android/`)**:
  - Pure Kotlin engine (`F2mdPacker.kt`, `SafFolderPacker.kt`) with 1:1 CLI parity and JVM unit tests (`F2mdPackerTest.kt`).
  - Fast `DocumentsContract` single-cursor SAF folder scanning + `ZipFile` random-access archive packing.
  - Handles `ACTION_VIEW` & `ACTION_SEND` for `.zip` files.
  - Local offline HTML/CSS/JS WebView UI with Share (`FileProvider`), Save As (`ACTION_CREATE_DOCUMENT`), Copy, Tree/Files/Preview tabs.
- `-S` / `--sort KEY[:asc|:desc]` — keys `size`, `created`, `modified`. Default desc.
- `-F` / `--format md|txt` — also inferred from `-o` extension.
- `-fn` pins README variants to the top and names the output after the folder/archive.
- Created time: Linux `statx` btime, else ctime. Darwin birthtime. Windows CreationTime.

Tests: Go (`sort_test.go`, `pack_test.go`, `format_test.go`, `zip_test.go`), Android (`F2mdPackerTest.kt`).

---

## 3. Build

```bash
# CLI
nix-shell -p go --run "go test ./... && go build -o f2md ."
install -Dm755 f2md ~/.local/bin/f2md

# Android APK
cd android
NIXPKGS_ALLOW_UNFREE=1 NIXPKGS_ACCEPT_ANDROID_SDK_LICENSE=1 nix-shell shell.nix --run "gradle testDebugUnitTest assembleRelease"
```
