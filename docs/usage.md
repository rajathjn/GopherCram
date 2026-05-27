# Usage guide

This page walks through the common ways to drive GopherCram from the
command line. For the full option list see the
[CLI reference](cli-reference.md).

## Packing a local project

```bash
gophercram
```

With no arguments, GopherCram packs the current directory and writes
`gophercram-output.xml`. It honours `.gitignore`, `.ignore`,
`.gophercramignore`, and the built-in default patterns (`node_modules/`,
`dist/`, `.git/`, log files, OS junk, …).

You can pass one or more positional paths:

```bash
gophercram src/ docs/ README.md
```

Each path may be a directory or a single file. Directories are walked
recursively.

## Choosing an output style

```bash
gophercram --style markdown   # writes gophercram-output.md
gophercram --style json       # writes gophercram-output.json
gophercram --style plain      # writes gophercram-output.txt
gophercram --style xml        # default
```

Override the output filename with `--output`:

```bash
gophercram --style markdown --output ./build/context.md
```

Pass `--output -` (or `--stdout`) to stream the packed file to stdout,
which is handy when piping into another tool:

```bash
gophercram --stdout --style markdown | pbcopy
```

## Filtering files

### Include only certain paths

```bash
gophercram --include 'src/**/*.ts,tests/**/*.test.ts'
```

`--include` accepts a comma-separated list of glob patterns. The flag is
repeatable; multiple invocations are concatenated.

### Ignore additional patterns

```bash
gophercram --ignore '*.log,scratch/**'
```

`--ignore` is also comma-separated and repeatable.

### Disable defaults

If you want a totally bare run:

```bash
gophercram --no-default-patterns --no-gitignore --no-dot-ignore .
```

### Per-directory ignore files

GopherCram automatically picks up:

- `.gitignore` (unless `--no-gitignore`)
- `.ignore` and `.gophercramignore` (unless `--no-dot-ignore`)

Rules in these files use standard gitignore syntax: `*.log`, `build/`,
`!keep-me.log`, anchored patterns starting with `/`, and so on.

## Content transforms

| Flag                              | Effect                                                |
|----------------------------------|-------------------------------------------------------|
| `--remove-comments`              | Strip comments using a language-aware lexer.          |
| `--remove-empty-lines`           | Drop blank lines from file contents.                  |
| `--truncate-base64`              | Replace long base64 blobs with a short placeholder.   |
| `--compress`                     | Keep declarations/signatures, drop function bodies.   |
| `--output-show-line-numbers`     | Prefix every line with its 1-based number.            |

Transforms compose; e.g. `--remove-comments --compress` first strips
comments, then leaves the trimmed declarations.

## Splitting the output

Long packed files can be too big to share. Use `--split-output` to break
the result into numbered parts:

```bash
gophercram --split-output 2mb --output ctx.xml
# produces ctx.1.xml, ctx.2.xml, ... each <= 2 MiB
```

Accepted size suffixes: `b`, `kb`, `mb`, `gb` (decimal or integer values).

## Security scanning

By default, GopherCram scans every text file for embedded secrets. When
something is flagged:

- The file is dropped from the rendered output (replaced with a "skipped"
  marker so its existence is still visible).
- The CLI prints a one-line report per finding (path, line, rule).

Turn it off with `--no-security-check`. The bundled rule set covers AWS
keys, GitHub/GitLab tokens, Google API keys, Slack tokens & webhooks,
Stripe keys, JWTs, PEM-encoded private keys, and several database
connection-string shapes.

## Git integration

If your target directory is inside a git work tree, GopherCram can:

- Sort files by change frequency (`--git-sort-by-changes`, on by default).
- Append working-tree and staged diffs (`--include-diffs`).
- Append a recent commit log (`--include-logs --include-logs-count 50`).

These features are no-ops when the directory isn't a git repo or when
`git` isn't on PATH.

## Remote repositories

GopherCram can clone & pack remote repos in one shot:

```bash
gophercram --remote owner/repo
gophercram --remote https://github.com/owner/repo
gophercram --remote owner/repo --remote-branch v1.2.3
gophercram --remote https://github.com/owner/repo/tree/main/sub/dir
```

The `owner/repo` shorthand is expanded to `https://github.com/owner/repo.git`.
The clone happens in a temp directory that's removed when the run
finishes.

By default, any `gophercram.config.json` inside the cloned repo is
**ignored**. Pass `--remote-trust-config` to opt into it (useful when
you control the repo).

## Reports

After every run (unless `--quiet` or `--stdout`), GopherCram prints:

```
Summary:
  files:   <N>
  chars:   <N>
  tokens:  ~<N>
  bytes:   <N> (X.XX MB)

Top 5 files by tokens:
  1. path/to/file — N tokens, M chars
  ...
```

Pass `--top-files-len 20` to expand the list.

Add `--token-count-tree` (optionally with a threshold) to print a flat
per-file token count list after the run:

```bash
gophercram --token-count-tree 100   # only files with >=100 tokens
```

## Reading paths from stdin

For wiring up with other tools (e.g. `fzf`, `git ls-files`, `find`):

```bash
git ls-files | gophercram --stdin --stdout --style markdown
```

Each non-blank, non-`#` line is treated as a file path.
