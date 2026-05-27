# CLI reference

This page lists every option exposed by `gophercram`. For grouped
walkthroughs see the [usage guide](usage.md).

```
gophercram [options] [path ...]
```

If no positional path is given, GopherCram packs the current directory.

## Logging

| Flag | Effect |
|---|---|
| `--verbose` | Print detailed progress and debug information. |
| `--quiet` | Suppress non-error console output. |
| `--stdout` | Write the packed output to stdout (implies `--quiet`). |
| `--stdin` | Read file paths to include from stdin, one per line. |
| `--copy` | Copy the packed output to the system clipboard. |
| `--top-files-len N` | Number of largest files (by tokens) to show in the run report. |
| `--token-count-tree [N]` | After packing, print a flat per-file token count list; optional threshold filters small files. |

## Output

| Flag | Effect |
|---|---|
| `-o, --output FILE` | Output file path (use `-` for stdout). |
| `--style STYLE` | Output style: `xml`, `markdown`, `json`, or `plain`. |
| `--parsable-style` | Escape file content so the output remains valid XML/Markdown. |
| `--compress` | Compress code by keeping signatures and dropping bodies. |
| `--output-show-line-numbers` | Prefix every line in the output with its line number. |
| `--no-file-summary` | Omit the summary section. |
| `--no-directory-structure` | Omit the directory tree. |
| `--no-files` | Omit file contents (metadata-only output). |
| `--remove-comments` | Strip comments from supported languages. |
| `--remove-empty-lines` | Drop blank lines from file content. |
| `--truncate-base64` | Replace long base64 blobs with a placeholder. |
| `--header-text TEXT` | Custom text prepended to the packed output. |
| `--instruction-file-path PATH` | Path to a file whose contents are appended to the output. |
| `--split-output SIZE` | Split the output into multiple numbered files (e.g. `2mb`). |
| `--include-empty-directories` | Include empty directories in the tree view. |
| `--include-full-directory-structure` | Render the full tree even when `--include` narrows the file list. |
| `--no-git-sort-by-changes` | Don't reorder files by git change frequency. |
| `--include-diffs` | Append working-tree and staged diffs to the output. |
| `--include-logs` | Append a recent git log. |
| `--include-logs-count N` | Number of commits to include when `--include-logs` is set. |

## File selection

| Flag | Effect |
|---|---|
| `--include PATTERNS` | Comma-separated glob patterns to include (repeatable). |
| `-i, --ignore PATTERNS` | Comma-separated glob patterns to ignore (repeatable). |
| `--no-gitignore` | Don't apply `.gitignore` rules. |
| `--no-dot-ignore` | Don't apply `.ignore` / `.gophercramignore` files. |
| `--no-default-patterns` | Don't apply the built-in default ignore patterns. |
| `--max-file-size N` | Skip files larger than N bytes (default 52428800). |

## Remote repositories

| Flag | Effect |
|---|---|
| `--remote URL` | Clone and pack a remote repository (URL or `owner/repo`). |
| `--remote-branch NAME` | Branch, tag, or commit to check out from the remote. |
| `--remote-trust-config` | Trust `gophercram.config.json` files inside the cloned repo. |

## Configuration

| Flag | Effect |
|---|---|
| `-c, --config PATH` | Path to a config file. |
| `--init` | Generate a default config file and exit. |
| `--global` | With `--init`, write the config to the user-global location. |

## Security

| Flag | Effect |
|---|---|
| `--no-security-check` | Don't scan for embedded secrets. |

## Token counting

| Flag | Effect |
|---|---|
| `--token-count-encoding NAME` | Token estimator algorithm (currently `approx` only). |

## Meta

| Flag | Effect |
|---|---|
| `-v, --version` | Print the GopherCram version and exit. |
| `-h, --help` | Print the help screen and exit. |

## Exit codes

| Code | Meaning |
|------|---------|
| `0`  | Pack completed successfully. |
| `1`  | A runtime error occurred (file system, git, parse, …). |
| `2`  | Argument parsing failed. |

## Conventions

- Boolean flags may also be passed as `--flag=true`, `--flag=false`,
  `--flag=1`, etc.
- Use `--` to mark the end of options; subsequent tokens are treated as
  paths even if they begin with `-`.
- Size flags accept `b`, `kb`, `mb`, `gb` suffixes (decimal-friendly:
  `2.5mb`).

## Worked examples

```bash
# Pack everything under src/, drop binary files, write Markdown
gophercram src/ --style markdown -o src.md

# Strip comments, then compress for context-window economy
gophercram --remove-comments --compress -o slim.xml

# Pack only changes since the last commit, plus a git log
gophercram --include-diffs --include-logs --include-logs-count 20

# Stream into another tool
gophercram --stdout --style markdown | gh gist create --filename context.md

# Use a custom config for a sub-package
gophercram --config configs/frontend.json client/
```
