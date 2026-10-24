package tools

// catalogAll returns the All The Things tools.
// Registry concatenates these functions in CategoryOrder. Keep each function's
// tools sorted by name. Add a new category by adding a function and an entry
// in CategoryOrder, rather than relying on alphabetical order.
func catalogAll() []Tool {
	return []Tool{
		{
			Name:             "Do Not Track",
			Description:      "Cross-tool environment variable convention for disabling telemetry and tracking",
			DocumentationURL: "https://donottrack.sh",
			DataCollection:   "Signals a global preference to disable telemetry, analytics, crash reporting, and non-essential tracking.",
			// No executable. The convention itself is what the audit checks.
			AlwaysInstalled: true,
			Category:        "All The Things",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"DO_NOT_TRACK",
					"Signals a global do-not-track preference to tools that honor it",
					"1", "true",
				),
			},
		},
	}
}
