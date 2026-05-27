package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rajathjn/GopherCram/internal/config"
)

// InitOptions controls behaviour of the `--init` action.
type InitOptions struct {
	Cwd    string
	Global bool
	Out    io.Writer
}

// RunInit writes a default config file to either the project root or the
// user-global location, depending on opts.Global.
func RunInit(opts InitOptions) error {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	target := filepath.Join(opts.Cwd, config.ConfigFilename)
	if opts.Global {
		gp := config.GlobalConfigPath()
		if gp == "" {
			return errors.New("could not determine global config path")
		}
		target = gp
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("config file already exists at %s", target)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	body := defaultConfigJSON()
	if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(opts.Out, "wrote default config to %s\n", target)
	return nil
}

// defaultConfigJSON produces the JSON body for `gophercram --init`.
func defaultConfigJSON() string {
	cfg := config.Defaults()
	type fileShape struct {
		Schema     string             `json:"$schema"`
		Input      config.Input       `json:"input"`
		Output     marshallableOutput `json:"output"`
		Include    []string           `json:"include"`
		Ignore     config.Ignore      `json:"ignore"`
		Security   config.Security    `json:"security"`
		TokenCount config.TokenCount  `json:"tokenCount"`
	}
	shape := fileShape{
		Schema:     "https://gophercram.dev/schema.json",
		Input:      cfg.Input,
		Include:    []string{},
		Ignore:     cfg.Ignore,
		Security:   cfg.Security,
		TokenCount: cfg.TokenCount,
		Output: marshallableOutput{
			FilePath:           cfg.Output.FilePath,
			Style:              string(cfg.Output.Style),
			FileSummary:        cfg.Output.FileSummary,
			DirectoryStructure: cfg.Output.DirectoryStructure,
			Files:              cfg.Output.Files,
			TopFilesLength:     cfg.Output.TopFilesLength,
			Git: marshallableGit{
				SortByChanges:           cfg.Output.Git.SortByChanges,
				SortByChangesMaxCommits: cfg.Output.Git.SortByChangesMaxCommits,
				IncludeLogsCount:        cfg.Output.Git.IncludeLogsCount,
			},
		},
	}
	if shape.Ignore.CustomPatterns == nil {
		shape.Ignore.CustomPatterns = []string{}
	}
	body, _ := json.MarshalIndent(shape, "", "  ")
	return string(body) + "\n"
}

// marshallableOutput and marshallableGit are tightly-typed JSON shapes used
// only for `--init`; we omit transient runtime-only fields like Stdout.
type marshallableOutput struct {
	FilePath           string          `json:"filePath"`
	Style              string          `json:"style"`
	FileSummary        bool            `json:"fileSummary"`
	DirectoryStructure bool            `json:"directoryStructure"`
	Files              bool            `json:"files"`
	TopFilesLength     int             `json:"topFilesLength"`
	Git                marshallableGit `json:"git"`
}

type marshallableGit struct {
	SortByChanges           bool `json:"sortByChanges"`
	SortByChangesMaxCommits int  `json:"sortByChangesMaxCommits"`
	IncludeLogsCount        int  `json:"includeLogsCount"`
}
