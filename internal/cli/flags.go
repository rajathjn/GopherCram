// Package cli implements the command-line interface for GopherCram.
//
// The parser in this package is intentionally minimal: it supports POSIX-style
// long and short options, value-or-no-value forms, repeated --include style
// arguments, and a `--` end-of-options separator. It exists so the binary
// stays dependency-free.
package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// FlagKind tells the parser how to consume the next argv element after an
// option.
type FlagKind int

const (
	// FlagBool is a switch that takes no value (--copy, --quiet).
	FlagBool FlagKind = iota
	// FlagString consumes exactly one value (--output FILE).
	FlagString
	// FlagInt consumes one integer value.
	FlagInt
	// FlagInt64 consumes one int64.
	FlagInt64
	// FlagOptionalInt accepts either no value (default 0) or one integer.
	FlagOptionalInt
	// FlagOptionalString accepts either no value or one string.
	FlagOptionalString
)

// FlagSpec describes a single option for the parser.
type FlagSpec struct {
	Name        string   // canonical name, e.g. "output"
	Short       string   // single-char alias, e.g. "o"
	Aliases     []string // additional long aliases
	Kind        FlagKind
	Description string
	Group       string
	// NoFlag is true when --no-flag toggles the inverse. We auto-generate
	// the negative form for FlagBool by default; set this to false to
	// suppress it.
	NoFlag bool
}

// FlagValue is the parsed value of a single flag. Only the relevant field is
// populated based on Spec.Kind.
type FlagValue struct {
	Spec    FlagSpec
	Set     bool
	Bool    bool
	Str     string
	Int     int
	Int64   int64
	HadVal  bool // for optional-value flags
	StrList []string
}

// ParsedArgs is the result of parsing argv.
type ParsedArgs struct {
	Positional []string
	Flags      map[string]*FlagValue
}

// HasFlag reports whether the user explicitly passed `name` on the CLI.
func (p ParsedArgs) HasFlag(name string) bool {
	v, ok := p.Flags[name]
	return ok && v.Set
}

// Get returns the FlagValue for `name`, creating an empty one if absent.
func (p ParsedArgs) Get(name string) *FlagValue {
	if v, ok := p.Flags[name]; ok {
		return v
	}
	return &FlagValue{}
}

// Parser holds a set of flag specs.
type Parser struct {
	specs []FlagSpec
	byKey map[string]int
}

// NewParser returns an empty Parser.
func NewParser() *Parser {
	return &Parser{byKey: map[string]int{}}
}

// Register adds a flag spec to the parser.
func (p *Parser) Register(spec FlagSpec) {
	idx := len(p.specs)
	p.specs = append(p.specs, spec)
	p.byKey["--"+spec.Name] = idx
	if spec.Kind == FlagBool {
		p.byKey["--no-"+spec.Name] = idx
	}
	for _, a := range spec.Aliases {
		p.byKey["--"+a] = idx
	}
	if spec.Short != "" {
		p.byKey["-"+spec.Short] = idx
	}
}

// Specs returns the registered flags (for help generation).
func (p *Parser) Specs() []FlagSpec { return p.specs }

