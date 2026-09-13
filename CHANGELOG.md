# Changelog

## 0.3.0 — 2026-09-13

- `-S` / `--sort KEY[:asc|:desc]` with keys `size`, `created`, `modified`. Default direction is descending (largest / newest first).
- `-F` / `--format md|txt`. Plain text has no Markdown fences. Format is also inferred from `-o` (`.md` / `.txt`).
- `-fn` still names the file after the folder, and pins README variants at the top of the tree and contents.
- Linux created-time uses `statx` btime when available.

## 0.2.0

- `--date` / `-d`, `--nodate` / `-nd`, `--time` / `-t`, `--folder-name` / `-fn`, argument reordering.

## 0.1.0

- Initial release: zero-dependency packer, anti-ouroboros, dynamic fences, hybrid text detection.
