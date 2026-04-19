# Changelog

All notable changes to tsuba are documented here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow [SemVer](https://semver.org/).

## [Unreleased]

## [0.1.2] - 2026-04-19

### Fixed

- **plugin.json injection (Round 1 blocker):** a user whose name or description contained a JSON-sensitive character (`"`, `\`, newline, tab, Unicode line separator) produced broken plugin.json that failed `hanko check` on the first try. plugin.json is now generated via `encoding/json.MarshalIndent` instead of a `text/template`, eliminating the entire injection class. `TestPluginJSONInjection` pins the round-trip behaviour.
- `--force` now removes any existing target before rewriting, so stale files from an earlier scaffold (including files from an older tsuba version with a different template set) no longer linger.
- `--force` when the target path is a plain file (not a directory) now clears the file cleanly instead of crashing with a cryptic `mkdir: not a directory` error.
- README em-dashes replaced with hyphens per the project style rule.
- `CONTRIBUTING.md` dropped the promise of a non-existent `testdata/expected/<kind>/` fixture directory; reference now points at the actual round-trip test pattern.

### Removed

- `internal/gitctx.OriginURL` was exported but never called. Deleted. If v0.2 needs to read the origin URL for `tsuba publish`, it can come back with credential stripping.

## [0.1.1] - 2026-04-19

### Fixed

- 8 golangci-lint errors from the initial CI run: `errcheck` on `defer f.Close()` (replaced with `fs.ReadFile`), `gosec` G204 annotations on the `exec.Command` call sites in `gitctx` and `hanko` (both inputs are trusted), `gosec` G301 tightened new-directory permissions from 0755 to 0750, `unparam` on unused `stderr` parameters in the doctor and new-command constructors.

## [0.1.0] - 2026-04-19

### Added

- Initial release: scaffold Claude Code skills and plugins with validated frontmatter and marketplace-ready directory layouts.
- `tsuba new plugin <name>` scaffolds a plugin directory with `.claude-plugin/plugin.json`, a sample skill, LICENSE, and README.
- `tsuba new skill <name>` scaffolds a standalone `SKILL.md` with the Anthropic-recommended section structure.
- `tsuba validate` delegates to the [Hanko](https://github.com/RoninForge/hanko) CLI for plugin manifest validation, with a clear error if Hanko is not on PATH.
- `tsuba doctor` reports the status of the local environment (hanko on PATH, git identity configured, current working directory safe to scaffold into).
- `tsuba list marketplaces` prints the 4 supported submission targets and their per-marketplace conventions.
- Author name and email auto-detected from `git config` with CLI-flag overrides.
- Every scaffolded plugin's README footer includes an opt-out "scaffolded with tsuba" attribution.
- Composite GitHub Action wrapper that runs `tsuba validate` on every PR against a plugin repo.

[Unreleased]: https://github.com/RoninForge/tsuba/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/RoninForge/tsuba/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/RoninForge/tsuba/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/RoninForge/tsuba/releases/tag/v0.1.0
