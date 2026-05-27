// Package security scans file contents for embedded secrets such as API
// keys, certificates, and connection strings, so callers can omit risky
// files from the packed output.
//
// The scanner uses a curated set of regular expressions tuned for low false
// positives. Each rule has a name (used in reports) and an optional minimum
// entropy threshold; if the matched value's Shannon entropy is too low we
// treat the hit as a false positive and skip it.
package security

import (
	"math"
	"regexp"
)

// Finding is a single suspected secret occurrence.
type Finding struct {
	Path  string
	Line  int
	Rule  string
	Match string
}

// Rule describes a secret-detection rule.
type Rule struct {
	Name       string
	Pattern    *regexp.Regexp
	MinEntropy float64
}

// builtinRules is the default rule set.
func builtinRules() []Rule {
	return []Rule{
		{Name: "AWS Access Key", Pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
		{Name: "AWS Secret Key", Pattern: regexp.MustCompile(`(?i)aws_secret_access_key\s*[:=]\s*['"]?[A-Za-z0-9/+=]{40}['"]?`), MinEntropy: 3.5},
		{Name: "GitHub Token", Pattern: regexp.MustCompile(`gh[pousr]_[0-9A-Za-z]{36}`)},
		{Name: "GitHub OAuth", Pattern: regexp.MustCompile(`gho_[0-9A-Za-z]{36}`)},
		{Name: "GitLab Personal Access Token", Pattern: regexp.MustCompile(`glpat-[0-9A-Za-z\-_]{20}`)},
		{Name: "Google API Key", Pattern: regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`)},
		{Name: "Slack Token", Pattern: regexp.MustCompile(`xox[baprs]-[0-9A-Za-z\-]{10,}`)},
		{Name: "Slack Webhook", Pattern: regexp.MustCompile(`https://hooks\.slack\.com/services/[A-Z0-9/]+`)},
		{Name: "Stripe Secret Key", Pattern: regexp.MustCompile(`sk_live_[0-9A-Za-z]{24,}`)},
		{Name: "Stripe Restricted Key", Pattern: regexp.MustCompile(`rk_live_[0-9A-Za-z]{24,}`)},
		{Name: "Twilio API Key", Pattern: regexp.MustCompile(`SK[0-9a-fA-F]{32}`)},
		{Name: "SendGrid API Key", Pattern: regexp.MustCompile(`SG\.[0-9A-Za-z\-_]{22}\.[0-9A-Za-z\-_]{43}`)},
		{Name: "Mailgun API Key", Pattern: regexp.MustCompile(`key-[0-9a-f]{32}`)},
		{Name: "PEM Private Key", Pattern: regexp.MustCompile(`-----BEGIN ((RSA|EC|DSA|OPENSSH|PGP) )?PRIVATE KEY-----`)},
		{Name: "JWT Token", Pattern: regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`), MinEntropy: 3.8},
		{Name: "Generic API Key", Pattern: regexp.MustCompile(`(?i)api[_-]?key['"]?\s*[:=]\s*['"][A-Za-z0-9_\-]{16,}['"]`), MinEntropy: 3.5},
		{Name: "Generic Secret", Pattern: regexp.MustCompile(`(?i)(secret|password|passwd|pwd|token)['"]?\s*[:=]\s*['"][^'"\s]{8,}['"]`), MinEntropy: 3.0},
		{Name: "Postgres Connection", Pattern: regexp.MustCompile(`postgres(?:ql)?://[^:]+:[^@]+@[^/]+/[^\s]+`)},
		{Name: "MongoDB Connection", Pattern: regexp.MustCompile(`mongodb(?:\+srv)?://[^:]+:[^@]+@[^/]+`)},
		{Name: "MySQL Connection", Pattern: regexp.MustCompile(`mysql://[^:]+:[^@]+@[^/]+/[^\s]+`)},
		{Name: "Redis Connection", Pattern: regexp.MustCompile(`redis://[^:]*:[^@]+@[^/]+`)},
	}
}

// Scanner runs a configured set of rules against content.
type Scanner struct {
	rules []Rule
}

// New returns a Scanner pre-loaded with the built-in rules.
func New() *Scanner { return &Scanner{rules: builtinRules()} }

// WithExtraRules appends extra rules to the scanner.
func (s *Scanner) WithExtraRules(rules ...Rule) *Scanner {
	s.rules = append(s.rules, rules...)
	return s
}

// Scan returns every finding in `content`. `path` is recorded on each finding.
func (s *Scanner) Scan(path, content string) []Finding {
	if s == nil || len(s.rules) == 0 || content == "" {
		return nil
	}
	var out []Finding
	for _, r := range s.rules {
		idxs := r.Pattern.FindAllStringIndex(content, -1)
		for _, idx := range idxs {
			match := content[idx[0]:idx[1]]
			if r.MinEntropy > 0 && shannonEntropy(match) < r.MinEntropy {
				continue
			}
			out = append(out, Finding{
				Path:  path,
				Line:  lineNumberAt(content, idx[0]),
				Rule:  r.Name,
				Match: redact(match),
			})
		}
	}
	return out
}

// shannonEntropy computes the Shannon entropy (bits/byte) of the input. Used
// to discard low-entropy false positives like obvious placeholder secrets.
func shannonEntropy(s string) float64 {
	if s == "" {
		return 0
	}
	var counts [256]int
	for i := 0; i < len(s); i++ {
		counts[s[i]]++
	}
	var ent float64
	n := float64(len(s))
	for _, c := range counts {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		ent -= p * math.Log2(p)
	}
	return ent
}

// redact replaces the middle of a match with '*' so report output doesn't
// itself leak the secret.
func redact(s string) string {
	if len(s) <= 8 {
		return "[redacted]"
	}
	keep := 3
	masked := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		switch {
		case i < keep, i >= len(s)-keep:
			masked[i] = s[i]
		default:
			masked[i] = '*'
		}
	}
	return string(masked)
}

func lineNumberAt(s string, off int) int {
	if off > len(s) {
		off = len(s)
	}
	line := 1
	for i := 0; i < off; i++ {
		if s[i] == '\n' {
			line++
		}
	}
	return line
}
