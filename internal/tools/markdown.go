package tools

import (
	"fmt"
	"strings"
)

// CatalogMarkdown renders the registry as the Markdown tables in the README.
//
// The output is deterministic. `notel list --format=markdown` prints it, and
// the README test compares the block between the catalog markers to this
// function. Hand-edited tables drift. The Next.js row once contained a Nuxt
// command because the two tables were maintained separately from this registry.
func CatalogMarkdown(list []Tool) string {
	grouped := make(map[string][]Tool)
	var order []string
	for _, tool := range list {
		if _, ok := grouped[tool.Category]; !ok {
			order = append(order, tool.Category)
		}
		grouped[tool.Category] = append(grouped[tool.Category], tool)
	}

	var b strings.Builder
	for i, category := range order {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "### %s\n\n", category)
		b.WriteString("| Tool | Environment Variable | Opt-out Value | Data Collection |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, tool := range grouped[category] {
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
				toolLink(tool),
				envColumn(tool, func(check EnvCheck) string { return "`" + check.Name + "`" }),
				envColumn(tool, optOutValue),
				dataCell(tool),
			)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func toolLink(tool Tool) string {
	return "[" + escapeCell(tool.Name) + "](" + tool.DocumentationURL + ")"
}

// envColumn joins one fragment per env check.
// AnyOf tools use " / " because any one variable is enough.
// FirstSet keeps registry order on separate lines. " / " would read as a
// pure OR, and a later variable does not override an earlier one.
// AllOf tools use a line break because every variable is required.
func envColumn(tool Tool, fragment func(EnvCheck) string) string {
	parts := make([]string, 0, len(tool.EnvChecks))
	for _, check := range tool.EnvChecks {
		parts = append(parts, fragment(check))
	}
	separator := "<br>"
	if tool.OptOut == AnyOf {
		separator = " / "
	}
	return strings.Join(parts, separator)
}

func optOutValue(check EnvCheck) string {
	if len(check.ValidValues) == 0 {
		return "`<any non-empty value>`"
	}
	quoted := make([]string, len(check.ValidValues))
	for i, value := range check.ValidValues {
		// A blank opt-out is a real accepted value. Empty backticks would
		// disappear in the table. "(empty)" is the same spelling the
		// terminal formatter uses.
		if value == "" {
			quoted[i] = "`(empty)`"
			continue
		}
		quoted[i] = "`" + value + "`"
	}
	text := strings.Join(quoted, " or ")
	if check.IgnoreCase {
		text += " (any case)"
	}
	return text
}

func dataCell(tool Tool) string {
	parts := []string{escapeCell(tool.DataCollection)}
	if tool.Default == DefaultDisabled {
		// Unset passes the check. That is not the same as proving the
		// install is opted out: Google Cloud SDK stores the opt-in in its
		// config file, which notel does not read.
		parts = append(parts, "Unset passes. The vendor documents collection as opt-in, and notel does not read config files.")
	}
	if tool.OptOut == AnyOf {
		parts = append(parts, "Any one of the variables opts out.")
	}
	if tool.OptOut == FirstSet {
		parts = append(parts, "The first variable that is set decides. Later variables apply only when it is absent.")
	}
	for _, command := range tool.Commands {
		// Commands are hints. The backticks keep them visually distinct from
		// the prose description of what the tool collects.
		parts = append(parts, "`"+escapeCell(command)+"`")
	}
	if len(tool.Detect) == 0 && !tool.AlwaysInstalled {
		parts = append(parts, "notel does not detect this tool from PATH. It is listed here and omitted from the default audit.")
	}
	return strings.Join(parts, "<br>")
}

func escapeCell(text string) string {
	text = strings.ReplaceAll(text, "|", "\\|")
	text = strings.ReplaceAll(text, "\n", " ")
	return text
}
