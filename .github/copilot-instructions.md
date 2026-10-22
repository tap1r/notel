# Copilot instructions for `notel`

## Build and test commands

- Build the CLI binary from source: `mkdir -p bin && go build -o bin/notel .`
- Preferred release/local build with embedded version: `mkdir -p bin && go build -ldflags "-X github.com/stennie/notel/cmd.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o bin/notel .`
- Build cross-platform release artifacts into `dist/`: `just release`
- Run release preflight plus artifact builds: `just release-check`
- Shared release packaging lives in `scripts/release.sh`; keep `Justfile` and `.github/workflows/release.yml` using that script so targets and archive contents do not drift.
- Compile-check the module: `go build ./...`
- Run the full test suite: `go test ./...`
- Run a single test by name once `*_test.go` files exist: `go test ./... -run '^TestName$'`
- CI enforcement lives in `.github/workflows/ci.yml` and runs `gofmt`, `go vet`, `go test ./...`, and `go build ./...` on Go 1.26.3.

## High-level architecture

- The intended product is a single-binary Go CLI built around Cobra commands and Charm-based terminal UX. Keep the app cross-platform and dependency-light: standard library plus Cobra and Charm libraries.
- `main.go` is only the entrypoint; it delegates to `cmd.Execute()`.
- The command layer is the Cobra tree in `cmd/`. Command handlers should stay thin and hand off work to the audit and output layers.
- Do not introduce package-level function variables for tests. `check.Checker` and `fix.Environment` carry the PATH and environment dependencies. `cmd.runAudit` takes an `auditOptions` value and an `auditRuntime` value. A nil function field on `Checker` falls back to the process. `LookupEnv` falls back to `os.LookupEnv`, not to `Getenv`: a TrimSpace check must tell an unset variable from one set to empty, and stubbing `Getenv` does not change `LookupEnv`. Tests that do not call `t.Setenv` can run in parallel. Cobra flag bindings are read inside `RunE` via `auditOptionsFrom`.
- The telemetry source of truth is the catalog in `internal/tools/catalog_*.go`. `Registry()` concatenates those files. Each `Tool` declares its display metadata, category, install signal, and one or more `EnvCheck` definitions.
- The audit engine in `internal/check/checker.go` turns registry entries into `ToolResult` values by:
  - resolving installation via `exec.LookPath`
  - reading environment variables with `os.Getenv`
  - marking a tool as passing only when it is installed and its `OptOut` policy is satisfied
- `Tool.OptOut` defaults to `AllOf`: every `EnvCheck` must pass. `AnyOf` passes when one check passes. Use `AnyOf` for alternative opt-outs such as `DO_NOT_TRACK` or a tool-specific variable. `FirstSet` uses the first variable that is present; GitHub CLI uses it because `GH_TELEMETRY` overrides `DO_NOT_TRACK`. Do not combine `AnyOf` or `FirstSet` with `DefaultDisabled`.
- `Tool.Default` defaults to `DefaultEnabled`: an unset variable fails. `DefaultDisabled` means an unset variable is already opted out because the vendor collects only after an explicit opt-in. Google Cloud SDK's opt-in lives in its config file, which notel does not read. A set value that is not a recognized opt-out still fails.
- `EnvCheck.TrimSpace` compares a trimmed value and uses `LookupEnv`, so a present blank value is not the same as an unset variable. `GH_TELEMETRY` sets it. A blank opt-out is `""` in `ValidValues`, listed after the spelling `--fix` should suggest.
- `EnvCheck.IgnoreCase` compares with `strings.EqualFold`. Prefer `CaseInsensitive(...)` and list canonical spellings once. Leave `IgnoreCase` false when the vendor documents one exact value.
- `Tool.Commands` are CLI hints rendered by list and verbose audit output. notel must not execute them, and they must not flip `AllPassing`. A passing env check does not prove a command has been run.
- The rendering layer in `internal/output/formatter.go` is shared by multiple commands:
  - `audit` prints a compact installed-tool view by default, switches to the detailed per-tool block when `--verbose` is set, and can print shell-specific fix commands when `--fix` is used
  - `notel audit --fix` prints a delta for visible tools whose opt-out policy is not satisfied. A missing binary is omitted unless `--all` is set, and `--all` still skips a tool whose variables already match. `notel audit --fix --profile` prints every visible opt-out. Shared variable names are emitted once. CLI hints are shell comments and are not executed.
  - `list` reuses registry-driven results so it can show the documented env vars alongside current installation status
- If interactive flows are added later, prefer Charm libraries (`bubbles`, `huh`) for those surfaces instead of introducing a separate TUI/form stack.

## Key conventions

- The original bootstrap scope was: Go modules, Cobra command structure, a simple `audit` command for environment variables, and Lipgloss output. Keep new work aligned with that layered structure rather than folding logic back into `main.go`.
- Keep the registry as the only place where supported tools are defined. Adding a tool means adding one `Tool` entry with its `EnvChecks`; command handlers should not duplicate tool metadata.
- Installation is `Tool.Detect` plus `Tool.AlwaysInstalled`.
  - `Detect` is a list of PATH executable names. The tool is installed when any name resolves. Names are alternatives for one product, not unrelated tools.
  - `AlwaysInstalled` is for env-only conventions such as Do Not Track. Set it explicitly.
  - An empty `Detect` list without `AlwaysInstalled` means the catalog has no honest PATH signal. The tool stays in `list` and `audit --all` and is omitted from the default audit.
  - Docker Agent's executable is `docker-agent`. Do not list `docker`: that is the Docker CLI, and a hit would treat every Docker install as the agent.
  - Do not attach a generic environment variable to a broadly installed runtime. `DISABLE_TELEMETRY` is not a Node.js switch; Node's check is `NODE_NO_TELEMETRY`.
- `EnvCheck.ValidValues` is authoritative:
  - if it contains values, the env var must match one of them, using `strings.EqualFold` when `IgnoreCase` is set and an exact string comparison otherwise
  - if it is empty, any non-empty value counts as opted out
  - the first value is the spelling `notel audit --fix` suggests
- The policy types live in `internal/tools/types.go`. `internal/tools/registry.go` only concatenates the catalog files. Tool entries do not live there.
- Catalog entries live in `internal/tools/catalog_*.go`, one file per category. `Registry` concatenates them. Group order is `CategoryOrder` in `internal/tools/category.go`, not alphabetical order of the category name. Within a file, keep tools sorted by name.
- The README catalog between `<!-- notel:catalog:start -->` and `<!-- notel:catalog:end -->` is generated by `notel list --format=markdown` (`tools.CatalogMarkdown`). Update the registry and regenerate the block together. The test `TestREADMECatalogMatchesRegistry` fails when they drift.
- `audit` hides uninstalled tools unless `--all` is set, whether it is in compact or `--verbose` mode. `list` operates over the full registry, so changes to filtering should be made deliberately across commands.
- Styling and output rules belong in `internal/output/formatter.go`, not in Cobra handlers. Prefer extending the output package over spreading Lipgloss formatting through command files. Render functions take an `io.Writer`. Pad plain text and then apply a Lipgloss style. Padding a string that already contains ANSI escapes counts those escapes as columns.
- The project goal is to cover common developer tools across ecosystems (package managers, runtimes, SDKs, frameworks, cloud CLIs, etc.). Prefer extending the registry in a way that keeps tool definitions data-driven and easy to append.
- Keep README, imports, and filesystem layout synchronized when moving packages so the module stays buildable.
