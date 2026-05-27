// Package pack orchestrates the end-to-end flow: file selection → content
// loading → transformations → security scan → metrics → rendering.
//
// The Packager type exists so callers (CLI, MCP server, library users) can
// configure dependencies and override individual phases for testing.
package pack

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rajathjn/GopherCram/internal/config"
	"github.com/rajathjn/GopherCram/internal/fileselect"
	"github.com/rajathjn/GopherCram/internal/gitops"
	"github.com/rajathjn/GopherCram/internal/manipulate"
	"github.com/rajathjn/GopherCram/internal/metrics"
	"github.com/rajathjn/GopherCram/internal/output"
	"github.com/rajathjn/GopherCram/internal/security"
)

// Result is what Pack returns to its caller.
type Result struct {
	Output         string
	Parts          []string // Non-empty when output is split.
	Metrics        metrics.Aggregate
	Findings       []security.Finding
	SuspiciousPath map[string]struct{}
	Files          []fileselect.File
	Renderer       output.Renderer
}

// Packager glues the pieces together.
type Packager struct {
	AppName    string
	AppVersion string

	// Optional overrides — left nil they use sensible defaults.
	Walker     func(opts fileselect.Options) ([]fileselect.Result, error)
	Reader     func([]fileselect.Result) ([]fileselect.File, error)
	GitContext context.Context
}

// New returns a Packager with default dependencies.
func New(appName, appVersion string) *Packager {
	return &Packager{AppName: appName, AppVersion: appVersion}
}

