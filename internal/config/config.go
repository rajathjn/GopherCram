// Package config defines the configuration schema for GopherCram and provides
// helpers for loading, merging, and validating configuration data from files,
// CLI flags, and built-in defaults.
package config

import (
	"errors"
	"fmt"
	"strings"
)

// OutputStyle identifies the format of the generated packed file.
type OutputStyle string

const (
	StyleXML      OutputStyle = "xml"
	StyleMarkdown OutputStyle = "markdown"
	StyleJSON     OutputStyle = "json"
	StylePlain    OutputStyle = "plain"
)

// AllStyles returns every recognised output style.
func AllStyles() []OutputStyle {
	return []OutputStyle{StyleXML, StyleMarkdown, StyleJSON, StylePlain}
}

// Valid reports whether the style is one of the recognised values.
func (s OutputStyle) Valid() bool {
	switch s {
	case StyleXML, StyleMarkdown, StyleJSON, StylePlain:
		return true
	}
	return false
}

// DefaultFilenameFor returns the conventional output filename for a style.
func DefaultFilenameFor(style OutputStyle) string {
	switch style {
	case StyleMarkdown:
		return "gophercram-output.md"
	case StyleJSON:
		return "gophercram-output.json"
	case StylePlain:
		return "gophercram-output.txt"
	default:
		return "gophercram-output.xml"
	}
}

// Input describes input-related configuration.
type Input struct {
	// MaxFileSize is the largest file (bytes) we will read.
	MaxFileSize int64 `json:"maxFileSize,omitempty"`
}

// GitOptions configures git-aware behaviour.
type GitOptions struct {
	SortByChanges           bool `json:"sortByChanges"`
	SortByChangesMaxCommits int  `json:"sortByChangesMaxCommits,omitempty"`
	IncludeDiffs            bool `json:"includeDiffs"`
	IncludeLogs             bool `json:"includeLogs"`
	IncludeLogsCount        int  `json:"includeLogsCount,omitempty"`
}

// Output controls how the packed file is rendered.
type Output struct {
	FilePath                      string      `json:"filePath,omitempty"`
	Style                         OutputStyle `json:"style,omitempty"`
	ParsableStyle                 bool        `json:"parsableStyle,omitempty"`
	HeaderText                    string      `json:"headerText,omitempty"`
	InstructionFilePath           string      `json:"instructionFilePath,omitempty"`
	FileSummary                   bool        `json:"fileSummary"`
	DirectoryStructure            bool        `json:"directoryStructure"`
	Files                         bool        `json:"files"`
	RemoveComments                bool        `json:"removeComments,omitempty"`
	RemoveEmptyLines              bool        `json:"removeEmptyLines,omitempty"`
	Compress                      bool        `json:"compress,omitempty"`
	TopFilesLength                int         `json:"topFilesLength,omitempty"`
	ShowLineNumbers               bool        `json:"showLineNumbers,omitempty"`
	TruncateBase64                bool        `json:"truncateBase64,omitempty"`
	CopyToClipboard               bool        `json:"copyToClipboard,omitempty"`
	IncludeEmptyDirectories       bool        `json:"includeEmptyDirectories,omitempty"`
	IncludeFullDirectoryStructure bool        `json:"includeFullDirectoryStructure,omitempty"`
	SplitOutputBytes              int64       `json:"splitOutput,omitempty"`
	TokenCountTree                bool        `json:"-"`
	TokenCountTreeThreshold       int         `json:"-"`
	Stdout                        bool        `json:"-"`
	Git                           GitOptions  `json:"git"`
}

// Ignore controls which paths are filtered out of the packed file.
type Ignore struct {
	UseGitignore      bool     `json:"useGitignore"`
	UseDotIgnore      bool     `json:"useDotIgnore"`
	UseDefaultPatterns bool    `json:"useDefaultPatterns"`
	CustomPatterns    []string `json:"customPatterns,omitempty"`
}

// Security controls secret scanning behaviour.
type Security struct {
	EnableSecurityCheck bool `json:"enableSecurityCheck"`
}

// TokenCount controls token counting behaviour.
type TokenCount struct {
	// Encoding selects the approximate token model. Currently only "approx"
	// is supported, which uses a heuristic byte-pair-style estimator that
	// produces results within ~10% of GPT-4 tokenizers without external deps.
	Encoding string `json:"encoding,omitempty"`
}

// Config is the merged, fully-resolved configuration.
type Config struct {
	Schema     string     `json:"$schema,omitempty"`
	Input      Input      `json:"input"`
	Output     Output     `json:"output"`
	Include    []string   `json:"include,omitempty"`
	Ignore     Ignore     `json:"ignore"`
	Security   Security   `json:"security"`
	TokenCount TokenCount `json:"tokenCount"`

	// Working directory in which the run was launched.
	Cwd string `json:"-"`
}

// Defaults returns a Config populated with sensible default values.
func Defaults() Config {
	return Config{
		Input: Input{MaxFileSize: 50 * 1024 * 1024},
		Output: Output{
			FilePath:           DefaultFilenameFor(StyleXML),
			Style:              StyleXML,
			FileSummary:        true,
			DirectoryStructure: true,
			Files:              true,
			TopFilesLength:     5,
			Git: GitOptions{
				SortByChanges:           true,
				SortByChangesMaxCommits: 100,
				IncludeLogsCount:        50,
			},
		},
		Ignore: Ignore{
			UseGitignore:       true,
			UseDotIgnore:       true,
			UseDefaultPatterns: true,
		},
		Security:   Security{EnableSecurityCheck: true},
		TokenCount: TokenCount{Encoding: "approx"},
	}
}

// Validate returns an error if any required fields are missing or invalid.
func (c *Config) Validate() error {
	if !c.Output.Style.Valid() {
		return fmt.Errorf("invalid output style %q: must be one of xml|markdown|json|plain", c.Output.Style)
	}
	if c.Input.MaxFileSize < 0 {
		return errors.New("input.maxFileSize must be non-negative")
	}
	if c.Output.TopFilesLength < 0 {
		return errors.New("output.topFilesLength must be non-negative")
	}
	if c.Output.Git.IncludeLogsCount < 0 {
		return errors.New("output.git.includeLogsCount must be non-negative")
	}
	if c.Output.SplitOutputBytes < 0 {
		return errors.New("output.splitOutput must be non-negative")
	}
	for _, p := range c.Include {
		if strings.TrimSpace(p) == "" {
			return errors.New("include patterns must be non-empty strings")
		}
	}
	return nil
}
