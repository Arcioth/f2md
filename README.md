# f2md

Pack a directory into one Markdown or plain-text file: an ASCII tree plus the text files, for pasting into an LLM. Pure Go, standard library only.

## Status

Works. Tested on Linux. Birth-time sort (`--sort created`) uses `statx` btime when the filesystem provides it; otherwise ctime.

## Requirements

- Go 1.22 or later to build
- `~/.local/bin` on `$PATH` if you `make install`

## Install

```bash
git clone https://github.com/Arcioth/f2md.git
cd f2md
make install
```

That installs `~/.local/bin/f2md` and a `folder2md` symlink.

```bash
go install github.com/Arcioth/f2md@latest
```

## CLI

```text
f2md [options] [directory]
```

Default write target is `project_context.md` in the current directory. The output file is never packed into itself.

| Flag | Short | Default | Meaning |
| --- | --- | --- | --- |
| `--folder-name` | `-fn` | off | Put the directory name in the output filename. Also pins `README.md` (and README variants) at the top of the tree and file list. |
| `--date` | `-d` | off | Add `YYYY-MM-DD` to the output filename. |
| `--nodate` | `-nd` | off | Do not add a date even if `-d` is set. |
| `--time` | `-t` | off | Add `HH-MM-SS` to the output filename. |
| `--output` | `-o` | generated name | Explicit path. `-` is stdout. A `.txt` / `.md` suffix sets the format unless `-F` is given. |
| `--format` | `-F` | `md` | `md` or `txt`. `txt` is a fence-free dump with `----- path -----` banners. |
| `--sort` | `-S` | name / walk order | `size`, `created`, or `modified`. Default direction is descending (largest / newest first). Append `:asc` to reverse. |
| `--stdout` | `-s` | off | Write to stdout. |
| `--max-size` | `-m` | `500` | Skip file bodies over this many KB (`0` = unlimited). |
| `--skip-locks` | `-sl` | on | Skip `package-lock.json`, `Cargo.lock`, `yarn.lock`, and similar. |
| `--noskip-locks` | `-nsl` | off | Keep lockfiles. |
| `--help` | `-h` | | Help. |

Flags may appear anywhere (`f2md ~/proj -fn -d`).

Ignored by default: `.git`, `node_modules`, `dist`, `target`, `.venv`, `.cache`, hidden names, `.env*`, sourcemaps, minified JS/CSS, and the `f2md` / `folder2md` binaries. Known binary extensions are skipped without a full read; unknown files are sniffed for NUL bytes and UTF-8.

Markdown output sizes the code fence to one backtick longer than the longest run inside each file, so nested fences do not close the block early.

## Examples

```bash
f2md
f2md -fn -d
f2md -fn -S created
f2md -S size:asc
f2md -S modified -F txt
f2md -o dump.txt ~/Documents/daetron
f2md -s | wl-copy
f2md -m 2048 -nsl
```

## Docs

README this file · HANDOFF.md session notes · CHANGELOG.md releases · LICENSE Apache-2.0
