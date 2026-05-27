# Configuration reference

GopherCram is configured by layering, in order:

1. Built-in defaults (see `internal/config/config.go::Defaults`).
2. Per-user global file at `$XDG_CONFIG_HOME/gophercram/gophercram.config.json`
   (or `~/.config/gophercram/gophercram.config.json` on Unix-like systems,
   or `%APPDATA%/GopherCram/gophercram.config.json` on Windows).
3. Project-root file: `gophercram.config.json`. As a fallback, the legacy
   filenames `gophercram.json` and `repomix.config.json` are also tried.
4. CLI flags.

You can override the project-root location with `--config <path>`.

Use `gophercram --init` to write a defaulted config file in the current
directory. `gophercram --init --global` writes the user-global file.

## Top-level shape

```jsonc
{
  "$schema": "https://gophercram.dev/schema.json",
  "input": { /* ... */ },
  "output": { /* ... */ },
  "include": ["..."],
  "ignore": { /* ... */ },
  "security": { /* ... */ },
  "tokenCount": { /* ... */ }
}
```

Every block is optional; omitted fields inherit the default.

## `input`

| Field         | Type    | Default     | Description                                          |
|---------------|---------|-------------|------------------------------------------------------|
| `maxFileSize` | integer | `52428800`  | Skip individual files larger than this many bytes.  |

## `output`

| Field                              | Type                | Default                  | Description                                                                   |
|------------------------------------|---------------------|--------------------------|-------------------------------------------------------------------------------|
| `filePath`                         | string              | `gophercram-output.xml` | Output file path (use `"-"` for stdout).                                      |
| `style`                            | `"xml" \| "markdown" \| "json" \| "plain"` | `"xml"` | Output format.                                                              |
| `parsableStyle`                    | bool                | `false`                  | Escape content so the output remains valid XML/Markdown.                      |
| `headerText`                       | string              | `""`                     | Custom text prepended to the output.                                          |
| `instructionFilePath`              | string              | `""`                     | Path whose contents are appended as an `<instruction>` block.                 |
| `fileSummary`                      | bool                | `true`                   | Include the metadata summary section.                                         |
| `directoryStructure`               | bool                | `true`                   | Include the directory tree.                                                   |
| `files`                            | bool                | `true`                   | Include file contents (set `false` to produce a metadata-only listing).       |
| `removeComments`                   | bool                | `false`                  | Strip comments from supported languages.                                      |
| `removeEmptyLines`                 | bool                | `false`                  | Drop blank lines from file content.                                           |
| `compress`                         | bool                | `false`                  | Apply the signature-keeping compression pass.                                 |
| `topFilesLength`                   | integer             | `5`                      | How many largest-by-token files to list in the summary.                       |
| `showLineNumbers`                  | bool                | `false`                  | Prefix every output line with its 1-based number.                             |
| `truncateBase64`                   | bool                | `false`                  | Replace long base64 blobs with a placeholder.                                 |
| `copyToClipboard`                  | bool                | `false`                  | Pipe the rendered output to the system clipboard.                             |
| `includeEmptyDirectories`          | bool                | `false`                  | Surface empty directories in the tree.                                        |
| `includeFullDirectoryStructure`    | bool                | `false`                  | Render the full tree even when `include` narrows the file list.               |
| `splitOutput`                      | integer (bytes)     | `0`                      | When >0, split the output into numbered parts of this size.                   |
| `git`                              | object              | see below                | Git-aware behaviour.                                                          |

### `output.git`

| Field                       | Type    | Default | Description                                                       |
|-----------------------------|---------|---------|-------------------------------------------------------------------|
| `sortByChanges`             | bool    | `true`  | Order files by how often they change in recent commits.           |
| `sortByChangesMaxCommits`   | integer | `100`   | How many commits to scan for change-frequency calculations.       |
| `includeDiffs`              | bool    | `false` | Append working-tree and staged diffs to the output.               |
| `includeLogs`               | bool    | `false` | Append recent commit log to the output.                            |
| `includeLogsCount`          | integer | `50`    | How many commits to include with `includeLogs`.                    |

## `include`

A list of glob patterns. If empty, all non-ignored files are packed. If
non-empty, only files matching one of the patterns are packed.

```jsonc
"include": ["src/**/*.ts", "docs/**/*.md"]
```

Patterns support `*`, `**`, `?`, character classes (`[abc]`, `[a-z]`,
negation with `!`), and anchored patterns starting with `/`.

## `ignore`

| Field               | Type    | Default | Description                                                                    |
|---------------------|---------|---------|--------------------------------------------------------------------------------|
| `useGitignore`      | bool    | `true`  | Apply `.gitignore` rules encountered while walking.                            |
| `useDotIgnore`      | bool    | `true`  | Apply `.ignore` and `.gophercramignore` files.                                 |
| `useDefaultPatterns`| bool    | `true`  | Apply GopherCram's built-in ignore patterns.                                   |
| `customPatterns`    | string[]| `[]`    | Additional patterns (gitignore syntax). Merged with the rules above.           |

## `security`

| Field                  | Type | Default | Description                                                  |
|------------------------|------|---------|--------------------------------------------------------------|
| `enableSecurityCheck`  | bool | `true`  | Scan packed files for embedded secrets; drop matching files. |

## `tokenCount`

| Field      | Type   | Default    | Description                                                |
|------------|--------|------------|------------------------------------------------------------|
| `encoding` | string | `"approx"` | Token estimator algorithm. Only `"approx"` is supported.   |

## CLI ↔ config mapping

The CLI uses the same configuration model under the hood. Each flag
shadows a single field; the table below documents the mapping.

| Flag                                | Config path                              |
|-------------------------------------|------------------------------------------|
| `--output FILE`                     | `output.filePath`                        |
| `--style STYLE`                     | `output.style`                           |
| `--compress`                        | `output.compress`                        |
| `--remove-comments`                 | `output.removeComments`                  |
| `--remove-empty-lines`              | `output.removeEmptyLines`                |
| `--truncate-base64`                 | `output.truncateBase64`                  |
| `--output-show-line-numbers`        | `output.showLineNumbers`                 |
| `--no-file-summary`                 | `output.fileSummary = false`             |
| `--no-directory-structure`          | `output.directoryStructure = false`      |
| `--no-files`                        | `output.files = false`                   |
| `--header-text TEXT`                | `output.headerText`                      |
| `--instruction-file-path PATH`      | `output.instructionFilePath`             |
| `--split-output SIZE`               | `output.splitOutput`                     |
| `--include-empty-directories`       | `output.includeEmptyDirectories`         |
| `--include-full-directory-structure`| `output.includeFullDirectoryStructure`   |
| `--git-sort-by-changes`             | `output.git.sortByChanges`               |
| `--include-diffs`                   | `output.git.includeDiffs`                |
| `--include-logs`                    | `output.git.includeLogs`                 |
| `--include-logs-count N`            | `output.git.includeLogsCount`            |
| `--include PATTERNS`                | `include`                                |
| `--ignore PATTERNS`                 | `ignore.customPatterns`                  |
| `--no-gitignore`                    | `ignore.useGitignore = false`            |
| `--no-dot-ignore`                   | `ignore.useDotIgnore = false`            |
| `--no-default-patterns`             | `ignore.useDefaultPatterns = false`      |
| `--max-file-size N`                 | `input.maxFileSize`                      |
| `--no-security-check`               | `security.enableSecurityCheck = false`   |
| `--token-count-encoding NAME`       | `tokenCount.encoding`                    |
| `--copy`                            | `output.copyToClipboard = true`          |
| `--top-files-len N`                 | `output.topFilesLength`                  |
