package tools

// catalogSecurity returns the Security & Observability tools.
// Registry concatenates these functions in CategoryOrder. Keep each function's
// tools sorted by name. Add a new category by adding a function and an entry
// in CategoryOrder, rather than relying on alphabetical order.
func catalogSecurity() []Tool {
	return []Tool{
		{
			Name:             "Semgrep",
			Description:      "Static analysis and code security scanning CLI",
			DocumentationURL: "https://semgrep.dev/docs/metrics",
			DataCollection:   "Usage metrics about Semgrep runs, registry usage, and login-related activity.",
			Detect:           []string{"semgrep"},
			Category:         "Security & Observability",
			EnvChecks: []EnvCheck{
				{
					// Semgrep documents the exact value "off".
					Name:        "SEMGREP_SEND_METRICS",
					ValidValues: []string{"off"},
					Description: "Disables Semgrep metrics reporting",
				},
			},
			Commands: []string{"semgrep --metrics=off"},
		},
		{
			Name:             "Sentry CLI",
			Description:      "Sentry error tracking command-line tool",
			DocumentationURL: "https://cli.sentry.dev/configuration/",
			DataCollection:   "Anonymous CLI usage and diagnostic telemetry.",
			Detect:           []string{"sentry-cli"},
			Category:         "Security & Observability",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"SENTRY_CLI_NO_TELEMETRY",
					"Disables Sentry CLI telemetry",
					"1", "true",
				),
			},
		},
	}
}
