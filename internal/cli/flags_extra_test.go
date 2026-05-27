package cli

import "testing"

func TestParse_ShortIntInline(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "count", Short: "n", Kind: FlagInt})
	args, err := p.Parse([]string{"-n42"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Get("count").Int != 42 {
		t.Errorf("expected 42, got %d", args.Get("count").Int)
	}
}

func TestParse_ShortBoolWithValueErrors(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "verbose", Short: "v", Kind: FlagBool})
	if _, err := p.Parse([]string{"-vextra"}); err == nil {
		t.Error("expected error for value attached to bool short flag")
	}
}

func TestParse_ShortIntMissingValue(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "count", Short: "n", Kind: FlagInt})
	if _, err := p.Parse([]string{"-n"}); err == nil {
		t.Error("expected missing-value error")
	}
}

func TestParse_ShortStringMissingValue(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "out", Short: "o", Kind: FlagString})
	if _, err := p.Parse([]string{"-o"}); err == nil {
		t.Error("expected missing-value error")
	}
}

func TestParse_ShortIntInvalid(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "count", Short: "n", Kind: FlagInt})
	if _, err := p.Parse([]string{"-n", "garbage"}); err == nil {
		t.Error("expected parse error for non-numeric")
	}
}

func TestParse_BareDashIsPositional(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "verbose", Kind: FlagBool})
	args, err := p.Parse([]string{"-", "--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if len(args.Positional) != 1 || args.Positional[0] != "-" {
		t.Errorf("positional=%v", args.Positional)
	}
	if !args.Get("verbose").Bool {
		t.Error("verbose should be set")
	}
}

func TestParse_UnknownShortFlag(t *testing.T) {
	p := NewParser()
	if _, err := p.Parse([]string{"-z"}); err == nil {
		t.Error("expected unknown-flag error")
	}
}

func TestParse_OptionalIntInvalidEqValue(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "t", Kind: FlagOptionalInt})
	if _, err := p.Parse([]string{"--t=abc"}); err == nil {
		t.Error("expected error parsing non-numeric optional value")
	}
}

func TestParse_OptionalStringWithValue(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "name", Kind: FlagOptionalString})
	args, _ := p.Parse([]string{"--name=foo"})
	if !args.Get("name").HadVal || args.Get("name").Str != "foo" {
		t.Errorf("got %+v", args.Get("name"))
	}
	args, _ = p.Parse([]string{"--name", "bar"})
	if !args.Get("name").HadVal || args.Get("name").Str != "bar" {
		t.Errorf("got %+v", args.Get("name"))
	}
	args, _ = p.Parse([]string{"--name"})
	if args.Get("name").HadVal {
		t.Errorf("got %+v", args.Get("name"))
	}
}

func TestParse_Int64Inline(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "size", Kind: FlagInt64})
	args, err := p.Parse([]string{"--size=999999"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Get("size").Int64 != 999999 {
		t.Errorf("size=%d", args.Get("size").Int64)
	}
}

func TestParse_IntMissingValue(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "n", Kind: FlagInt})
	if _, err := p.Parse([]string{"--n"}); err == nil {
		t.Error("expected missing-value")
	}
}

func TestParse_Int64MissingValue(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "n", Kind: FlagInt64})
	if _, err := p.Parse([]string{"--n"}); err == nil {
		t.Error("expected missing-value")
	}
}

func TestParse_IntInvalid(t *testing.T) {
	p := NewParser()
	p.Register(FlagSpec{Name: "n", Kind: FlagInt})
	if _, err := p.Parse([]string{"--n=abc"}); err == nil {
		t.Error("expected parse error")
	}
}

func TestParsedArgs_Get_Default(t *testing.T) {
	pa := ParsedArgs{Flags: map[string]*FlagValue{}}
	if got := pa.Get("missing"); got.Set {
		t.Error("expected unset flag value")
	}
}
