package pack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rajathjn/GopherCram/internal/config"
)

func makeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mk := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("src/main.go", "package main\n\nfunc main() {}\n")
	mk("src/util.go", "package main\n\nfunc util() {}\n")
	mk("README.md", "# project\nhello\n")
	mk("secrets.txt", "aws=AKIAIOSFODNN7EXAMPLE\n")
	mk(".gitignore", "*.log\n")
	mk("ignored.log", "should be skipped\n")
	return dir
}

func TestPack_Default(t *testing.T) {
	dir := makeFixture(t)
	cfg := config.Defaults()
	cfg.Cwd = dir
	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if res.Metrics.TotalFiles == 0 {
		t.Fatal("expected files")
	}
	if !strings.Contains(res.Output, "<gophercram>") {
		t.Errorf("expected XML output")
	}
	if !strings.Contains(res.Output, "README.md") {
		t.Errorf("expected README in tree")
	}
	if len(res.Findings) == 0 {
		t.Error("expected security findings")
	}
	if _, ok := res.SuspiciousPath["secrets.txt"]; !ok {
		t.Error("secrets.txt should be flagged")
	}
}

func TestPack_GitignoreApplied(t *testing.T) {
	dir := makeFixture(t)
	cfg := config.Defaults()
	cfg.Cwd = dir
	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.Output, "ignored.log") {
		t.Error("ignored.log should be excluded")
	}
}

func TestPack_StylesAllRender(t *testing.T) {
	dir := makeFixture(t)
	pk := New("test", "0.0.1")
	for _, style := range config.AllStyles() {
		t.Run(string(style), func(t *testing.T) {
			cfg := config.Defaults()
			cfg.Cwd = dir
			cfg.Output.Style = style
			res, err := pk.Pack(&cfg, []string{dir})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Output) == 0 {
				t.Error("empty output")
			}
		})
	}
}

func TestPack_SplitOutput(t *testing.T) {
	dir := makeFixture(t)
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.SplitOutputBytes = 200
	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Parts) < 2 {
		t.Errorf("expected multiple parts, got %d", len(res.Parts))
	}
}

func TestWriteToDisk(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.FilePath = "out.xml"
	res := &Result{Output: "<x/>"}
	written, err := WriteToDisk(&cfg, res)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 1 {
		t.Errorf("expected one file, got %v", written)
	}
	body, _ := os.ReadFile(written[0])
	if string(body) != "<x/>" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestWriteToDiskSplit(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Defaults()
	cfg.Cwd = dir
	cfg.Output.FilePath = "out.xml"
	res := &Result{Parts: []string{"one", "two", "three"}}
	written, err := WriteToDisk(&cfg, res)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 3 {
		t.Errorf("expected 3 parts, got %v", written)
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		0:          "0 B",
		512:        "512 B",
		2048:       "2.00 KB",
		1500000:    "1.43 MB",
		2000000000: "1.86 GB",
	}
	for n, want := range cases {
		if got := HumanSize(n); got != want {
			t.Errorf("HumanSize(%d)=%q want %q", n, got, want)
		}
	}
}

func TestReportTopFiles(t *testing.T) {
	dir := makeFixture(t)
	cfg := config.Defaults()
	cfg.Cwd = dir
	pk := New("test", "0.0.1")
	res, err := pk.Pack(&cfg, []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	rpt := ReportTopFiles(res.Metrics, 3)
	if rpt == "" {
		t.Error("expected non-empty report")
	}
	if ReportTopFiles(res.Metrics, 0) != "" {
		t.Error("zero should be empty")
	}
}

func TestPack_InvalidConfig(t *testing.T) {
	cfg := config.Defaults()
	cfg.Output.Style = "bogus"
	pk := New("test", "0.0.1")
	if _, err := pk.Pack(&cfg, []string{"."}); err == nil {
		t.Error("expected validation error")
	}
}

func TestPack_NilConfig(t *testing.T) {
	pk := New("test", "0.0.1")
	if _, err := pk.Pack(nil, []string{"."}); err == nil {
		t.Error("expected error for nil config")
	}
}
