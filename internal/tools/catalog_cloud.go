package tools

// catalogCloud returns the Cloud & Deployment tools.
// Registry concatenates these functions in CategoryOrder. Keep each function's
// tools sorted by name. Add a new category by adding a function and an entry
// in CategoryOrder, rather than relying on alphabetical order.
func catalogCloud() []Tool {
	return []Tool{
		{
			Name:             "Algolia CLI",
			Description:      "Algolia command-line interface for search platform management",
			DocumentationURL: "https://www.algolia.com/doc/tools/cli/telemetry",
			DataCollection:   "Command usage, operating system details, CLI version, and local profile metadata.",
			Detect:           []string{"algolia"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				{
					Name:        "ALGOLIA_CLI_TELEMETRY",
					ValidValues: []string{"0"},
					Description: "Disables Algolia CLI telemetry collection",
				},
			},
		},
		{
			Name:             "Azure CLI",
			Description:      "Microsoft Azure command-line interface",
			DocumentationURL: "https://learn.microsoft.com/en-us/cli/azure/azure-cli-configuration?view=azure-cli-latest#cli-configuration-values-and-environment-variables",
			DataCollection:   "Command usage, performance metrics, and error-rate telemetry.",
			Detect:           []string{"az"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"AZURE_CORE_COLLECT_TELEMETRY",
					"Disables Azure CLI telemetry collection",
					"0", "false",
				),
			},
			Commands: []string{"az config set core.collect_telemetry=no"},
		},
		{
			Name:             "Cloudflare Wrangler CLI",
			Description:      "Cloudflare CLI for building and deploying Workers and Pages applications",
			DocumentationURL: "https://github.com/cloudflare/workers-sdk/blob/main/packages/wrangler/telemetry.md#wrangler-cli-telemetry",
			DataCollection:   "Command names, anonymized arguments and flags, anonymized project metadata, API timing, CLI and runtime details, session duration, CI or interactive usage, AI agent usage, sanitized error details, and general machine information.",
			Detect:           []string{"wrangler"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"WRANGLER_SEND_METRICS",
					"Disables Cloudflare Wrangler CLI telemetry",
					"false",
				),
			},
			Commands: []string{"wrangler telemetry disable"},
		},
		{
			Name:             "CockroachDB",
			Description:      "Distributed SQL database and CockroachDB command-line tools",
			DocumentationURL: "https://www.cockroachlabs.com/docs/stable/telemetry",
			DataCollection:   "Cluster diagnostics, telemetry, and crash reports sent to Cockroach Labs.",
			Detect:           []string{"cockroach"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING",
					// The variable is read at node start. The README used to carry this
					// constraint as a hand-written note, which is how it drifted
					// away from the catalog. `notel list` prints this description.
					"Disables CockroachDB diagnostic reporting when set before the first node of a cluster is initialised. After the cluster is running, change the diagnostics.reporting.enabled cluster setting. Telemetry stays on during the 30-day self-service Enterprise Trial.",
					"true", "1",
				),
			},
		},
		{
			Name:             "Docker Agent",
			Description:      "Docker CLI agent for running AI agents from Docker",
			DocumentationURL: "https://docs.docker.com/ai/docker-agent/community/telemetry/",
			DataCollection:   "Command names, agent and model types, tool-call outcomes, token counts, estimated costs, and session metadata.",
			// The installed executable is docker-agent. Homebrew's docker-agent
			// formula (formerly cagent) and the release archives both put that
			// name on PATH. `docker` is the Docker CLI: LookPath("docker")
			// would mark every Docker installation as Docker Agent and then
			// fail the generic TELEMETRY_ENABLED check. A copy that exists
			// only under ~/.docker/cli-plugins is invisible here, the same
			// PATH limit as every other tool. notel does not run
			// `docker agent version` to probe for the subcommand.
			Detect:   []string{"docker-agent"},
			Category: "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"TELEMETRY_ENABLED",
					"Disables Docker Agent telemetry",
					"false",
				),
			},
		},
		{
			Name:             "GitHub CLI",
			Description:      "Official command-line tool for GitHub",
			DocumentationURL: "https://docs.github.com/en/github-cli/github-cli/github-cli-telemetry",
			DataCollection:   "Command usage, feature usage, and diagnostic telemetry from the GitHub CLI.",
			Detect:           []string{"gh"},
			Category:         "Cloud & Deployment",
			// cli/cli reads GH_TELEMETRY before DO_NOT_TRACK. A present
			// GH_TELEMETRY that is not an opt-out enables upload and never
			// consults DO_NOT_TRACK, so AnyOf would report a false pass.
			// Blank, after trim, is an opt-out. "log" does not upload, but
			// it is a debug mode rather than an opt-out, so it stays a failure.
			OptOut:    FirstSet,
			EnvChecks: githubCLIEnvChecks(),
			Commands:  []string{"gh config set telemetry disabled"},
		},
		{
			Name:             "Google Cloud SDK",
			Description:      "Google Cloud command-line tools",
			DocumentationURL: "https://cloud.google.com/sdk/docs/usage-statistics",
			DataCollection:   "Anonymized command execution metrics, timing, and error status.",
			Detect:           []string{"gcloud"},
			Category:         "Cloud & Deployment",
			// A default install does not collect usage statistics. The opt-in is
			// the active config property core/disable_usage_reporting, which
			// notel does not read. This variable overrides that property only
			// when it is set, so an opted-in install with the variable unset
			// still passes. A set value that is not a recognized opt-out fails.
			Default: DefaultDisabled,
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"CLOUDSDK_CORE_DISABLE_USAGE_REPORTING",
					"Disables Google Cloud SDK usage reporting",
					"true", "1",
				),
			},
			Commands: []string{"gcloud config set disable_usage_reporting true"},
		},
		{
			Name:             "Hasura CLI",
			Description:      "Hasura command-line interface for managing GraphQL Engine projects",
			DocumentationURL: "https://hasura.io/docs/2.0/policies/telemetry/#cli",
			DataCollection:   "Command name, timestamp, error status, server version and UUID, operating system platform and architecture, and CLI version.",
			Detect:           []string{"hasura"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"HASURA_GRAPHQL_ENABLE_TELEMETRY",
					"Disables Hasura CLI telemetry",
					"false",
				),
			},
		},
		{
			Name:             "Hugging Face CLI",
			Description:      "Hugging Face command-line interface for interacting with the Hub",
			DocumentationURL: "https://huggingface.co/docs/huggingface_hub/package_reference/environment_variables#hfhubdisabletelemetry",
			DataCollection:   "Usage data collected by Hugging Face libraries to monitor usage, debug issues, and help prioritize features.",
			Detect:           []string{"hf"},
			Category:         "Cloud & Deployment",
			// The Hub client honors either the global convention or its own variable.
			OptOut: AnyOf,
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"DO_NOT_TRACK",
					"Disables Hugging Face telemetry via the global do-not-track convention",
					"1", "true",
				),
				CaseInsensitive(
					"HF_HUB_DISABLE_TELEMETRY",
					"Disables Hugging Face telemetry",
					"1", "true",
				),
			},
		},
		{
			Name:             "Netlify CLI",
			Description:      "Netlify platform CLI",
			DocumentationURL: "https://docs.netlify.com/api-and-cli-guides/cli-guides/get-started-with-cli/#usage-data-collection",
			DataCollection:   "Anonymous CLI usage and diagnostic telemetry.",
			Detect:           []string{"netlify"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"NETLIFY_TELEMETRY_DISABLED",
					"Disables Netlify CLI telemetry",
					"1", "true",
				),
			},
			Commands: []string{"netlify --telemetry-disable"},
		},
		{
			Name:             "Railway CLI",
			Description:      "Railway command-line interface for deploying and managing projects",
			DocumentationURL: "https://docs.railway.com/cli/telemetry",
			DataCollection:   "Command and subcommand names, duration, success, truncated error message, OS, architecture, CLI version, and whether the command ran in CI.",
			Detect:           []string{"railway"},
			Category:         "Cloud & Deployment",
			OptOut:           AnyOf,
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"DO_NOT_TRACK",
					"Disables Railway CLI telemetry via the global do-not-track convention",
					"1", "true",
				),
				CaseInsensitive(
					"RAILWAY_NO_TELEMETRY",
					"Disables Railway CLI telemetry",
					"1", "true",
				),
			},
		},
		{
			Name:             "Redocly CLI",
			Description:      "Redocly command-line interface for API governance and documentation workflows",
			DocumentationURL: "https://redocly.com/docs/realm/reunite/project/telemetry",
			DataCollection:   "Anonymous usage information about Redocly project feature usage, with additional filtering to avoid sending secrets and other sensitive data.",
			Detect:           []string{"redocly"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"REDOCLY_TELEMETRY",
					"Disables Redocly CLI telemetry",
					"off",
				),
			},
		},
		{
			Name:             "Stripe CLI",
			Description:      "Stripe command-line interface for developer workflows and API interactions",
			DocumentationURL: "https://docs.stripe.com/cli/telemetry",
			DataCollection:   "Command usage patterns, error rates and types, performance metrics, operating system details, CLI version information, and detected AI coding agent names.",
			Detect:           []string{"stripe"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"STRIPE_CLI_TELEMETRY_OPTOUT",
					"Disables Stripe CLI telemetry",
					"1", "true",
				),
			},
		},
		{
			Name:             "Turborepo",
			Description:      "High-performance build system for JavaScript/TypeScript",
			DocumentationURL: "https://turborepo.com/docs/telemetry",
			DataCollection:   "Anonymous command usage, host information, and repo or task metrics.",
			Detect:           []string{"turbo"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"TURBO_TELEMETRY_DISABLED",
					"Disables Turborepo telemetry",
					"1", "true",
				),
			},
			Commands: []string{"turbo telemetry disable"},
		},
		{
			Name:             "Vercel CLI",
			Description:      "Frontend cloud platform CLI for deployment and project management",
			DocumentationURL: "https://vercel.com/docs/cli/about-telemetry",
			DataCollection:   "Command usage, arguments, CLI version, and general machine information.",
			Detect:           []string{"vercel"},
			Category:         "Cloud & Deployment",
			EnvChecks: []EnvCheck{
				CaseInsensitive(
					"VERCEL_TELEMETRY_DISABLED",
					"Disables Vercel CLI telemetry",
					"1", "true",
				),
			},
			Commands: []string{"vercel telemetry disable"},
		},
	}
}

// githubCLIEnvChecks returns GH_TELEMETRY, then DO_NOT_TRACK.
//
// Order is the precedence. TrimSpace is set because cli/cli trims and
// lowercases both variables, and a present blank GH_TELEMETRY is falsey.
// The blank spelling is "" so --fix suggests "false", the first value.
func githubCLIEnvChecks() []EnvCheck {
	telemetry := CaseInsensitive(
		"GH_TELEMETRY",
		"Disables GitHub CLI telemetry when set. false, 0, no, disabled, off, and a blank value opt out. Any other set value, including log, enables or logs telemetry and overrides DO_NOT_TRACK.",
		"false", "0", "no", "disabled", "off", "",
	)
	telemetry.TrimSpace = true
	doNotTrack := CaseInsensitive(
		"DO_NOT_TRACK",
		"Disables GitHub CLI telemetry when GH_TELEMETRY is absent",
		"1", "true",
	)
	doNotTrack.TrimSpace = true
	return []EnvCheck{telemetry, doNotTrack}
}
