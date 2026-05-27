package cli

// BuildParser registers every command-line flag GopherCram understands.
//
// The flags are grouped by purpose so the generated help text can render
// them in the order users expect. The parser preserves group ordering by
// registration order.
func BuildParser() *Parser {
	p := NewParser()
	register := func(s FlagSpec) { p.Register(s) }

	// I/O behaviour
	register(FlagSpec{Name: "verbose", Kind: FlagBool, Group: "Logging",
		Description: "Print detailed progress and debug information."})
	register(FlagSpec{Name: "quiet", Kind: FlagBool, Group: "Logging",
		Description: "Suppress non-error console output."})
	register(FlagSpec{Name: "stdout", Kind: FlagBool, Group: "Logging",
		Description: "Write the packed output to stdout (implies --quiet)."})
	register(FlagSpec{Name: "stdin", Kind: FlagBool, Group: "Logging",
		Description: "Read file paths to include from stdin (one per line)."})
	register(FlagSpec{Name: "copy", Kind: FlagBool, Group: "Logging",
		Description: "Copy the packed output to the system clipboard."})
	register(FlagSpec{Name: "top-files-len", Kind: FlagInt, Group: "Logging",
		Description: "Number of largest files (by tokens) to show in the run report."})
	register(FlagSpec{Name: "token-count-tree", Kind: FlagOptionalInt, Group: "Logging",
		Description: "After packing, print a directory tree with per-file token counts; optional threshold filters small files."})

	// Output content
	register(FlagSpec{Name: "output", Short: "o", Kind: FlagString, Group: "Output",
		Description: "Output file path (use '-' for stdout)."})
	register(FlagSpec{Name: "style", Kind: FlagString, Group: "Output",
		Description: "Output style: xml|markdown|json|plain."})
	register(FlagSpec{Name: "parsable-style", Kind: FlagBool, Group: "Output",
		Description: "Escape file content so the output remains valid XML/Markdown."})
	register(FlagSpec{Name: "compress", Kind: FlagBool, Group: "Output",
		Description: "Compress code by keeping signatures and dropping bodies."})
	register(FlagSpec{Name: "output-show-line-numbers", Kind: FlagBool, Group: "Output",
		Description: "Prefix every line in the output with its line number."})
	register(FlagSpec{Name: "file-summary", Kind: FlagBool, Group: "Output",
		Description: "Include the summary section in the output (default true)."})
	register(FlagSpec{Name: "directory-structure", Kind: FlagBool, Group: "Output",
		Description: "Include the directory tree in the output (default true)."})
	register(FlagSpec{Name: "files", Kind: FlagBool, Group: "Output",
		Description: "Include file contents in the output (default true)."})
	register(FlagSpec{Name: "remove-comments", Kind: FlagBool, Group: "Output",
		Description: "Strip comments from supported languages."})
	register(FlagSpec{Name: "remove-empty-lines", Kind: FlagBool, Group: "Output",
		Description: "Drop blank lines from file content."})
	register(FlagSpec{Name: "truncate-base64", Kind: FlagBool, Group: "Output",
		Description: "Replace long base64 blobs with a placeholder."})
	register(FlagSpec{Name: "header-text", Kind: FlagString, Group: "Output",
		Description: "Custom text prepended to the packed output."})
	register(FlagSpec{Name: "instruction-file-path", Kind: FlagString, Group: "Output",
		Description: "Path to an instruction file whose contents are appended to the output."})
	register(FlagSpec{Name: "split-output", Kind: FlagString, Group: "Output",
		Description: "Split the output into multiple numbered files of the given size (e.g. 2mb)."})
	register(FlagSpec{Name: "include-empty-directories", Kind: FlagBool, Group: "Output",
		Description: "Include empty directories in the tree view."})
	register(FlagSpec{Name: "include-full-directory-structure", Kind: FlagBool, Group: "Output",
		Description: "Render the full tree even when --include narrows the file list."})
	register(FlagSpec{Name: "git-sort-by-changes", Kind: FlagBool, Group: "Output",
		Description: "Order files by git change frequency (default true)."})
	register(FlagSpec{Name: "include-diffs", Kind: FlagBool, Group: "Output",
		Description: "Append working-tree and staged diffs."})
	register(FlagSpec{Name: "include-logs", Kind: FlagBool, Group: "Output",
		Description: "Append a recent git log."})
	register(FlagSpec{Name: "include-logs-count", Kind: FlagInt, Group: "Output",
		Description: "Number of commits to include when --include-logs is set."})

	// File selection
	register(FlagSpec{Name: "include", Kind: FlagString, Group: "Selection",
		Description: "Comma-separated glob patterns to include (repeatable)."})
	register(FlagSpec{Name: "ignore", Short: "i", Kind: FlagString, Group: "Selection",
		Description: "Comma-separated glob patterns to ignore (repeatable)."})
	register(FlagSpec{Name: "gitignore", Kind: FlagBool, Group: "Selection",
		Description: "Apply .gitignore rules (default true)."})
	register(FlagSpec{Name: "dot-ignore", Kind: FlagBool, Group: "Selection",
		Description: "Apply .ignore / .gophercramignore files (default true)."})
	register(FlagSpec{Name: "default-patterns", Kind: FlagBool, Group: "Selection",
		Description: "Apply the built-in default ignore patterns (default true)."})
	register(FlagSpec{Name: "max-file-size", Kind: FlagInt64, Group: "Selection",
		Description: "Skip files larger than this many bytes (default 52428800)."})

	// Remote repos
	register(FlagSpec{Name: "remote", Kind: FlagString, Group: "Remote",
		Description: "Clone and pack a remote repository (URL or owner/repo)."})
	register(FlagSpec{Name: "remote-branch", Kind: FlagString, Group: "Remote",
		Description: "Branch, tag, or commit to check out from the remote."})
	register(FlagSpec{Name: "remote-trust-config", Kind: FlagBool, Group: "Remote",
		Description: "Trust gophercram config files inside the cloned repo (off by default)."})

	// Configuration
	register(FlagSpec{Name: "config", Short: "c", Kind: FlagString, Group: "Config",
		Description: "Path to a config file."})
	register(FlagSpec{Name: "init", Kind: FlagBool, Group: "Config",
		Description: "Generate a default config file and exit."})
	register(FlagSpec{Name: "global", Kind: FlagBool, Group: "Config",
		Description: "With --init, write the config to the user-global location."})

	// Security
	register(FlagSpec{Name: "security-check", Kind: FlagBool, Group: "Security",
		Description: "Scan for embedded secrets (default true)."})

	// Token counting
	register(FlagSpec{Name: "token-count-encoding", Kind: FlagString, Group: "Tokens",
		Description: "Token-count encoding (currently 'approx' only)."})

	// Meta
	register(FlagSpec{Name: "version", Short: "v", Kind: FlagBool, Group: "Meta",
		Description: "Print the GopherCram version and exit."})
	register(FlagSpec{Name: "help", Short: "h", Kind: FlagBool, Group: "Meta",
		Description: "Print this help text and exit."})

	return p
}
