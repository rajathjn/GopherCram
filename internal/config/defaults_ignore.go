package config

// DefaultIgnorePatterns returns the built-in ignore patterns. These cover
// version control directories, language ecosystems, build outputs, editor
// scratch files, and previous GopherCram outputs.
//
// The patterns use gitignore-style globbing (with `**` for recursive match).
func DefaultIgnorePatterns() []string {
	return []string{
		// Version control metadata
		".git/**",
		".hg/**",
		".hgignore",
		".svn/**",

		// JavaScript ecosystem
		"**/node_modules/**",
		"**/bower_components/**",
		"**/jspm_packages/**",
		"**/.npm/**",
		"**/.yarn/**",
		"**/.yarn-integrity",
		"**/package-lock.json",
		"**/yarn.lock",
		"**/yarn-error.log",
		"**/pnpm-lock.yaml",
		"**/bun.lockb",
		"**/bun.lock",

		// Python ecosystem
		"**/__pycache__/**",
		"**/*.py[cod]",
		"**/venv/**",
		"**/.venv/**",
		"**/.pytest_cache/**",
		"**/.mypy_cache/**",
		"**/.ipynb_checkpoints/**",
		"**/Pipfile.lock",
		"**/poetry.lock",
		"**/uv.lock",

		// Rust ecosystem
		"**/Cargo.lock",
		"**/Cargo.toml.orig",
		"**/target/**",
		"**/*.rs.bk",

		// PHP, Ruby, Elixir, Haskell ecosystems
		"**/composer.lock",
		"**/Gemfile.lock",
		"**/mix.lock",
		"**/stack.yaml.lock",
		"**/cabal.project.freeze",

		// Go ecosystem
		"**/go.sum",

		// JVM ecosystems
		"**/.gradle/**",
		"**/.bundle/**",
		"vendor/**",
		"target/**",

		// Build outputs
		"build/**",
		"build/Release/**",
		"out/**",
		"dist/**",
		"typings/**",

		// Logs / process state
		"logs/**",
		"**/*.log",
		"**/npm-debug.log*",
		"**/yarn-debug.log*",
		"**/yarn-error.log*",
		"pids/**",
		"*.pid",
		"*.seed",
		"*.pid.lock",

		// Coverage and CI tool caches
		"lib-cov/**",
		"coverage/**",
		".nyc_output/**",
		".grunt/**",
		".lock-wscript",

		// Generic caches
		".eslintcache",
		".rollup.cache/**",
		".webpack.cache/**",
		".parcel-cache/**",
		".sass-cache/**",
		"*.cache",
		".fusebox/**",
		".dynamodb/**",
		".serverless/**",

		// Framework build outputs
		".next/**",
		".nuxt/**",
		".vuepress/dist/**",

		// Misc runtime / packaging artefacts
		".node_repl_history",
		"*.tgz",
		".env",

		// OS-generated junk
		"**/.DS_Store",
		"**/Thumbs.db",

		// Editor scratch files
		".idea/**",
		".vscode/**",
		"**/*.swp",
		"**/*.swo",
		"**/*.swn",
		"**/*.bak",

		// Temporary directories
		"tmp/**",
		"temp/**",

		// Previous GopherCram/repomix output files
		"**/gophercram-output.*",
		"**/repomix-output.*",
		"**/repopack-output.*",
	}
}

// BinaryFileExtensions returns a set of file extensions GopherCram treats as
// binary by default. Binary files are excluded from content output and only
// recorded by name in directory listings.
func BinaryFileExtensions() map[string]struct{} {
	exts := []string{
		// Images
		".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".ico", ".tiff",
		".tif", ".heic", ".heif", ".avif", ".psd", ".raw", ".cr2", ".nef",
		// Vector / design (may be text but commonly large/binary-ish)
		".eps", ".ai", ".sketch", ".fig",
		// Audio
		".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a", ".wma", ".opus",
		// Video
		".mp4", ".mkv", ".mov", ".avi", ".wmv", ".webm", ".flv", ".m4v",
		// Archives
		".zip", ".tar", ".gz", ".tgz", ".bz2", ".xz", ".7z", ".rar", ".zst",
		// Executables / binaries
		".exe", ".dll", ".so", ".dylib", ".a", ".lib", ".o", ".obj",
		".class", ".jar", ".war", ".ear", ".pyc", ".pyo", ".wasm",
		// Documents (binary formats)
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".odt", ".ods", ".odp",
		// Fonts
		".ttf", ".otf", ".woff", ".woff2", ".eot",
		// Database files
		".db", ".sqlite", ".sqlite3", ".mdb",
		// Misc binary
		".iso", ".dmg", ".pkg", ".deb", ".rpm",
	}
	out := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		out[e] = struct{}{}
	}
	return out
}
