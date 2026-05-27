package cli

import (
	"strings"

	"github.com/rajathjn/GopherCram/internal/config"
)

// OverlayFromArgs converts parsed CLI flags into a PartialConfig that can be
// merged onto file/defaults.
func OverlayFromArgs(args *ParsedArgs) (*config.PartialConfig, error) {
	if args == nil {
		return &config.PartialConfig{}, nil
	}
	out := &config.PartialConfig{}
	o := &config.PartialOutput{}
	in := &config.PartialInput{}
	ig := &config.PartialIgnore{}
	sec := &config.PartialSecurity{}
	tc := &config.PartialTokenCount{}
	g := &config.PartialGit{}

	setStr := func(name string, dst **string) {
		if args.HasFlag(name) {
			v := args.Get(name).Str
			*dst = &v
		}
	}
	setBool := func(name string, dst **bool) {
		if args.HasFlag(name) {
			v := args.Get(name).Bool
			*dst = &v
		}
	}
	setInt := func(name string, dst **int) {
		if args.HasFlag(name) {
			v := args.Get(name).Int
			*dst = &v
		}
	}
	setInt64 := func(name string, dst **int64) {
		if args.HasFlag(name) {
			v := args.Get(name).Int64
			*dst = &v
		}
	}

	if args.HasFlag("output") {
		v := args.Get("output").Str
		o.FilePath = &v
	}
	if args.HasFlag("style") {
		s := config.OutputStyle(strings.ToLower(args.Get("style").Str))
		o.Style = &s
	}
	setBool("parsable-style", &o.ParsableStyle)
	setBool("compress", &o.Compress)
	setBool("output-show-line-numbers", &o.ShowLineNumbers)
	setBool("file-summary", &o.FileSummary)
	setBool("directory-structure", &o.DirectoryStructure)
	setBool("files", &o.Files)
	setBool("remove-comments", &o.RemoveComments)
	setBool("remove-empty-lines", &o.RemoveEmptyLines)
	setBool("truncate-base64", &o.TruncateBase64)
	setBool("copy", &o.CopyToClipboard)
	setBool("include-empty-directories", &o.IncludeEmptyDirectories)
	setBool("include-full-directory-structure", &o.IncludeFullDirectoryStructure)
	setStr("header-text", &o.HeaderText)
	setStr("instruction-file-path", &o.InstructionFilePath)
	setInt("top-files-len", &o.TopFilesLength)

	if args.HasFlag("split-output") {
		n, err := HumanBytesParse(args.Get("split-output").Str)
		if err != nil {
			return nil, err
		}
		o.SplitOutputBytes = &n
	}

	setBool("git-sort-by-changes", &g.SortByChanges)
	setBool("include-diffs", &g.IncludeDiffs)
	setBool("include-logs", &g.IncludeLogs)
	setInt("include-logs-count", &g.IncludeLogsCount)
	o.Git = g

	if args.HasFlag("include") {
		patterns := splitMultiCSV(args.Get("include").StrList)
		out.Include = &patterns
	}
	if args.HasFlag("ignore") {
		patterns := splitMultiCSV(args.Get("ignore").StrList)
		ig.CustomPatterns = &patterns
	}
	setBool("gitignore", &ig.UseGitignore)
	setBool("dot-ignore", &ig.UseDotIgnore)
	setBool("default-patterns", &ig.UseDefaultPatterns)

	setInt64("max-file-size", &in.MaxFileSize)
	setBool("security-check", &sec.EnableSecurityCheck)
	setStr("token-count-encoding", &tc.Encoding)

	out.Output = o
	out.Input = in
	out.Ignore = ig
	out.Security = sec
	out.TokenCount = tc

	return out, nil
}

func splitMultiCSV(values []string) []string {
	var out []string
	for _, v := range values {
		for _, p := range strings.Split(v, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	return out
}
