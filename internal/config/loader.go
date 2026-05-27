package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ConfigFilename is the default name GopherCram looks for in the project root.
const ConfigFilename = "gophercram.config.json"

// AlternateFilenames are additional filenames recognised when the default is
// not present. This lets users migrate from existing repomix configs.
var AlternateFilenames = []string{
	"gophercram.json",
	"repomix.config.json", // compatibility with repomix-formatted configs
}

// LoadFromFile reads and parses a config file. A nil error and nil config
// indicates the file did not exist (which is not an error in itself).
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := &Config{}
	dec := json.NewDecoder(stripJSONComments(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(cfg); err != nil {
		// retry without strict mode in case the file has fields we don't model
		if err2 := json.Unmarshal(data, cfg); err2 != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}
	}
	return cfg, nil
}

// Discover walks up from cwd looking for a config file. It returns the path
// of the discovered file, or "" if none was found.
func Discover(cwd string) string {
	candidates := append([]string{ConfigFilename}, AlternateFilenames...)
	for _, name := range candidates {
		p := filepath.Join(cwd, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// GlobalConfigPath returns the per-user global config path. The first
// existing file (or the default location) is returned.
func GlobalConfigPath() string {
	if cfg := os.Getenv("XDG_CONFIG_HOME"); cfg != "" {
		return filepath.Join(cfg, "gophercram", ConfigFilename)
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "GopherCram", ConfigFilename)
		}
	}
	return filepath.Join(home, ".config", "gophercram", ConfigFilename)
}

// Merge layers a partial config (from file or CLI) onto a base config. Zero
// values in the overlay are treated as "unset" and do not override the base
// for booleans / numbers; slices and strings replace when non-empty.
//
// To express "set this boolean to false" from a config file, use a pointer-
// based partial loader (see PartialConfig).
func Merge(base Config, overlay PartialConfig) Config {
	out := base
	if overlay.Schema != nil {
		out.Schema = *overlay.Schema
	}
	if overlay.Input != nil {
		if overlay.Input.MaxFileSize != nil {
			out.Input.MaxFileSize = *overlay.Input.MaxFileSize
		}
	}
	if overlay.Output != nil {
		o := overlay.Output
		if o.FilePath != nil {
			out.Output.FilePath = *o.FilePath
		}
		if o.Style != nil {
			out.Output.Style = *o.Style
		}
		if o.ParsableStyle != nil {
			out.Output.ParsableStyle = *o.ParsableStyle
		}
		if o.HeaderText != nil {
			out.Output.HeaderText = *o.HeaderText
		}
		if o.InstructionFilePath != nil {
			out.Output.InstructionFilePath = *o.InstructionFilePath
		}
		if o.FileSummary != nil {
			out.Output.FileSummary = *o.FileSummary
		}
		if o.DirectoryStructure != nil {
			out.Output.DirectoryStructure = *o.DirectoryStructure
		}
		if o.Files != nil {
			out.Output.Files = *o.Files
		}
		if o.RemoveComments != nil {
			out.Output.RemoveComments = *o.RemoveComments
		}
		if o.RemoveEmptyLines != nil {
			out.Output.RemoveEmptyLines = *o.RemoveEmptyLines
		}
		if o.Compress != nil {
			out.Output.Compress = *o.Compress
		}
		if o.TopFilesLength != nil {
			out.Output.TopFilesLength = *o.TopFilesLength
		}
		if o.ShowLineNumbers != nil {
			out.Output.ShowLineNumbers = *o.ShowLineNumbers
		}
		if o.TruncateBase64 != nil {
			out.Output.TruncateBase64 = *o.TruncateBase64
		}
		if o.CopyToClipboard != nil {
			out.Output.CopyToClipboard = *o.CopyToClipboard
		}
		if o.IncludeEmptyDirectories != nil {
			out.Output.IncludeEmptyDirectories = *o.IncludeEmptyDirectories
		}
		if o.IncludeFullDirectoryStructure != nil {
			out.Output.IncludeFullDirectoryStructure = *o.IncludeFullDirectoryStructure
		}
		if o.SplitOutputBytes != nil {
			out.Output.SplitOutputBytes = *o.SplitOutputBytes
		}
		if o.Git != nil {
			g := o.Git
			if g.SortByChanges != nil {
				out.Output.Git.SortByChanges = *g.SortByChanges
			}
			if g.SortByChangesMaxCommits != nil {
				out.Output.Git.SortByChangesMaxCommits = *g.SortByChangesMaxCommits
			}
			if g.IncludeDiffs != nil {
				out.Output.Git.IncludeDiffs = *g.IncludeDiffs
			}
			if g.IncludeLogs != nil {
				out.Output.Git.IncludeLogs = *g.IncludeLogs
			}
			if g.IncludeLogsCount != nil {
				out.Output.Git.IncludeLogsCount = *g.IncludeLogsCount
			}
		}
	}
	if overlay.Include != nil {
		out.Include = append([]string{}, *overlay.Include...)
	}
	if overlay.Ignore != nil {
		i := overlay.Ignore
		if i.UseGitignore != nil {
			out.Ignore.UseGitignore = *i.UseGitignore
		}
		if i.UseDotIgnore != nil {
			out.Ignore.UseDotIgnore = *i.UseDotIgnore
		}
		if i.UseDefaultPatterns != nil {
			out.Ignore.UseDefaultPatterns = *i.UseDefaultPatterns
		}
		if i.CustomPatterns != nil {
			out.Ignore.CustomPatterns = append([]string{}, *i.CustomPatterns...)
		}
	}
	if overlay.Security != nil {
		if overlay.Security.EnableSecurityCheck != nil {
			out.Security.EnableSecurityCheck = *overlay.Security.EnableSecurityCheck
		}
	}
	if overlay.TokenCount != nil {
		if overlay.TokenCount.Encoding != nil {
			out.TokenCount.Encoding = *overlay.TokenCount.Encoding
		}
	}
	return out
}
