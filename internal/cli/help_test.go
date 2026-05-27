package cli

import (
	"strings"
	"testing"
)

func TestHelpText(t *testing.T) {
	p := BuildParser()
	out := HelpText(p)
	for _, want := range []string{
		"Usage:",
		"Logging options",
		"Output options",
		"Selection options",
		"Remote options",
		"Config options",
		"--output",
		"--style",
		"--remote",
		"Examples",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestFlagDisplay(t *testing.T) {
	cases := []struct {
		spec FlagSpec
		want string
	}{
		{FlagSpec{Name: "verbose", Kind: FlagBool}, "--verbose"},
		{FlagSpec{Name: "out", Short: "o", Kind: FlagString}, "-o, --out <value>"},
		{FlagSpec{Name: "n", Kind: FlagInt}, "--n <n>"},
		{FlagSpec{Name: "t", Kind: FlagOptionalInt}, "--t [n]"},
		{FlagSpec{Name: "s", Kind: FlagOptionalString}, "--s [value]"},
		{FlagSpec{Name: "a", Aliases: []string{"alt"}, Kind: FlagBool}, "--a, --alt"},
	}
	for _, c := range cases {
		if got := flagDisplay(c.spec); got != c.want {
			t.Errorf("flagDisplay(%+v)=%q, want %q", c.spec, got, c.want)
		}
	}
}

func TestGroupRank(t *testing.T) {
	if groupRank("Logging") >= groupRank("Output") {
		t.Error("Logging should come before Output")
	}
	if groupRank("Unknown") != 99 {
		t.Error("unknown groups should sort to the end")
	}
}
