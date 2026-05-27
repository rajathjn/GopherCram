package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rajathjn/GopherCram/internal/config"
	"github.com/rajathjn/GopherCram/internal/gitops"
	"github.com/rajathjn/GopherCram/internal/pack"
)

// RunOptions configures the CLI Run function and is mostly populated from
// process state (os.Args, os.Stdout, etc.). Tests construct it directly.
type RunOptions struct {
	Argv   []string  // does NOT include argv[0]
	Cwd    string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// ExitCode is the conventional process exit code returned by Run.
type ExitCode int

const (
	ExitOK     ExitCode = 0
	ExitError  ExitCode = 1
	ExitUsage  ExitCode = 2
)

// Run parses arguments, loads configuration, and executes the requested
// action. The returned exit code mirrors POSIX conventions.
func Run(opts RunOptions) ExitCode {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.Cwd == "" {
		opts.Cwd, _ = os.Getwd()
	}

	parser := BuildParser()
	args, err := parser.Parse(opts.Argv)
	if err != nil {
		fmt.Fprintln(opts.Stderr, "error:", err)
		fmt.Fprintln(opts.Stderr, "run 'gophercram --help' for usage.")
		return ExitUsage
	}

	if args.HasFlag("help") {
		fmt.Fprint(opts.Stdout, HelpText(parser))
		return ExitOK
	}
	if args.HasFlag("version") {
		fmt.Fprintf(opts.Stdout, "%s v%s\n", AppName, Version)
		return ExitOK
	}
	if args.HasFlag("init") {
		err := RunInit(InitOptions{Cwd: opts.Cwd, Global: args.HasFlag("global"), Out: opts.Stdout})
		if err != nil {
			fmt.Fprintln(opts.Stderr, "error:", err)
			return ExitError
		}
		return ExitOK
	}

	cfg, err := buildConfig(opts, args)
	if err != nil {
		fmt.Fprintln(opts.Stderr, "error:", err)
		return ExitError
	}

	// Resolve roots. Positional args take priority; if none, use ".".
	roots := args.Positional
	if args.HasFlag("stdin") {
		readers, err := readStdinPaths(opts.Stdin)
		if err != nil {
			fmt.Fprintln(opts.Stderr, "error reading stdin:", err)
			return ExitError
		}
		roots = readers
	}
	if len(roots) == 0 {
		roots = []string{"."}
	}

	// Detect bare remote URL in positional args.
	if len(roots) == 1 && gitops.IsExplicitRemoteURL(roots[0]) && !args.HasFlag("remote") {
		fv := &FlagValue{Set: true, Str: roots[0]}
		args.Flags["remote"] = fv
		roots = []string{"."}
	}

	if args.HasFlag("remote") {
		return runRemote(opts, args, cfg)
	}

	return runLocal(opts, args, cfg, roots)
}

func buildConfig(opts RunOptions, args *ParsedArgs) (*config.Config, error) {
	cfg := config.Defaults()
	cfg.Cwd = opts.Cwd

	// Layer global config (if any).
	if gp := config.GlobalConfigPath(); gp != "" {
		if p, err := config.LoadPartialFromFile(gp); err == nil && p != nil {
			cfg = config.Merge(cfg, *p)
		}
	}
	// Layer project config: explicit --config beats discovery.
	configPath := ""
	if args.HasFlag("config") {
		configPath = args.Get("config").Str
	} else {
		configPath = config.Discover(opts.Cwd)
	}
	if configPath != "" {
		p, err := config.LoadPartialFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("load config %s: %w", configPath, err)
		}
		if p != nil {
			cfg = config.Merge(cfg, *p)
		}
	}
	// Layer CLI overlay last.
	overlay, err := OverlayFromArgs(args)
	if err != nil {
		return nil, err
	}
	cfg = config.Merge(cfg, *overlay)

	// Force stdout mode when --output=-
	if cfg.Output.FilePath == "-" {
		cfg.Output.Stdout = true
	}
	if args.HasFlag("stdout") {
		cfg.Output.Stdout = true
	}

	// Apply the default filename when the user changed --style but not
	// --output, so users don't end up with `repo.xml` containing markdown.
	if !args.HasFlag("output") && cfg.Output.Style != config.StyleXML {
		cfg.Output.FilePath = config.DefaultFilenameFor(cfg.Output.Style)
	}
	return &cfg, nil
}

func readStdinPaths(r io.Reader) ([]string, error) {
	var out []string
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("no paths read from stdin")
	}
	return out, nil
}

