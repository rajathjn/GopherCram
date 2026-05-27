package cli

import (
	"reflect"
	"testing"
)

func TestParseLong(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "verbose", Kind: FlagBool})
	p.Register(FlagSpec{Name: "output", Short: "o", Kind: FlagString})
	p.Register(FlagSpec{Name: "n", Kind: FlagInt})
	args, err := p.Parse([]string{"--verbose", "--output", "out.xml", "--n=3", "a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if !args.Get("verbose").Bool {
		t.Error("verbose should be true")
	}
	if args.Get("output").Str != "out.xml" {
		t.Errorf("output=%q", args.Get("output").Str)
	}
	if args.Get("n").Int != 3 {
		t.Errorf("n=%d", args.Get("n").Int)
	}
	if !reflect.DeepEqual(args.Positional, []string{"a", "b"}) {
		t.Errorf("positional=%v", args.Positional)
	}
}

func TestParseShort(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "output", Short: "o", Kind: FlagString})
	p.Register(FlagSpec{Name: "x", Short: "x", Kind: FlagBool})
	args, err := p.Parse([]string{"-x", "-oXOUT"})
	if err != nil {
		t.Fatal(err)
	}
	if !args.Get("x").Bool || args.Get("output").Str != "XOUT" {
		t.Errorf("got %+v", args.Flags)
	}
}

func TestParseNoFlag(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "thing", Kind: FlagBool})
	args, err := p.Parse([]string{"--no-thing"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Get("thing").Bool {
		t.Error("--no-thing should set false")
	}
	if !args.HasFlag("thing") {
		t.Error("flag should still register as set")
	}
}

func TestParseUnknownFlag(t *testing.T) {
	p := NewParser()
	if _, err := p.Parse([]string{"--whatever"}); err == nil {
		t.Error("expected unknown flag error")
	}
}

func TestParseEndOfFlags(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "verbose", Kind: FlagBool})
	args, err := p.Parse([]string{"--", "--verbose", "x"})
	if err != nil {
		t.Fatal(err)
	}
	if args.HasFlag("verbose") {
		t.Error("flags after -- should be positional")
	}
	if len(args.Positional) != 2 {
		t.Errorf("positional=%v", args.Positional)
	}
}

func TestParseOptionalInt(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "thresh", Kind: FlagOptionalInt})
	args, _ := p.Parse([]string{"--thresh"})
	if args.Get("thresh").HadVal {
		t.Error("no value should not be marked HadVal")
	}
	args, _ = p.Parse([]string{"--thresh=50"})
	if !args.Get("thresh").HadVal || args.Get("thresh").Int != 50 {
		t.Errorf("expected 50, got %+v", args.Get("thresh"))
	}
	args, _ = p.Parse([]string{"--thresh", "10", "extra"})
	if args.Get("thresh").Int != 10 {
		t.Errorf("expected 10, got %d", args.Get("thresh").Int)
	}
}

func TestParseValueRequired(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "out", Kind: FlagString})
	if _, err := p.Parse([]string{"--out"}); err == nil {
		t.Error("expected error for missing value")
	}
}

func TestParseInt64(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "size", Kind: FlagInt64})
	args, err := p.Parse([]string{"--size", "12345"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Get("size").Int64 != 12345 {
		t.Errorf("size=%d", args.Get("size").Int64)
	}
	if _, err := p.Parse([]string{"--size", "abc"}); err == nil {
		t.Error("expected parse error")
	}
}

func TestHumanBytesParse(t *testing.T) {
	cases := map[string]int64{
		"100":   100,
		"100b":  100,
		"1kb":   1024,
		"2mb":   2 * 1024 * 1024,
		"3gb":   3 * 1024 * 1024 * 1024,
		"2.5mb": int64(2.5 * 1024 * 1024),
		"5k":    5 * 1024,
	}
	for in, want := range cases {
		got, err := HumanBytesParse(in)
		if err != nil {
			t.Errorf("HumanBytesParse(%q) err: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("HumanBytesParse(%q)=%d, want %d", in, got, want)
		}
	}
	if _, err := HumanBytesParse(""); err == nil {
		t.Error("empty should err")
	}
	if _, err := HumanBytesParse("xyz"); err == nil {
		t.Error("garbage should err")
	}
}

func TestBoolValueForms(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "x", Kind: FlagBool})
	for in, want := range map[string]bool{
		"true": true, "false": false, "1": true, "0": false,
	} {
		args, err := p.Parse([]string{"--x=" + in})
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if args.Get("x").Bool != want {
			t.Errorf("--x=%s -> %v want %v", in, args.Get("x").Bool, want)
		}
	}
	if _, err := p.Parse([]string{"--x=bogus"}); err == nil {
		t.Error("expected bogus to fail")
	}
}
