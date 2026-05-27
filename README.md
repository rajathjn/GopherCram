# GopherCram

A zero-dependency Go CLI that packs a software repository into a single
file you can hand to a language model. Inspired by the workflow popularised
by repomix, but built from the ground up in idiomatic Go.

## Highlights

- **Four output styles** — XML, Markdown, JSON, and plain text, all
  produced by hand-rolled renderers (no template engines).
- **Gitignore-aware traversal** — built-in default patterns, per-directory
  `.gitignore` / `.ignore` / `.gophercramignore` files, plus CLI- and
  config-supplied glob patterns.
- **Embedded secret scanner** — flags AWS keys, GitHub tokens, Stripe keys,
  JWTs, database connection strings, PEM blocks, and more. Flagged files
  are dropped from the output and reported separately.
- **Approximate token counter** — calibrated against GPT-4 tokenizers; on
  the bundled fixture it agrees with repomix's `tiktoken` count to within
  6%, with no 4 MB vocabulary file to ship.
- **Code compression** — best-effort signature-keeping pass for Go, JS/TS,
  Python, and C-family languages so you can fit larger repos in context.
- **Comment stripping** — language-aware lexers for C-style, hash, Lua,
  SQL, and XML-style comments.
- **Git integration** — optional working-tree / staged diffs, recent
  commit log, and change-frequency ordering of packed files.
- **Remote repos** — clone & pack a GitHub URL or `owner/repo` shorthand
  in a single invocation.
- **Split output** — break the packed file into numbered parts on a
  size cap (`--split-output 2mb`).
- **No third-party dependencies** — pure stdlib, builds with `go build`.

## Install

```bash
git clone https://github.com/rajathjn/GopherCram.git
cd GopherCram
go build -o gophercram ./
```

The single static binary is around 6 MB; drop it on your PATH.

## Quick start

```bash
# Pack the current directory (writes gophercram-output.xml)
gophercram

# Pick a different output style and target file
gophercram --style markdown --output context.md src/ docs/

# Clone & pack a remote repo (uses git on PATH)
gophercram --remote owner/repo --remote-branch main

# Pack only TS sources, drop comments + empty lines, dump to stdout
gophercram --include 'src/**/*.ts' \
           --remove-comments --remove-empty-lines \
           --stdout > context.xml

# Write a default config file
gophercram --init
```

## Configuration

GopherCram reads `gophercram.config.json` from the project root by default.
For compatibility, it will also read `repomix.config.json` if no native
config is found. The shape mirrors the schema documented in
[`docs/configuration.md`](docs/configuration.md).

Resolution order (later layers override earlier ones):

1. Built-in defaults
2. Per-user file at `$XDG_CONFIG_HOME/gophercram/gophercram.config.json`
3. Project-root config
4. CLI flags

## Output formats

| Style    | Best for                                                            |
|----------|---------------------------------------------------------------------|
| `xml`    | Default. CDATA-wrapped contents survive special characters.         |
| `markdown` | Human-readable; uses fenced code blocks with language hints.      |
| `json`   | Tool-friendly. Stable shape, easy to post-process.                 |
| `plain`  | ASCII separators; minimal markup; small.                            |

## Tests

```bash
# Unit tests across all packages
go test ./...

# With coverage
go test -cover ./...

# Comparison against npx repomix (requires npx + network)
GOPHERCRAM_INTEGRATION=1 go test ./test/integration/... -v

# Side-by-side bash harness
./test/integration/compare.sh
```

The bundled fixture in [`test/fixtures/dummy-ts-repo`](test/fixtures/dummy-ts-repo)
is a synthetic TypeScript project covering the cases the test suite cares
about: a real `.gitignore`, fake placeholder credentials for the security
scanner, mixed source/test/doc layout, and excluded directories like
`node_modules/`.

## Documentation

- [Usage guide](docs/usage.md)
- [Configuration reference](docs/configuration.md)
- [CLI reference](docs/cli-reference.md)
- [Architecture overview](docs/architecture.md)
- GitHub Pages landing page: [`docs/index.html`](docs/index.html)

## License

MIT. See [LICENSE](LICENSE).