// runLocal executes the default action on a local file tree.
func runLocal(opts RunOptions, args *ParsedArgs, cfg *config.Config, roots []string) ExitCode {
	if !cfg.Output.Stdout && !args.HasFlag("quiet") {
		fmt.Fprintf(opts.Stderr, "%s v%s — packing %d root(s)\n", AppName, Version, len(roots))
	}
	pk := pack.New(AppName, Version)
	pk.GitContext = context.Background()
	res, err := pk.Pack(cfg, roots)
	if err != nil {
		fmt.Fprintln(opts.Stderr, "error:", err)
		return ExitError
	}

	// Write or print output.
	if cfg.Output.Stdout {
		fmt.Fprint(opts.Stdout, res.Output)
	} else {
		written, err := pack.WriteToDisk(cfg, res)
		if err != nil {
			fmt.Fprintln(opts.Stderr, "error writing output:", err)
			return ExitError
		}
		if !args.HasFlag("quiet") {
			for _, p := range written {
				fmt.Fprintf(opts.Stderr, "wrote %s\n", p)
			}
		}
	}

	// Optional clipboard
	if cfg.Output.CopyToClipboard {
		if tool, err := CopyToClipboard(res.Output); err != nil {
			fmt.Fprintf(opts.Stderr, "clipboard copy failed: %v\n", err)
		} else if !args.HasFlag("quiet") {
			fmt.Fprintf(opts.Stderr, "copied to clipboard via %s\n", tool)
		}
	}

	// Report
	if !args.HasFlag("quiet") && !cfg.Output.Stdout {
		writeReport(opts.Stderr, cfg, res)
	}

	// Token-count tree
	if args.HasFlag("token-count-tree") {
		threshold := args.Get("token-count-tree").Int
		writeTokenTree(opts.Stdout, res, threshold)
	}

	return ExitOK
}

func writeReport(w io.Writer, cfg *config.Config, res *pack.Result) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary:")
	fmt.Fprintf(w, "  files:   %d\n", res.Metrics.TotalFiles)
	fmt.Fprintf(w, "  chars:   %d\n", res.Metrics.TotalChars)
	fmt.Fprintf(w, "  tokens:  ~%d\n", res.Metrics.TotalTokens)
	fmt.Fprintf(w, "  bytes:   %d (%s)\n", res.Metrics.TotalBytes, pack.HumanSize(res.Metrics.TotalBytes))
	n := cfg.Output.TopFilesLength
	if n <= 0 {
		n = 5
	}
	if res.Metrics.TotalFiles > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Top %d files by tokens:\n", n)
		fmt.Fprint(w, pack.ReportTopFiles(res.Metrics, n))
	}
	if len(res.Findings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Security: %d finding(s) — listed below; matching files were excluded from output.\n", len(res.Findings))
		grouped := map[string][]string{}
		for _, f := range res.Findings {
			grouped[f.Path] = append(grouped[f.Path], fmt.Sprintf("line %d: %s", f.Line, f.Rule))
		}
		paths := make([]string, 0, len(grouped))
		for p := range grouped {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		for _, p := range paths {
			fmt.Fprintf(w, "  %s\n", p)
			for _, line := range grouped[p] {
				fmt.Fprintf(w, "    %s\n", line)
			}
		}
	}
}

func writeTokenTree(w io.Writer, res *pack.Result, threshold int) {
	if res.Metrics.TotalFiles == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Token-count tree:")
	tokens := map[string]int{}
	for _, f := range res.Metrics.Files {
		tokens[f.Path] = f.Tokens
	}
	// We render a simple flat list sorted by path with tokens to the right.
	paths := make([]string, 0, len(tokens))
	for p := range tokens {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		if tokens[p] < threshold {
			continue
		}
		fmt.Fprintf(w, "  %s\t%d tokens\n", p, tokens[p])
	}
}

// runRemote clones a remote repository into a temporary directory, applies
// the same configuration, and then writes the output to the user's cwd.
func runRemote(opts RunOptions, args *ParsedArgs, cfg *config.Config) ExitCode {
	remote := args.Get("remote").Str
	spec, err := gitops.ParseRemote(remote)
	if err != nil {
		fmt.Fprintln(opts.Stderr, "error: invalid remote", remote)
		return ExitError
	}
	branch := ""
	if args.HasFlag("remote-branch") {
		branch = args.Get("remote-branch").Str
	}
	if spec.Branch != "" && branch == "" {
		branch = spec.Branch
	}

	tmp, err := os.MkdirTemp("", "gophercram-remote-")
	if err != nil {
		fmt.Fprintln(opts.Stderr, "error: tmp dir:", err)
		return ExitError
	}
	defer os.RemoveAll(tmp)

	dest := filepath.Join(tmp, "repo")
	if !args.HasFlag("quiet") {
		fmt.Fprintf(opts.Stderr, "cloning %s ... ", spec.URL)
	}
	if err := gitops.Clone(context.Background(), spec.URL, dest, branch); err != nil {
		if !args.HasFlag("quiet") {
			fmt.Fprintln(opts.Stderr, "failed")
		}
		fmt.Fprintln(opts.Stderr, "error:", err)
		return ExitError
	}
	if !args.HasFlag("quiet") {
		fmt.Fprintln(opts.Stderr, "done")
	}

	// Roots become the cloned directory (or its subdir if specified).
	root := dest
	if spec.Subdir != "" {
		root = filepath.Join(dest, spec.Subdir)
	}

	// If --remote-trust-config is set, look for a gophercram config inside
	// the clone and merge it in.
	if args.HasFlag("remote-trust-config") {
		if found := config.Discover(root); found != "" {
			if p, err := config.LoadPartialFromFile(found); err == nil && p != nil {
				merged := config.Merge(*cfg, *p)
				cfg = &merged
			}
		}
	}

	cfg.Cwd = opts.Cwd
	return runLocal(opts, args, cfg, []string{root})
}
