package security

import (
	"strings"
	"testing"
)

func TestScan_NoFindings(t *testing.T) {
	s := New()
	if got := s.Scan("p", "plain text with no secrets"); len(got) != 0 {
		t.Errorf("expected zero findings, got %d: %v", len(got), got)
	}
}

func TestScan_AWSAccessKey(t *testing.T) {
	s := New()
	body := "key = AKIAIOSFODNN7EXAMPLE\n"
	got := s.Scan("aws.txt", body)
	if len(got) == 0 {
		t.Fatal("expected AWS finding")
	}
	if !strings.Contains(got[0].Rule, "AWS") {
		t.Errorf("rule=%q", got[0].Rule)
	}
}

func TestScan_GenericSecret(t *testing.T) {
	s := New()
	body := `password = "Sup3rS3cr3tValue!!"`
	got := s.Scan("p", body)
	if len(got) == 0 {
		t.Error("expected generic secret hit")
	}
}

func TestScan_LowEntropySkipped(t *testing.T) {
	s := New()
	body := `password = "aaaaaaaaaaaa"`
	got := s.Scan("p", body)
	if len(got) > 0 {
		t.Errorf("low-entropy match should be skipped, got %v", got)
	}
}

func TestScan_PrivateKey(t *testing.T) {
	s := New()
	body := "-----BEGIN RSA PRIVATE KEY-----\nMIICX...\n-----END RSA PRIVATE KEY-----"
	got := s.Scan("k", body)
	if len(got) == 0 {
		t.Error("expected private key finding")
	}
}

func TestScan_Redact(t *testing.T) {
	s := New()
	body := "AKIAIOSFODNN7EXAMPLE"
	got := s.Scan("p", body)
	if len(got) == 0 {
		t.Fatal("expected hit")
	}
	if got[0].Match == body {
		t.Error("match should be redacted in report")
	}
	if !strings.Contains(got[0].Match, "*") {
		t.Errorf("expected stars, got %q", got[0].Match)
	}
}

func TestScan_NilSafe(t *testing.T) {
	var s *Scanner
	if got := s.Scan("p", "x"); len(got) != 0 {
		t.Error("nil scanner should return nil")
	}
}

func TestShannonEntropy(t *testing.T) {
	if shannonEntropy("") != 0 {
		t.Error("empty -> 0")
	}
	low := shannonEntropy("aaaaaaaa")
	high := shannonEntropy("abcdABCD1234!@#$")
	if low >= high {
		t.Errorf("expected lower entropy for repeats: %f vs %f", low, high)
	}
}

func TestWithExtraRules(t *testing.T) {
	s := New()
	r := builtinRules()[0]
	s.WithExtraRules(r)
	if len(s.rules) <= 1 {
		t.Error("rule not appended")
	}
}

func TestLineNumberAt(t *testing.T) {
	s := "a\nb\nc"
	if lineNumberAt(s, 0) != 1 {
		t.Error("offset 0 should be line 1")
	}
	if lineNumberAt(s, 4) != 3 {
		t.Errorf("offset 4 should be line 3, got %d", lineNumberAt(s, 4))
	}
}

func TestScan_GitHubToken(t *testing.T) {
	s := New()
	body := "token=ghp_abcdef1234567890abcdef1234567890abcdef"
	got := s.Scan("p", body)
	if len(got) == 0 {
		t.Error("expected GitHub token finding")
	}
}
