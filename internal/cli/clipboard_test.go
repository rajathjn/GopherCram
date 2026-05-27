package cli

import "testing"

func TestCopyToClipboard(t *testing.T) {
	// The function may succeed or fail depending on the host. We just want
	// to ensure it doesn't panic and returns a sensible result type.
	tool, err := CopyToClipboard("test")
	if err != nil && tool != "" {
		t.Error("tool should be empty when err != nil")
	}
	if err == nil && tool == "" {
		t.Error("expected tool name on success")
	}
}
