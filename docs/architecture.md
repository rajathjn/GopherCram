# Architecture

GopherCram is a small, self-contained Go program. This page sketches the
shape of the codebase so contributors know where to look first.

## Package layout

```
GopherCram/
├── cmd/
│   └── gophercram/              # process entry point (package main)
│       └── main.go
├── internal/
│   ├── cli/                     # arg parsing, run loop, help, init action
│   ├── config/                  # config schema, defaults, loader, merge
│   ├── ignore/                  # gitignore-style glob matcher & rule set
│   ├── fileselect/              # walker, content reader, tree builder
│   ├── manipulate/              # comment stripper, base64 truncator, compressor
│   ├── security/                # secret-detection rules and scanner
│   ├── metrics/                 # rune / line / token estimator + aggregator
│   ├── output/                  # XML / Markdown / JSON / Plain renderers
│   ├── gitops/                  # `git` subprocess wrappers and URL parsing
│   └── pack/                    # orchestration: pipeline glue + disk writer
├── docs/                        # documentation (you are here)
├── scripts/                     # developer scripts (e.g. compare.sh)
└── test/
    ├── fixtures/dummy-ts-repo/  # synthetic TypeScript fixture
    └── integration/             # compare against npx repomix
```

## Data flow

```
                    ┌──────────────────────┐
       argv ───────►│   cli.Parse / overlay│───┐
                    └──────────────────────┘   │
                                               ▼
defaults  ──►  global cfg  ──►  project cfg ──►  CLI overlay
                                               │
                                               ▼
                                     ┌─────────────────┐
                                     │  config.Config  │
                                     └────────┬────────┘
                                              │
                                              ▼
                              ┌───────────────────────────┐
                              │  pack.Packager.Pack(roots)│
                              └────────────┬──────────────┘
                                           │
       ┌──────────────────────┬────────────┼─────────────┬─────────────────┐
       ▼                      ▼            ▼             ▼                 ▼
fileselect.Walk    fileselect.ReadAll  security.Scan  manipulate.Apply  metrics.Compute
       │                      │            │             │                 │
       └────► tree ◄──────────┘            │             │                 │
                                           └───┬─────────┘                 │
                                               ▼                           │
                                       output.For(style)  ◄────────────────┘
                                               │
                                               ▼
                                          rendered text
                                               │
                                               ▼
                                  pack.WriteToDisk  /  stdout  /  clipboard
```

## Configuration merging

`config.Config` is the resolved form used at runtime. `config.PartialConfig`
mirrors it with pointer-typed fields so we can distinguish "unset" from
"set to false". `config.Merge(base, overlay)` layers a partial onto a
resolved config without ever silently demoting a boolean.

The merge order is fixed: defaults → global file → project file → CLI.
CLI flags always win because they are the most explicit user intent.

## Walker

The walker (`fileselect.Walk`) lazily loads per-directory ignore files as
it descends, so a nested `.gitignore` can re-include or exclude files in
its subtree. Rules from outer directories continue to apply, and the
last-matching rule wins (matching gitignore's semantics).

Default ignore patterns and the binary-extension lookup table both live
in `internal/config` so tests can exercise them without spinning up a
walker.

## Output rendering

All renderers conform to the `output.Renderer` interface:

```go
type Renderer interface {
    Render(doc *Document) (string, error)
    FileExtension() string
}
```

The XML renderer is hand-rolled (with CDATA wrapping that escapes the
forbidden `]]>` terminator), the JSON renderer leans on
`encoding/json`, and the Markdown/Plain renderers are simple
`strings.Builder` walks. The shared `output.Document` type carries
everything renderers need — including pre-computed metrics and the
directory tree — so the rendering layer is stateless.

## Security scanning

`security.Scanner` runs a curated set of regexes, each tagged with a
human-readable rule name and an optional minimum entropy threshold. The
entropy filter trims obvious placeholder matches (e.g. literally
`password = "aaaaaaaa"`).

When a file produces any finding, the packager flags it as
`Skipped: true` with a "removed by security scanner" reason. Renderers
treat skipped files like binary files: their path stays visible, but
the content does not.

## Token estimation

`metrics.EstimateTokens` is a tiered approximator:

- For ASCII-heavy text we split on word boundaries with corrections for
  digit runs and camel/snake case transitions, which match how BPE
  tokenizers segment identifiers.
- For text dominated by non-ASCII codepoints (CJK, emoji), we fall back
  to roughly one token per rune.

Calibration against GPT-4's `o200k_base` over the bundled fixture lands
within a few percent without shipping a vocabulary file.

## Testing strategy

- **Unit tests** live next to the code in `_test.go` files. Coverage
  hovers around 80–95% per package.
- **Integration tests** in `test/integration` run GopherCram against the
  bundled fixture and assert on its output. A second set, gated on
  `GOPHERCRAM_INTEGRATION=1`, additionally invokes `npx repomix` and
  cross-checks the file list and token total.
- **Comparison harness** (`scripts/compare.sh`) is a bash
  script that side-by-side packs the fixture with both tools across
  every output style, printing a summary of file inclusion/exclusion
  and byte counts.

## Adding a new output style

1. Implement the `output.Renderer` interface in a new file under
   `internal/output/`.
2. Register the style in `output.For` and `config.OutputStyle`.
3. Add a default filename to `config.DefaultFilenameFor`.
4. Update the CLI `--style` description.
5. Add a smoke-test renderer case in `internal/output/output_test.go`.