// Parse walks argv (excluding argv[0]) and populates ParsedArgs.
func (p *Parser) Parse(argv []string) (*ParsedArgs, error) {
	pa := &ParsedArgs{Flags: map[string]*FlagValue{}}
	i := 0
	endOfFlags := false
	for i < len(argv) {
		tok := argv[i]
		if endOfFlags {
			pa.Positional = append(pa.Positional, tok)
			i++
			continue
		}
		if tok == "--" {
			endOfFlags = true
			i++
			continue
		}
		if !strings.HasPrefix(tok, "-") || tok == "-" {
			pa.Positional = append(pa.Positional, tok)
			i++
			continue
		}

		// long form
		if strings.HasPrefix(tok, "--") {
			name, val, hasEq := strings.Cut(tok, "=")
			idx, ok := p.byKey[name]
			if !ok {
				return nil, fmt.Errorf("unknown flag %q", name)
			}
			spec := p.specs[idx]
			fv, exists := pa.Flags[spec.Name]
			if !exists {
				fv = &FlagValue{Spec: spec}
				pa.Flags[spec.Name] = fv
			}
			fv.Set = true
			// negated long form
			if spec.Kind == FlagBool && strings.HasPrefix(name, "--no-") {
				fv.Bool = false
				i++
				continue
			}
			switch spec.Kind {
			case FlagBool:
				fv.Bool = true
				if hasEq {
					switch strings.ToLower(val) {
					case "true", "1", "yes", "on":
						fv.Bool = true
					case "false", "0", "no", "off":
						fv.Bool = false
					default:
						return nil, fmt.Errorf("invalid boolean for --%s: %q", spec.Name, val)
					}
				}
				i++
			case FlagString:
				if !hasEq {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("flag --%s requires a value", spec.Name)
					}
					val = argv[i+1]
					i += 2
				} else {
					i++
				}
				fv.Str = val
				fv.StrList = append(fv.StrList, val)
			case FlagInt:
				if !hasEq {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("flag --%s requires a value", spec.Name)
					}
					val = argv[i+1]
					i += 2
				} else {
					i++
				}
				n, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid integer for --%s: %q", spec.Name, val)
				}
				fv.Int = n
			case FlagInt64:
				if !hasEq {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("flag --%s requires a value", spec.Name)
					}
					val = argv[i+1]
					i += 2
				} else {
					i++
				}
				n, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid integer for --%s: %q", spec.Name, val)
				}
				fv.Int64 = n
			case FlagOptionalInt:
				if hasEq {
					n, err := strconv.Atoi(val)
					if err != nil {
						return nil, fmt.Errorf("invalid integer for --%s: %q", spec.Name, val)
					}
					fv.Int = n
					fv.HadVal = true
					i++
				} else if i+1 < len(argv) && !strings.HasPrefix(argv[i+1], "-") {
					n, err := strconv.Atoi(argv[i+1])
					if err == nil {
						fv.Int = n
						fv.HadVal = true
						i += 2
					} else {
						i++
					}
				} else {
					i++
				}
			case FlagOptionalString:
				if hasEq {
					fv.Str = val
					fv.HadVal = true
					i++
				} else if i+1 < len(argv) && !strings.HasPrefix(argv[i+1], "-") {
					fv.Str = argv[i+1]
					fv.HadVal = true
					i += 2
				} else {
					i++
				}
			}
			continue
		}

		// short form -X(value)
		if len(tok) >= 2 {
			short := tok[:2]
			rest := tok[2:]
			idx, ok := p.byKey[short]
			if !ok {
				return nil, fmt.Errorf("unknown flag %q", short)
			}
			spec := p.specs[idx]
			fv, exists := pa.Flags[spec.Name]
			if !exists {
				fv = &FlagValue{Spec: spec}
				pa.Flags[spec.Name] = fv
			}
			fv.Set = true
			switch spec.Kind {
			case FlagBool:
				fv.Bool = true
				if rest != "" {
					return nil, fmt.Errorf("flag -%s does not accept a value", spec.Short)
				}
				i++
			case FlagString:
				if rest != "" {
					fv.Str = rest
					fv.StrList = append(fv.StrList, rest)
					i++
				} else {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("flag -%s requires a value", spec.Short)
					}
					fv.Str = argv[i+1]
					fv.StrList = append(fv.StrList, argv[i+1])
					i += 2
				}
			case FlagInt:
				v := rest
				if v == "" {
					if i+1 >= len(argv) {
						return nil, fmt.Errorf("flag -%s requires a value", spec.Short)
					}
					v = argv[i+1]
					i += 2
				} else {
					i++
				}
				n, err := strconv.Atoi(v)
				if err != nil {
					return nil, fmt.Errorf("invalid integer for -%s: %q", spec.Short, v)
				}
				fv.Int = n
			default:
				return nil, fmt.Errorf("flag -%s not supported in short form", spec.Short)
			}
			continue
		}
		return nil, errors.New("unexpected token: " + tok)
	}
	return pa, nil
}

// HumanBytesParse converts strings like "500kb", "2mb", "2.5mb" to bytes.
func HumanBytesParse(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, errors.New("empty size")
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "gb"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	case strings.HasSuffix(s, "mb"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	case strings.HasSuffix(s, "kb"):
		mult = 1024
		s = strings.TrimSuffix(s, "kb")
	case strings.HasSuffix(s, "g"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "g")
	case strings.HasSuffix(s, "m"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "m")
	case strings.HasSuffix(s, "k"):
		mult = 1024
		s = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "b"):
		s = strings.TrimSuffix(s, "b")
	}
	s = strings.TrimSpace(s)
	if strings.ContainsRune(s, '.') {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return int64(f * float64(mult)), nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return n * mult, nil
}
