// Package fileselect walks the source tree and applies include/ignore rules
// to produce the ordered list of files that will be packed.
package fileselect

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rajathjn/GopherCram/internal/config"
	"github.com/rajathjn/GopherCram/internal/ignore"
)

// Result describes a single file selected for packing.
type Result struct {
	// AbsPath is the absolute filesystem path.
	AbsPath string
	// RelPath is the path relative to the walk root, slash-separated.
	RelPath string
	// Size is the file size in bytes.
	Size int64
	// IsBinary indicates the file was detected as binary content.
	IsBinary bool
}

// Options configures a walk.
type Options struct {
	// Roots is one or more directories or files to walk. The first root is
	// considered the "primary" anchor used to resolve relative paths.
	Roots []string
	// Include patterns; if empty, all files pass the include filter.
	Include []string
	// CustomIgnores are additional ignore patterns (gitignore-style) supplied
	// via CLI flags or config files.
	CustomIgnores []string
	// UseDefaultPatterns enables the built-in default ignore patterns.
	UseDefaultPatterns bool
	// UseGitignore enables loading .gitignore files encountered during the walk.
	UseGitignore bool
	// UseDotIgnore enables loading .ignore and .gophercramignore files.
	UseDotIgnore bool
	// MaxFileSize is the largest file size we will read; larger files are
	// skipped entirely (not even listed in the tree).
	MaxFileSize int64
	// IncludeEmptyDirectories, when true, surfaces empty directories in the
	// tree generator (walker still does not emit them as Results).
	IncludeEmptyDirectories bool
}

// FromConfig translates a Config and CLI roots into a fully-populated Options.
func FromConfig(cfg *config.Config, roots []string) Options {
	return Options{
		Roots:                   roots,
		Include:                 cfg.Include,
		CustomIgnores:           cfg.Ignore.CustomPatterns,
		UseDefaultPatterns:      cfg.Ignore.UseDefaultPatterns,
		UseGitignore:            cfg.Ignore.UseGitignore,
		UseDotIgnore:            cfg.Ignore.UseDotIgnore,
		MaxFileSize:             cfg.Input.MaxFileSize,
		IncludeEmptyDirectories: cfg.Output.IncludeEmptyDirectories,
	}
}

// Walk traverses the requested roots and returns the files that survive the
// include and ignore filters. Results are sorted by RelPath.
func Walk(opts Options) ([]Result, error) {
	if len(opts.Roots) == 0 {
		opts.Roots = []string{"."}
	}

	baseIgnore := ignore.New()
	if opts.UseDefaultPatterns {
		baseIgnore.AddAll(config.DefaultIgnorePatterns())
	}
	baseIgnore.AddAll(opts.CustomIgnores)

	var (
		results []Result
		seen    = make(map[string]struct{})
	)

	for _, root := range opts.Roots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(absRoot)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			// Single-file root — bypass directory walking.
			r := Result{
				AbsPath: absRoot,
				RelPath: filepath.Base(absRoot),
				Size:    info.Size(),
			}
			r.IsBinary = isBinaryByExt(r.RelPath)
			if _, ok := seen[r.AbsPath]; !ok {
				results = append(results, r)
				seen[r.AbsPath] = struct{}{}
			}
			continue
		}

		// Build the per-root pattern set, layering in .gitignore/.ignore files
		// as we descend.
		rootIgnore := ignore.New()
		rootIgnore.Merge(baseIgnore)

		err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				// Skip unreadable directories rather than aborting the whole walk.
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			rel, err := filepath.Rel(absRoot, path)
			if err != nil {
				return nil
			}
			// Layer per-directory ignore files. We inspect them when we enter
			// any directory (including the root) so nested .gitignore files
			// contribute before we descend into their children.
			if d.IsDir() && (opts.UseGitignore || opts.UseDotIgnore) {
				loadPerDirIgnores(rootIgnore, path, opts)
			}
			if rel == "." {
				return nil
			}
			rel = filepath.ToSlash(rel)

			isDir := d.IsDir()
			if rootIgnore.MatchesPath(rel, isDir) {
				if isDir {
					return filepath.SkipDir
				}
				return nil
			}

			if isDir {
				return nil
			}

			// Apply include filter for files only.
			if len(opts.Include) > 0 && !matchesAnyInclude(opts.Include, rel) {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			if opts.MaxFileSize > 0 && info.Size() > opts.MaxFileSize {
				return nil
			}

			abs, _ := filepath.Abs(path)
			if _, ok := seen[abs]; ok {
				return nil
			}
			seen[abs] = struct{}{}

			results = append(results, Result{
				AbsPath:  abs,
				RelPath:  rel,
				Size:     info.Size(),
				IsBinary: isBinaryByExt(rel),
			})
			return nil
		})
		if err != nil && !errors.Is(err, fs.SkipAll) {
			return nil, err
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].RelPath < results[j].RelPath
	})
	return results, nil
}

// loadPerDirIgnores loads any of `.gitignore`, `.ignore`, or
// `.gophercramignore` found in `dir` and merges their rules into ps.
func loadPerDirIgnores(ps *ignore.PatternSet, dir string, opts Options) {
	files := []string{}
	if opts.UseGitignore {
		files = append(files, ".gitignore")
	}
	if opts.UseDotIgnore {
		files = append(files, ".ignore", ".gophercramignore")
	}
	for _, name := range files {
		_ = ps.LoadFile(filepath.Join(dir, name))
	}
}

func matchesAnyInclude(patterns []string, rel string) bool {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if ignore.Match(p, rel, false) {
			return true
		}
	}
	return false
}

// isBinaryByExt is a cheap pre-check based on file extension. It is later
// verified by content sniffing in fileselect.ReadAll.
func isBinaryByExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return false
	}
	if _, ok := config.BinaryFileExtensions()[ext]; ok {
		return true
	}
	return false
}
