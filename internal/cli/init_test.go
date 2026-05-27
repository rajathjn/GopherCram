package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit_Local(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	if err := RunInit(InitOptions{Cwd: dir, Out: &buf}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "gophercram.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("emitted invalid JSON: %v", err)
	}
	out, ok := v["output"].(map[string]any)
	if !ok {
		t.Fatal("missing output section")
	}
	if out["style"] != "xml" {
		t.Errorf("style=%v", out["style"])
	}
}

func TestRunInit_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "gophercram.config.json"), []byte("{}"), 0o644)
	if err := RunInit(InitOptions{Cwd: dir}); err == nil {
		t.Error("expected error for existing config")
	}
}

func TestRunInit_DefaultStdout(t *testing.T) {
	dir := t.TempDir()
	if err := RunInit(InitOptions{Cwd: dir}); err != nil {
		t.Fatal(err)
	}
}

func TestRunInit_Global(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := RunInit(InitOptions{Global: true}); err != nil {
		t.Fatal(err)
	}
}
