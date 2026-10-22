# notel

`notel` helps you audit and update telemetry settings across your CLI dev tools.

- [Overview](#overview)
- [Requirements](#requirements)
- [Install](#install)
- [Usage](#usage)
- [Example Output](#example-output)
- [Supported Tools by Category](#supported-tools-by-category)
- [Build](#build)
- [Feedback or Suggestions](#feedback-or-suggestions)

### Overview

Many popular developer tools collect telemetry by default. This can have
benefits for both end users and vendors, but can also be a legitimate
privacy concern. I've found it frustrating to have to individually configure
each tool to disable telemetry and keep track of which ones have telemetry
enabled by default, so I built this tool to:

 - Audit which installed dev tool CLIs have telemetry enabled
 - Check if you've set the right opt-out environment variables
 - Suggest the exact commands to disable telemetry (export env vars for
   shells or relevant CLI commands)

The initial design goal focuses on environment variable configuration to
keep the scope maintainable. `notel` does not make any changes to your
environment or run any CLI commands.

### How a tool passes

Each registry entry declares the rule the audit applies:

- `OptOut` defaults to `AllOf`. Every environment variable has to match.
  Use this when the variables are independent switches.
- `OptOut: AnyOf` passes when one variable matches. Hugging Face CLI and
  Railway CLI use this because `DO_NOT_TRACK` is an alternative to the
  tool-specific variable, not a second requirement.
- `OptOut: FirstSet` uses the first variable that is present. GitHub CLI
  uses this because a set `GH_TELEMETRY` decides, and `DO_NOT_TRACK` applies
  only when `GH_TELEMETRY` is absent. `GH_TELEMETRY=1` with `DO_NOT_TRACK=1`
  still fails.
- `Default: DefaultDisabled` means an unset variable is already opted out.
  Google Cloud SDK uses this because a default install does not collect
  usage statistics. The install-time choice is stored in gcloud's active
  configuration, which notel does not read. A value that is set to something
  notel does not recognize still fails.
- `IgnoreCase` accepts any capitalization of the listed spellings. Tools that
  document one exact value, such as Homebrew's `1` and Next.js's `1`, stay
  case-sensitive.
- `Commands` lists CLI hints such as `brew analytics off` and
  `flutter config --no-analytics`. notel prints them in `list` and
  `audit --verbose`. It does not run them, and a passing environment check
  does not prove the command has been run.

### Why disable telemetry?

- Disabling telemetry exercises your rights to privacy and control over your
  data and development environments.
- Poorly considered telemetry collection can potentially impact performance or
  resources.
- Telemetry collection may leak unexpected details (internal project names,
  dependencies, or build configurations) from private build or CI environments.
- Vendors aren't always up front about telemetry collection practices and
  details.

### Why enable telemetry?

- Telemetry signals can provide insight into tool adoption, usage patterns, and
  crash information that will help vendors improve the product experience.
- Aggregated telemetry signals can help product users understand usage of
  packages, features, and versions.
- Well-implemented telemetry can have negligible performance impact and can be
  configured to exclude CI environments.

### DO_NOT_TRACK

There is a proposed [`DO_NOT_TRACK`](https://donottrack.sh/) standard which
aims to unambiguously express a user's intent to opt out of:

 - Ad tracking
 - Usage reporting (anonymous or otherwise)
 - Telemetry
 - Crash reporting
 - Non-essential-to-functionality requests

However, this hasn't been widely adopted yet.

### Related projects

Some interesting open source projects I found on GitHub after scratching my own
itch to audit telemetry tools in my environments:

 - [`alloydwhitlock/do-not-track-cli`](https://github.com/alloydwhitlock/do-not-track-cli):
   A single `.env` file to opt out of telemetry across CLI tools, frameworks,
   SDKs, and runtimes.
 - [`beatcracker/toptout`](https://github.com/beatcracker/toptout`): This
   project collects data about telemetry in applications in machine-readable
   format and makes it easy to create derivative works such as awesome-lists,
   configuration scripts, APIs, etc.

### Related discussions

 - [Hacker News (3 May 2026): donottrack.sh](https://news.ycombinator.com/item?id=47988592)
 - [Go Telemetry (v1.23+)](https://go.dev/doc/telemetry): By default, telemetry data is
   kept only on the local computer, but users may opt in to uploading an
   approved subset of telemetry data to telemetry.go.dev.

## Requirements

- Go 1.26.3+
  - [Cobra](https://github.com/spf13/cobra) — CLI framework
  - [Lipgloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [Just](https://github.com/casey/just) — Task runner

## Install

Install from the Homebrew tap:

```sh
brew install stennie/notel/notel
```

Or tap the repository first and then install `notel`:

```sh
brew tap stennie/notel
brew install notel
```

If you manage tools with a Brewfile:

```ruby
tap "stennie/notel"
brew "notel"
```

The Homebrew formula lives in
[`stennie/homebrew-notel`](https://github.com/stennie/homebrew-notel) and
installs release artifacts published from this repository.

## Usage

```
notel [command]
```

### Commands

| Command | Description |
|---------|-------------|
| `notel audit` | Show telemetry status for all **installed** tools |
| `notel audit --all` | Show status for all tools (including those not installed) |
| `notel audit --verbose` | Detailed audit with fix hints for all **installed** tools |
| `notel audit --verbose --all` | Detailed audit for all tools (including those not installed) |
| `notel audit --fix` / `notel audit -f` | Print shell commands (using `stdout`) to disable telemetry for the current detected shell. Diagnostic messages and errors use `stderr` so the output is cleanly redirectable.|
| `notel audit --fix=zsh` / `notel audit -f=zsh` | Print shell commands for a specific shell (`bash`, `fish`, `powershell`, `zsh`) |
| `notel audit --fix --profile` | Print every visible opt-out, including values that already match. Shared variables are printed once. |
| `notel list` / `notel ls` | List all supported tools and their opt-out variables |
| `notel list --format=markdown` | Print the catalog as Markdown tables. This is the text between the markers in the supported-tools section. |
| `notel version` / `notel --version` | Show the application version |
| `notel help` | Show help |

When `--fix` is combined with `--all` and/or `--verbose`, the audit report is
still shown on `stderr` and the shell commands remain redirect-safe on `stdout`.

`--fix` prints a delta: the assignments still required for visible tools
whose opt-out policy is not satisfied, plus a comment for each unverified
CLI hint on those tools. A tool that is not on PATH is included only with
`--all`, and even then it is skipped when its variables already match.
`--fix --profile` prints the full set instead, including variables that
already match. Either way a variable shared by several tools, such as
`DO_NOT_TRACK`, is printed once. If two tools want that variable set to
different values, the first value is kept and the disagreement is written to
`stderr`. An empty delta says whether nothing is installed or everything
visible already passes. Those are different situations.

### Example Output

<table>
  <tr>
    <th>Command</th>
    <th>Screenshot</th>
  </tr>
  <tr>
    <td><code>notel audit</code></td>
    <td><img src="screenshots/notel-audit.png" alt="Screenshot: notel audit" width="500"></td>
  </tr>
  <tr>
    <td><code>notel audit --fix</code></td>
    <td><img src="screenshots/notel-audit-fix.png" alt="Screenshot: notel audit" width="500"></td>
  </tr>
  <tr>
    <td><code>notel audit --verbose</code></td>
    <td><img src="screenshots/notel-audit-verbose.png" alt="Screenshot: notel audit" width="500"></td>
  </tr>
</table>


## Supported Tools by Category

These tables are generated from the registry. Edit the catalog file for
the category, then replace the block between the markers with
`go run . list --format=markdown`. `go test` fails if the block drifts.

<!-- notel:catalog:start -->
### All The Things

| Tool | Environment Variable | Opt-out Value | Data Collection |
| --- | --- | --- | --- |
| [Do Not Track](https://donottrack.sh) | `DO_NOT_TRACK` | `1` or `true` (any case) | Signals a global preference to disable telemetry, analytics, crash reporting, and non-essential tracking. |


### Cloud & Deployment

| Tool | Environment Variable | Opt-out Value | Data Collection |
| --- | --- | --- | --- |
| [Algolia CLI](https://www.algolia.com/doc/tools/cli/telemetry) | `ALGOLIA_CLI_TELEMETRY` | `0` | Command usage, operating system details, CLI version, and local profile metadata. |
| [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/azure-cli-configuration?view=azure-cli-latest#cli-configuration-values-and-environment-variables) | `AZURE_CORE_COLLECT_TELEMETRY` | `0` or `false` (any case) | Command usage, performance metrics, and error-rate telemetry.<br>`az config set core.collect_telemetry=no` |
| [Cloudflare Wrangler CLI](https://github.com/cloudflare/workers-sdk/blob/main/packages/wrangler/telemetry.md#wrangler-cli-telemetry) | `WRANGLER_SEND_METRICS` | `false` (any case) | Command names, anonymized arguments and flags, anonymized project metadata, API timing, CLI and runtime details, session duration, CI or interactive usage, AI agent usage, sanitized error details, and general machine information.<br>`wrangler telemetry disable` |
| [CockroachDB](https://www.cockroachlabs.com/docs/stable/telemetry) | `COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING` | `true` or `1` (any case) | Cluster diagnostics, telemetry, and crash reports sent to Cockroach Labs. |
| [Docker Agent](https://docs.docker.com/ai/docker-agent/community/telemetry/) | `TELEMETRY_ENABLED` | `false` (any case) | Command names, agent and model types, tool-call outcomes, token counts, estimated costs, and session metadata. |
| [GitHub CLI](https://docs.github.com/en/github-cli/github-cli/github-cli-telemetry) | `GH_TELEMETRY`<br>`DO_NOT_TRACK` | `false` or `0` or `no` or `disabled` or `off` or `(empty)` (any case)<br>`1` or `true` (any case) | Command usage, feature usage, and diagnostic telemetry from the GitHub CLI.<br>The first variable that is set decides. Later variables apply only when it is absent.<br>`gh config set telemetry disabled` |
| [Google Cloud SDK](https://cloud.google.com/sdk/docs/usage-statistics) | `CLOUDSDK_CORE_DISABLE_USAGE_REPORTING` | `true` or `1` (any case) | Anonymized command execution metrics, timing, and error status.<br>Unset passes. The vendor documents collection as opt-in, and notel does not read config files.<br>`gcloud config set disable_usage_reporting true` |
| [Hasura CLI](https://hasura.io/docs/2.0/policies/telemetry/#cli) | `HASURA_GRAPHQL_ENABLE_TELEMETRY` | `false` (any case) | Command name, timestamp, error status, server version and UUID, operating system platform and architecture, and CLI version. |
| [Hugging Face CLI](https://huggingface.co/docs/huggingface_hub/package_reference/environment_variables#hfhubdisabletelemetry) | `DO_NOT_TRACK` / `HF_HUB_DISABLE_TELEMETRY` | `1` or `true` (any case) / `1` or `true` (any case) | Usage data collected by Hugging Face libraries to monitor usage, debug issues, and help prioritize features.<br>Any one of the variables opts out. |
| [Netlify CLI](https://docs.netlify.com/api-and-cli-guides/cli-guides/get-started-with-cli/#usage-data-collection) | `NETLIFY_TELEMETRY_DISABLED` | `1` or `true` (any case) | Anonymous CLI usage and diagnostic telemetry.<br>`netlify --telemetry-disable` |
| [Railway CLI](https://docs.railway.com/cli/telemetry) | `DO_NOT_TRACK` / `RAILWAY_NO_TELEMETRY` | `1` or `true` (any case) / `1` or `true` (any case) | Command and subcommand names, duration, success, truncated error message, OS, architecture, CLI version, and whether the command ran in CI.<br>Any one of the variables opts out. |
| [Redocly CLI](https://redocly.com/docs/realm/reunite/project/telemetry) | `REDOCLY_TELEMETRY` | `off` (any case) | Anonymous usage information about Redocly project feature usage, with additional filtering to avoid sending secrets and other sensitive data. |
| [Stripe CLI](https://docs.stripe.com/cli/telemetry) | `STRIPE_CLI_TELEMETRY_OPTOUT` | `1` or `true` (any case) | Command usage patterns, error rates and types, performance metrics, operating system details, CLI version information, and detected AI coding agent names. |
| [Turborepo](https://turborepo.com/docs/telemetry) | `TURBO_TELEMETRY_DISABLED` | `1` or `true` (any case) | Anonymous command usage, host information, and repo or task metrics.<br>`turbo telemetry disable` |
| [Vercel CLI](https://vercel.com/docs/cli/about-telemetry) | `VERCEL_TELEMETRY_DISABLED` | `1` or `true` (any case) | Command usage, arguments, CLI version, and general machine information.<br>`vercel telemetry disable` |


### Framework

| Tool | Environment Variable | Opt-out Value | Data Collection |
| --- | --- | --- | --- |
| [Angular CLI](https://angular.dev/cli/analytics) | `NG_CLI_ANALYTICS` | `false` or `ci` or `0` (any case) | Command usage, selected flags, workspace shape, and local version metadata.<br>`ng analytics disable` |
| [Astro](https://docs.astro.build/en/reference/cli-reference/#astro-telemetry) | `ASTRO_TELEMETRY_DISABLED` | `1` or `true` (any case) | Anonymous command usage, integration usage, and project metadata.<br>`astro telemetry disable` |
| [Expo CLI](https://docs.expo.dev/more/expo-cli/#telemetry) | `EXPO_NO_TELEMETRY` | `1` or `true` (any case) | Anonymous CLI usage and diagnostics for Expo development workflows. |
| [Gatsby](https://www.gatsbyjs.com/docs/telemetry/) | `GATSBY_TELEMETRY_DISABLED` | `1` or `true` (any case) | Anonymous command usage, plugin usage, and machine characteristics.<br>`gatsby telemetry --disable` |
| [Next.js](https://nextjs.org/telemetry) | `NEXT_TELEMETRY_DISABLED` | `1` | Anonymous command usage, session timing, and project or machine characteristics.<br>`next telemetry disable` |
| [Nuxt](https://github.com/nuxt/telemetry#nuxt-telemetry-module) | `NUXT_TELEMETRY_DISABLED` | `1` or `true` (any case) | Anonymous command usage, module usage, and environment characteristics. |
| [Storybook](https://storybook.js.org/docs/configure/telemetry) | `STORYBOOK_DISABLE_TELEMETRY` | `true` (any case) | Anonymous framework usage, addon data, and environment metadata.<br>`npm run storybook -- --disable-telemetry` |


### Package Manager

| Tool | Environment Variable | Opt-out Value | Data Collection |
| --- | --- | --- | --- |
| [CocoaPods](https://blog.cocoapods.org/Stats/) | `COCOAPODS_DISABLE_STATS` | `true` or `1` (any case) | Anonymous pod install statistics, including hashed project target identifiers, target product types, pod names, and CocoaPods version information. |
| [Homebrew](https://docs.brew.sh/Analytics) | `HOMEBREW_NO_ANALYTICS` | `1` | Anonymous install events, command usage, and build error metadata.<br>`brew analytics off` |
| [Poetry](https://python-poetry.org/docs/configuration/#using-environment-variables) | `POETRY_TELEMETRY_ENABLED` | `0` or `false` (any case) | Anonymous usage statistics and telemetry about Poetry command execution. |
| [Yarn](https://yarnpkg.com/advanced/telemetry) | `YARN_ENABLE_TELEMETRY` | `0` or `false` (any case) | Anonymous command usage and version information. |


### Runtimes & SDKs

| Tool | Environment Variable | Opt-out Value | Data Collection |
| --- | --- | --- | --- |
| [.NET SDK](https://learn.microsoft.com/en-us/dotnet/core/tools/telemetry) | `DOTNET_CLI_TELEMETRY_OPTOUT` | `1` or `true` (any case) | CLI command usage, SDK versions, and exception metadata. |
| [Bun](https://bun.sh/docs/runtime/bunfig#telemetry) | `DO_NOT_TRACK` | `1` or `true` (any case) | Anonymous telemetry and crash-reporting signals from Bun tooling. |
| [Flutter](https://docs.flutter.dev/reference/crash-reporting) | `FLUTTER_CLI_CRASH_REPORTING` | `false` or `0` (any case) | CLI crash reports, tool usage signals, and local environment metadata.<br>`flutter config --no-analytics` |
| [Node.js](https://nodejs.org/api/cli.html#node_no_telemetry1) | `NODE_NO_TELEMETRY` | `1` | Anonymous runtime and CLI telemetry in Node features that support it. |


### Security & Observability

| Tool | Environment Variable | Opt-out Value | Data Collection |
| --- | --- | --- | --- |
| [Semgrep](https://semgrep.dev/docs/metrics) | `SEMGREP_SEND_METRICS` | `off` | Usage metrics about Semgrep runs, registry usage, and login-related activity.<br>`semgrep --metrics=off` |
| [Sentry CLI](https://cli.sentry.dev/configuration/) | `SENTRY_CLI_NO_TELEMETRY` | `1` or `true` (any case) | Anonymous CLI usage and diagnostic telemetry. |
<!-- notel:catalog:end -->

## Build

- Local build: `just build` -> `bin/notel`
- Cross-platform release artifacts: `just release` -> `dist/`
- Full release preflight plus artifacts: `just release-check`
- Release packaging is shared by local and CI builds via `scripts/release.sh`

`just release` builds archives for:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`
- `windows/arm64`

Release artifacts follow the convention `notel_<version>_<os>_<arch>.tar.gz`
or `.zip`, and `dist/SHA256SUMS` is generated for the packaged archives.

## Adding a new tool

Add an entry to the catalog file for its category, for example
`internal/tools/catalog_framework.go`. `Registry` concatenates those files
in `CategoryOrder`. A new category needs a function and a line in
`CategoryOrder`.

Then regenerate the supported-tools section:

```sh
go run . list --format=markdown
```

Replace the block between `<!-- notel:catalog:start -->` and
`<!-- notel:catalog:end -->`. `go test` fails if that block drifts from the
registry.

```go
{
    Name:        "My Tool",
    Description: "Description of the tool",
    DocumentationURL: "https://example.com/docs/telemetry",
    DataCollection:   "Anonymous command usage and diagnostic metadata.",
    Detect:           []string{"mytool"}, // PATH names; any one match counts as installed
    // AlwaysInstalled: true,          // env-only conventions such as DO_NOT_TRACK
    Category:         "Framework",      // must match CategoryOrder and the file you edited
    // OptOut: AnyOf,                 // set only when one variable is enough
    // Default: DefaultDisabled,      // set only when collection is opt-in
    EnvChecks: []EnvCheck{
        CaseInsensitive(
            "MYTOOL_TELEMETRY_DISABLED",
            "Disables My Tool telemetry",
            "1", "true", // first value is what --fix suggests
        ),
    },
    Commands: []string{"mytool telemetry disable"}, // printed, never executed
},
```

`CaseInsensitive` lists each canonical spelling once. Use an `EnvCheck`
literal with `IgnoreCase: false` when the vendor accepts only one exact
value. Leave `OptOut` and `Default` at their zero values unless the tool's
documentation says otherwise.

`Detect` is the list of PATH names that identify the product. An empty list
does not mean the tool is installed. Set `AlwaysInstalled` for an environment
convention that has no executable. Leave both empty when no PATH name
identifies the product without also matching a different one. Docker Agent's
name is `docker-agent`; `docker` is the Docker CLI and is not evidence that
the agent is installed.

Keep tools sorted by name inside each catalog file. Group order comes from
`CategoryOrder`, not from alphabetical sorting of the category name.

## Project Structure

```
notel/
├── main.go
├── cmd/
│   ├── root.go      # Cobra root command
│   ├── audit.go     # notel audit
│   └── list.go      # notel list
└── internal/
    ├── tools/
    │   ├── registry.go          # concatenates the catalog files
    │   ├── category.go          # CategoryOrder
    │   ├── markdown.go          # README catalog tables
    │   └── catalog_*.go         # one file per category (add tools here)
    ├── check/
    │   └── checker.go    # Detection & env-var audit logic
    └── output/
        └── formatter.go  # Lipgloss-styled output
```

## License
Licensed under the Apache License, Version 2.0. See `LICENSE`.

## Feedback or Suggestions?

Bug reports, feature requests, and questions can be posted in the [Issues](https://github.com/stennie/notel/issues) section on GitHub.
