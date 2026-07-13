# Agent Notes

This file gives contributor agents enough project context to make focused changes without rediscovering the repo shape.

## Project Shape

- This is a Go HTTP service that proxies upstream iCalendar feeds, filters events, and serves filtered calendar feeds.
- YAML-backed config types use a `Config` suffix where there is a separate runtime representation.
- Runtime config is built in two phases:
  - `LoadConfig(file)` reads YAML, resolves secret files, and validates calendar-level settings.
  - `Config.Compile()` builds `RuntimeConfig`, compiles match rules, and rejects invalid regexes before serving.
- Keep the config/runtime split clear:
  - `CalendarConfig -> Calendar`
  - `FilterConfig -> Filter`
  - `EventMatchRulesConfig -> EventMatchRules`
  - `StringMatchRuleConfig -> StringMatchRule`
- Regex match rules should be compiled during config compilation, not during per-event processing.
- Transforms are not currently compiled; `StringTransformRule` is used directly at runtime.

## Common Commands

Run the full test suite:

```sh
go test ./...
```

Run the configured linter:

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --config .golangci.yml ./...
```

Run Dockerfile linting when touching container files:

```sh
hadolint Dockerfile
```

Run Nix checks when touching Nix files:

```sh
nix flake check
```

## Contribution Preferences

- Keep PRs focused. Prefer small, reviewable changes over broad refactors.
- Follow existing naming and structure before introducing new abstractions.
- Add or update tests for behavior changes, especially config validation, HTTP handling, filtering, metrics, and security-sensitive code.
- Prefer returning errors from lower-level code and logging at the application boundary.
- Preserve security behavior around token handling, URL redaction, normalized metrics labels, and management endpoint separation.
- If adding new match condition types, add them through the config-to-runtime compilation path rather than doing repeated parsing during event processing.

## CI And Packaging Notes

- GitHub Actions runs Go tests, linting, Docker checks, Nix checks, and workflow linting.
- Docker builds should run broadly, but image pushes are limited to trusted main/tag workflows.
- GoReleaser is used for release notes and release publishing.
- Nix builds use the flake source revision for build metadata; tagged Nix runs still build the tagged source.
