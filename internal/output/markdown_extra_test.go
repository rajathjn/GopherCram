package output

import "testing"

func TestLanguageFromPath_AllKnown(t *testing.T) {
	cases := map[string]string{
		"a.go":      "go",
		"a.py":      "python",
		"a.js":      "javascript",
		"a.cjs":     "javascript",
		"a.mjs":     "javascript",
		"a.jsx":     "jsx",
		"a.ts":      "typescript",
		"a.tsx":     "tsx",
		"a.rb":      "ruby",
		"a.rs":      "rust",
		"a.java":    "java",
		"a.kt":      "kotlin",
		"a.kts":     "kotlin",
		"a.swift":   "swift",
		"a.c":       "c",
		"a.h":       "c",
		"a.cpp":     "cpp",
		"a.cc":      "cpp",
		"a.cxx":     "cpp",
		"a.hpp":     "cpp",
		"a.cs":      "csharp",
		"a.php":     "php",
		"a.sh":      "bash",
		"a.bash":    "bash",
		"a.zsh":     "zsh",
		"a.fish":    "fish",
		"a.yml":     "yaml",
		"a.yaml":    "yaml",
		"a.toml":    "toml",
		"a.json":    "json",
		"a.xml":     "xml",
		"a.html":    "html",
		"a.htm":     "html",
		"a.css":     "css",
		"a.scss":    "scss",
		"a.sql":     "sql",
		"a.md":      "markdown",
		"a.markdown": "markdown",
		"a.lua":     "lua",
		"a.dart":    "dart",
		"a.scala":   "scala",
		"a.pl":      "perl",
		"a.pm":      "perl",
		"a.r":       "r",
		"none":      "",
		"a.unknown": "",
	}
	for in, want := range cases {
		if got := languageFromPath(in); got != want {
			t.Errorf("languageFromPath(%q)=%q, want %q", in, got, want)
		}
	}
}
