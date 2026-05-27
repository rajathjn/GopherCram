package cli

import (
	"sort"
	"strings"
)

// HelpText renders a structured help screen for the parser.
func HelpText(p *Parser) string {
	var b strings.Builder
	b.WriteString(AppName + " v" + Version + "\n")
	b.WriteString("Pack a repository into a single file you can hand to a language model.\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  gophercram [options] [path ...]\n\n")
	b.WriteString("If no path is given, the current directory is used.\n\n")

	groups := map[string][]FlagSpec{}
	order := []string{}
	for _, s := range p.Specs() {
		if _, ok := groups[s.Group]; !ok {
			order = append(order, s.Group)
		}
		groups[s.Group] = append(groups[s.Group], s)
	}
	sort.SliceStable(order, func(i, j int) bool {
		return groupRank(order[i]) < groupRank(order[j])
	})

	for _, g := range order {
		b.WriteString(g + " options:\n")
		for _, s := range groups[g] {
			b.WriteString("  ")
			b.WriteString(flagDisplay(s))
			b.WriteString("\n      ")
			b.WriteString(s.Description)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Examples:\n")
	b.WriteString("  gophercram .\n")
	b.WriteString("  gophercram --style markdown --output context.md src/ docs/\n")
	b.WriteString("  gophercram --remote owner/repo --remote-branch main\n")
	b.WriteString("  gophercram --include 'src/**/*.go' --ignore 'src/**/*_test.go'\n")
	b.WriteString("  gophercram --init             # write a default config file\n")
	return b.String()
}

func flagDisplay(s FlagSpec) string {
	var parts []string
	if s.Short != "" {
		parts = append(parts, "-"+s.Short)
	}
	parts = append(parts, "--"+s.Name)
	for _, a := range s.Aliases {
		parts = append(parts, "--"+a)
	}
	prefix := strings.Join(parts, ", ")
	switch s.Kind {
	case FlagBool:
		return prefix
	case FlagString:
		return prefix + " <value>"
	case FlagInt, FlagInt64:
		return prefix + " <n>"
	case FlagOptionalInt:
		return prefix + " [n]"
	case FlagOptionalString:
		return prefix + " [value]"
	}
	return prefix
}

func groupRank(g string) int {
	switch g {
	case "Logging":
		return 1
	case "Output":
		return 2
	case "Selection":
		return 3
	case "Remote":
		return 4
	case "Config":
		return 5
	case "Security":
		return 6
	case "Tokens":
		return 7
	case "Meta":
		return 8
	}
	return 99
}
