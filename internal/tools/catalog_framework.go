package tools

// catalogFramework returns the Framework tools.
// Registry concatenates these functions in CategoryOrder. Keep each function's
// tools sorted by name. Add a new category by adding a function and an entry
// in CategoryOrder, rather than relying on alphabetical order.
func catalogFramework() []Tool {
	return []Tool{
		{
			Name:             "Angular CLI",
			Description:      "Angular framework CLI",
			DocumentationURL: "https://angular.dev/cli/analytics",
			DataCollection:   "Command usage, selected flags, workspace shape, and local version metadata.",
			Detect:           []string{"ng"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"NG_CLI_ANALYTICS",
					"Disables Angular CLI analytics (set to false or ci)",
					"false", "ci", "0",
				),
			},
			Commands: []string{"ng analytics disable"},
		},
		{
			Name:             "Astro",
			Description:      "The web framework for content-driven websites",
			DocumentationURL: "https://docs.astro.build/en/reference/cli-reference/#astro-telemetry",
			DataCollection:   "Anonymous command usage, integration usage, and project metadata.",
			Detect:           []string{"astro"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"ASTRO_TELEMETRY_DISABLED",
					"Disables Astro telemetry",
					"1", "true",
				),
			},
			Commands: []string{"astro telemetry disable"},
		},
		{
			Name:             "Expo CLI",
			Description:      "Expo command-line tools for React Native and universal apps",
			DocumentationURL: "https://docs.expo.dev/more/expo-cli/#telemetry",
			DataCollection:   "Anonymous CLI usage and diagnostics for Expo development workflows.",
			Detect:           []string{"expo"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"EXPO_NO_TELEMETRY",
					"Disables Expo CLI telemetry",
					"1", "true",
				),
			},
		},
		{
			Name:             "Gatsby",
			Description:      "React-based static site generator",
			DocumentationURL: "https://www.gatsbyjs.com/docs/telemetry/",
			DataCollection:   "Anonymous command usage, plugin usage, and machine characteristics.",
			Detect:           []string{"gatsby"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"GATSBY_TELEMETRY_DISABLED",
					"Disables Gatsby telemetry",
					"1", "true",
				),
			},
			Commands: []string{"gatsby telemetry --disable"},
		},
		{
			Name:             "Next.js",
			Description:      "The React framework for production",
			DocumentationURL: "https://nextjs.org/telemetry",
			DataCollection:   "Anonymous command usage, session timing, and project or machine characteristics.",
			Detect:           []string{"next"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				{
					// Next.js documents the exact value "1". Other spellings
					// are not accepted, so this check stays case-sensitive
					// and does not list "true".
					Name:        "NEXT_TELEMETRY_DISABLED",
					ValidValues: []string{"1"},
					Description: "Disables Next.js telemetry collection",
				},
			},
			Commands: []string{"next telemetry disable"},
		},
		{
			Name:             "Nuxt",
			Description:      "The intuitive Vue framework",
			DocumentationURL: "https://github.com/nuxt/telemetry#nuxt-telemetry-module",
			DataCollection:   "Anonymous command usage, module usage, and environment characteristics.",
			Detect:           []string{"nuxt"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"NUXT_TELEMETRY_DISABLED",
					"Disables Nuxt telemetry collection",
					"1", "true",
				),
			},
		},
		{
			Name:             "Storybook",
			Description:      "Frontend workshop for building UI components and pages in isolation",
			DocumentationURL: "https://storybook.js.org/docs/configure/telemetry",
			DataCollection:   "Anonymous framework usage, addon data, and environment metadata.",
			Detect:           []string{"storybook"},
			Category:         "Framework",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"STORYBOOK_DISABLE_TELEMETRY",
					"Disables Storybook telemetry",
					"true",
				),
			},
			Commands: []string{"npm run storybook -- --disable-telemetry"},
		},
	}
}
