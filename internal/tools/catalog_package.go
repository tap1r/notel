package tools

// catalogPackage returns the Package Manager tools.
// Registry concatenates these functions in CategoryOrder. Keep each function's
// tools sorted by name. Add a new category by adding a function and an entry
// in CategoryOrder, rather than relying on alphabetical order.
func catalogPackage() []Tool {
	return []Tool{
		{
			Name:             "CocoaPods",
			Description:      "Dependency manager for Swift and Objective-C projects",
			DocumentationURL: "https://blog.cocoapods.org/Stats/",
			DataCollection:   "Anonymous pod install statistics, including hashed project target identifiers, target product types, pod names, and CocoaPods version information.",
			Detect:           []string{"pod"},
			Category:         "Package Manager",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"COCOAPODS_DISABLE_STATS",
					"Disables CocoaPods stats collection",
					"true", "1",
				),
			},
		},
		{
			Name:             "Homebrew",
			Description:      "The missing package manager for macOS",
			DocumentationURL: "https://docs.brew.sh/Analytics",
			DataCollection:   "Anonymous install events, command usage, and build error metadata.",
			Detect:           []string{"brew"},
			Category:         "Package Manager",
			EnvChecks: []EnvCheck{
				{
					// Homebrew documents the exact value "1".
					Name:        "HOMEBREW_NO_ANALYTICS",
					ValidValues: []string{"1"},
					Description: "Disables Homebrew analytics reporting",
				},
			},
			Commands: []string{"brew analytics off"},
		},
		{
			Name:             "Poetry",
			Description:      "Python dependency management and packaging tool",
			DocumentationURL: "https://python-poetry.org/docs/configuration/#using-environment-variables",
			DataCollection:   "Anonymous usage statistics and telemetry about Poetry command execution.",
			Detect:           []string{"poetry"},
			Category:         "Package Manager",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"POETRY_TELEMETRY_ENABLED",
					"Disables Poetry telemetry collection",
					"0", "false",
				),
			},
		},
		{
			Name:             "Yarn",
			Description:      "Fast, reliable JavaScript package manager",
			DocumentationURL: "https://yarnpkg.com/advanced/telemetry",
			DataCollection:   "Anonymous command usage and version information.",
			Detect:           []string{"yarn"},
			Category:         "Package Manager",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"YARN_ENABLE_TELEMETRY",
					"Disables Yarn telemetry (Yarn 2+)",
					"0", "false",
				),
			},
		},
	}
}
