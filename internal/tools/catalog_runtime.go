package tools

// catalogRuntime returns the Runtimes & SDKs tools.
// Registry concatenates these functions in CategoryOrder. Keep each function's
// tools sorted by name. Add a new category by adding a function and an entry
// in CategoryOrder, rather than relying on alphabetical order.
func catalogRuntime() []Tool {
	return []Tool{
		{
			Name:             ".NET SDK",
			Description:      "Microsoft .NET SDK and CLI",
			DocumentationURL: "https://learn.microsoft.com/en-us/dotnet/core/tools/telemetry",
			DataCollection:   "CLI command usage, SDK versions, and exception metadata.",
			Detect:           []string{"dotnet"},
			Category:         "Runtimes & SDKs",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"DOTNET_CLI_TELEMETRY_OPTOUT",
					"Disables .NET CLI telemetry collection",
					"1", "true",
				),
			},
		},
		{
			Name:             "Bun",
			Description:      "Fast all-in-one JavaScript runtime and toolkit",
			DocumentationURL: "https://bun.sh/docs/runtime/bunfig#telemetry",
			DataCollection:   "Anonymous telemetry and crash-reporting signals from Bun tooling.",
			Detect:           []string{"bun"},
			Category:         "Runtimes & SDKs",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"DO_NOT_TRACK",
					"Disables Bun telemetry via the global do-not-track convention",
					"1", "true",
				),
			},
		},
		{
			Name:             "Flutter",
			Description:      "Google's UI toolkit for cross-platform apps",
			DocumentationURL: "https://docs.flutter.dev/reference/crash-reporting",
			DataCollection:   "CLI crash reports, tool usage signals, and local environment metadata.",
			Detect:           []string{"flutter"},
			Category:         "Runtimes & SDKs",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"FLUTTER_CLI_CRASH_REPORTING",
					"Disables Flutter CLI crash reporting. This does not disable general analytics.",
					"false", "0",
				),
			},
			// The analytics setting has no environment variable. The command is
			// a hint only. A passing crash-reporting check does not mean
			// `flutter config --no-analytics` has been run.
			Commands: []string{"flutter config --no-analytics"},
		},
		{
			Name:             "Node.js",
			Description:      "JavaScript runtime built on Chrome's V8 engine",
			DocumentationURL: "https://nodejs.org/api/cli.html#node_no_telemetry1",
			DataCollection:   "Anonymous runtime and CLI telemetry in Node features that support it.",
			Detect:           []string{"node"},
			Category:         "Runtimes & SDKs",
			// NODE_NO_TELEMETRY is the switch Node documents.
			// DISABLE_TELEMETRY is a generic name used by other libraries. Keeping
			// it on this entry made every machine with `node` on PATH fail the
			// Node audit unless that unrelated variable was also set.
			EnvChecks: []EnvCheck{
				{
					Name:        "NODE_NO_TELEMETRY",
					ValidValues: []string{"1"},
					Description: "Disables Node.js built-in telemetry (Node 21+)",
				},
			},
		},
	}
}
