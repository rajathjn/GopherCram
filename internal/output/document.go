// Package output renders the packed payload in one of the supported styles
// (xml, markdown, json, plain).
//
// All renderers share a common Document type so the caller only has to gather
// the inputs once.
package output

import (
	"time"

	"github.com/rajathjn/GopherCram/internal/config"
	"github.com/rajathjn/GopherCram/internal/fileselect"
	"github.com/rajathjn/GopherCram/internal/metrics"
)

// Document carries everything the renderers need.
type Document struct {
	// GeneratedAt is the moment the run started.
	GeneratedAt time.Time

	// Config is the merged configuration used for this run.
	Config *config.Config

	// HeaderText is the optional user-supplied header.
	HeaderText string

	// Instruction is the contents of the file referenced by
	// Config.Output.InstructionFilePath, if any.
	Instruction string

	// DirectoryTree is the rendered tree of selected paths.
	DirectoryTree string

	// Files is the post-transform list of files for inclusion.
	Files []fileselect.File

	// Aggregate is the metrics over Files.
	Aggregate metrics.Aggregate

	// GitDiffWorkTree and GitDiffStaged hold optional diff output.
	GitDiffWorkTree string
	GitDiffStaged   string

	// GitLog is an optional list of recent commits.
	GitLog []GitCommit

	// AppName / AppVersion identify the producer.
	AppName    string
	AppVersion string
}

// GitCommit is a minimal commit record for the log section.
type GitCommit struct {
	Hash    string
	Author  string
	Date    string
	Message string
	Files   []string
}

// Renderer writes a Document to a textual representation.
type Renderer interface {
	Render(doc *Document) (string, error)
	FileExtension() string
}

// For returns the renderer for the requested style.
func For(style config.OutputStyle) Renderer {
	switch style {
	case config.StyleMarkdown:
		return &MarkdownRenderer{}
	case config.StyleJSON:
		return &JSONRenderer{}
	case config.StylePlain:
		return &PlainRenderer{}
	case config.StyleXML:
		fallthrough
	default:
		return &XMLRenderer{}
	}
}