// Pack runs the full pipeline against `cfg` and `roots`.
func (p *Packager) Pack(cfg *config.Config, roots []string) (*Result, error) {
	if cfg == nil {
		return nil, errors.New("pack: nil config")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	walk := p.Walker
	if walk == nil {
		walk = fileselect.Walk
	}
	read := p.Reader
	if read == nil {
		read = fileselect.ReadAll
	}

	selectedResults, err := walk(fileselect.FromConfig(cfg, roots))
	if err != nil {
		return nil, fmt.Errorf("collect files: %w", err)
	}
	files, err := read(selectedResults)
	if err != nil {
		return nil, fmt.Errorf("read files: %w", err)
	}

	// Security scan before content transforms, so we report on the original
	// bytes the user has on disk.
	var findings []security.Finding
	suspicious := map[string]struct{}{}
	if cfg.Security.EnableSecurityCheck {
		sc := security.New()
		for _, f := range files {
			if f.Skipped {
				continue
			}
			fs := sc.Scan(f.RelPath, f.Content)
			if len(fs) > 0 {
				findings = append(findings, fs...)
				suspicious[f.RelPath] = struct{}{}
			}
		}
	}

	// Drop suspicious files from the rendered output (their findings are
	// surfaced in the CLI report instead).
	cleaned := make([]fileselect.File, 0, len(files))
	for _, f := range files {
		if _, bad := suspicious[f.RelPath]; bad {
			f.Skipped = true
			f.SkipReason = "removed by security scanner"
		}
		cleaned = append(cleaned, f)
	}
	files = cleaned

	// Optional git-aware sort by change frequency.
	if cfg.Output.Git.SortByChanges && len(roots) > 0 {
		if freq, err := gitops.ChangeFreq(p.GitContext, roots[0], cfg.Output.Git.SortByChangesMaxCommits); err == nil && len(freq) > 0 {
			sort.SliceStable(files, func(i, j int) bool {
				return freq[files[i].RelPath] > freq[files[j].RelPath]
			})
		}
	}

	// Apply content transforms.
	for i := range files {
		if files[i].Skipped {
			continue
		}
		files[i].Content = manipulate.Apply(files[i].RelPath, files[i].Content, cfg.Output)
	}

	// Build the metrics from the post-transform content.
	var items []struct {
		Path    string
		Content string
		Bytes   int64
	}
	for _, f := range files {
		if f.Skipped {
			continue
		}
		items = append(items, struct {
			Path    string
			Content string
			Bytes   int64
		}{Path: f.RelPath, Content: f.Content, Bytes: f.Size})
	}
	agg := metrics.Compute(items)

	// Directory tree (uses the full set of selected paths or only included).
	treePaths := paths(files)
	if cfg.Output.IncludeFullDirectoryStructure {
		// Include even files dropped by the include filter — rerun the
		// walker with an empty include list.
		opts := fileselect.FromConfig(cfg, roots)
		opts.Include = nil
		if all, err := walk(opts); err == nil {
			treePaths = nil
			for _, r := range all {
				treePaths = append(treePaths, r.RelPath)
			}
		}
	}
	tree := fileselect.RenderTree(fileselect.BuildTree(treePaths))

	// Header and instruction content.
	headerText := cfg.Output.HeaderText
	instruction := ""
	if cfg.Output.InstructionFilePath != "" {
		if b, err := os.ReadFile(cfg.Output.InstructionFilePath); err == nil {
			instruction = string(b)
		}
	}

	// Optional git diffs and logs.
	var diffWT, diffStg string
	var gitLog []output.GitCommit
	if len(roots) > 0 {
		root := roots[0]
		if cfg.Output.Git.IncludeDiffs && gitops.IsRepo(root) {
			diffWT, _ = gitops.DiffWorkTree(p.GitContext, root)
			diffStg, _ = gitops.DiffStaged(p.GitContext, root)
		}
		if cfg.Output.Git.IncludeLogs && gitops.IsRepo(root) {
			if logs, err := gitops.Log(p.GitContext, root, cfg.Output.Git.IncludeLogsCount); err == nil {
				for _, c := range logs {
					gitLog = append(gitLog, output.GitCommit{
						Hash: c.Hash, Author: c.Author, Date: c.Date,
						Message: c.Message, Files: c.Files,
					})
				}
			}
		}
	}

	doc := &output.Document{
		GeneratedAt:     time.Now(),
		Config:          cfg,
		HeaderText:      headerText,
		Instruction:     instruction,
		DirectoryTree:   tree,
		Files:           files,
		Aggregate:       agg,
		GitDiffWorkTree: diffWT,
		GitDiffStaged:   diffStg,
		GitLog:          gitLog,
		AppName:         p.AppName,
		AppVersion:      p.AppVersion,
	}

	renderer := output.For(cfg.Output.Style)
	rendered, err := renderer.Render(doc)
	if err != nil {
		return nil, err
	}

	res := &Result{
		Output:         rendered,
		Metrics:        agg,
		Findings:       findings,
		SuspiciousPath: suspicious,
		Files:          files,
		Renderer:       renderer,
	}
	if cfg.Output.SplitOutputBytes > 0 {
		res.Parts = output.Split(rendered, cfg.Output.SplitOutputBytes)
	}
	return res, nil
}

// WriteToDisk writes the rendered output (or split parts) to the configured
// destination, returning the list of files actually written.
func WriteToDisk(cfg *config.Config, res *Result) ([]string, error) {
	outPath := cfg.Output.FilePath
	if outPath == "" {
		outPath = config.DefaultFilenameFor(cfg.Output.Style)
	}
	if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(cfg.Cwd, outPath)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, err
	}

	if len(res.Parts) > 1 {
		var written []string
		for i, part := range res.Parts {
			p := output.PartFilename(outPath, i+1)
			if err := os.WriteFile(p, []byte(part), 0o644); err != nil {
				return written, err
			}
			written = append(written, p)
		}
		return written, nil
	}
	if err := os.WriteFile(outPath, []byte(res.Output), 0o644); err != nil {
		return nil, err
	}
	return []string{outPath}, nil
}

func paths(files []fileselect.File) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.RelPath)
	}
	sort.Strings(out)
	return out
}

// HumanSize returns "1.2 MB" / "350 KB" style strings.
func HumanSize(n int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case n >= GB:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(KB))
	}
	return fmt.Sprintf("%d B", n)
}

// ReportTopFiles returns a one-per-line summary of the top files by tokens.
func ReportTopFiles(agg metrics.Aggregate, n int) string {
	if n <= 0 || agg.TotalFiles == 0 {
		return ""
	}
	top := agg.Top(n)
	var b strings.Builder
	for i, f := range top {
		b.WriteString(fmt.Sprintf("  %d. %s — %d tokens, %d chars\n", i+1, f.Path, f.Tokens, f.Chars))
	}
	return b.String()
}
