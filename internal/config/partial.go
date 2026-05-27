package config

import (
	"bytes"
	"encoding/json"
)

// PartialConfig mirrors Config but with every leaf field as a pointer so that
// callers can distinguish "field omitted" from "field set to zero value" when
// merging configuration layers (defaults < global file < project file < CLI).
type PartialConfig struct {
	Schema     *string            `json:"$schema,omitempty"`
	Input      *PartialInput      `json:"input,omitempty"`
	Output     *PartialOutput     `json:"output,omitempty"`
	Include    *[]string          `json:"include,omitempty"`
	Ignore     *PartialIgnore     `json:"ignore,omitempty"`
	Security   *PartialSecurity   `json:"security,omitempty"`
	TokenCount *PartialTokenCount `json:"tokenCount,omitempty"`
}

// PartialInput overlays input options.
type PartialInput struct {
	MaxFileSize *int64 `json:"maxFileSize,omitempty"`
}

// PartialGit overlays git options.
type PartialGit struct {
	SortByChanges           *bool `json:"sortByChanges,omitempty"`
	SortByChangesMaxCommits *int  `json:"sortByChangesMaxCommits,omitempty"`
	IncludeDiffs            *bool `json:"includeDiffs,omitempty"`
	IncludeLogs             *bool `json:"includeLogs,omitempty"`
	IncludeLogsCount        *int  `json:"includeLogsCount,omitempty"`
}

// PartialOutput overlays output options.
type PartialOutput struct {
	FilePath                      *string      `json:"filePath,omitempty"`
	Style                         *OutputStyle `json:"style,omitempty"`
	ParsableStyle                 *bool        `json:"parsableStyle,omitempty"`
	HeaderText                    *string      `json:"headerText,omitempty"`
	InstructionFilePath           *string      `json:"instructionFilePath,omitempty"`
	FileSummary                   *bool        `json:"fileSummary,omitempty"`
	DirectoryStructure            *bool        `json:"directoryStructure,omitempty"`
	Files                         *bool        `json:"files,omitempty"`
	RemoveComments                *bool        `json:"removeComments,omitempty"`
	RemoveEmptyLines              *bool        `json:"removeEmptyLines,omitempty"`
	Compress                      *bool        `json:"compress,omitempty"`
	TopFilesLength                *int         `json:"topFilesLength,omitempty"`
	ShowLineNumbers               *bool        `json:"showLineNumbers,omitempty"`
	TruncateBase64                *bool        `json:"truncateBase64,omitempty"`
	CopyToClipboard               *bool        `json:"copyToClipboard,omitempty"`
	IncludeEmptyDirectories       *bool        `json:"includeEmptyDirectories,omitempty"`
	IncludeFullDirectoryStructure *bool        `json:"includeFullDirectoryStructure,omitempty"`
	SplitOutputBytes              *int64       `json:"splitOutput,omitempty"`
	Git                           *PartialGit  `json:"git,omitempty"`
}

// PartialIgnore overlays ignore options.
type PartialIgnore struct {
	UseGitignore       *bool     `json:"useGitignore,omitempty"`
	UseDotIgnore       *bool     `json:"useDotIgnore,omitempty"`
	UseDefaultPatterns *bool     `json:"useDefaultPatterns,omitempty"`
	CustomPatterns     *[]string `json:"customPatterns,omitempty"`
}

// PartialSecurity overlays security options.
type PartialSecurity struct {
	EnableSecurityCheck *bool `json:"enableSecurityCheck,omitempty"`
}

// PartialTokenCount overlays token-count options.
type PartialTokenCount struct {
	Encoding *string `json:"encoding,omitempty"`
}

// LoadPartialFromFile loads a partial config from disk. A non-existent file
// returns (nil, nil).
func LoadPartialFromFile(path string) (*PartialConfig, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var p PartialConfig
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// stripJSONComments returns a reader that drops `//` and `/* */` comments from
// JSON-with-comments style config files. We feed the cleaned bytes through the
// stdlib decoder so we accept either strict JSON or JSON-with-comments.
func stripJSONComments(b []byte) *bytes.Buffer {
	out := bytes.NewBuffer(make([]byte, 0, len(b)))
	const (
		stateCode = iota
		stateString
		stateLine
		stateBlock
	)
	state := stateCode
	for i := 0; i < len(b); i++ {
		c := b[i]
		switch state {
		case stateCode:
			if c == '"' {
				state = stateString
				out.WriteByte(c)
			} else if c == '/' && i+1 < len(b) && b[i+1] == '/' {
				state = stateLine
				i++
			} else if c == '/' && i+1 < len(b) && b[i+1] == '*' {
				state = stateBlock
				i++
			} else {
				out.WriteByte(c)
			}
		case stateString:
			out.WriteByte(c)
			if c == '\\' && i+1 < len(b) {
				out.WriteByte(b[i+1])
				i++
				continue
			}
			if c == '"' {
				state = stateCode
			}
		case stateLine:
			if c == '\n' {
				out.WriteByte(c)
				state = stateCode
			}
		case stateBlock:
			if c == '*' && i+1 < len(b) && b[i+1] == '/' {
				state = stateCode
				i++
			}
		}
	}
	return out
}
