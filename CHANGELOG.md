# Changelog

All notable changes to tsuba are documented here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow [SemVer](https://semver.org/).

## [Unreleased]

## [0.1.5] - 2026-04-19

### Fixed

- `LICENSE` and `README.md` templates no longer emit a trailing space after the year when `Author.Name` is empty. Uses `{{with .Author.Name}} {{.}}{{end}}` so the name is only appended when present (Round 4, T3).
- `tsuba doctor` stale copy: told users `author fields will be placeholders` when an unset git identity actually omits the whole author object from `plugin.json` post-v0.1.4. Now says so (Round 4, T3).
- `.goreleaser.yaml` migrated from the deprecated `format: zip` override to the current `formats: ["zip"]` array form. goreleaser v2 still runs the deprecated path but warns on every release (Round 4, T2).

## [0.1.4] - 2026-04-19

### Fixed

- **Partial-author still broke hanko on first try (Round 3, T1):** the v0.1.3 guard used OR (`Name != "" || Email != ""`), which still emitted `{"name":"","email":"X"}` for an email-only author. Hanko rejected that with a `HANKO-SCHEMA` minLength:1 error on `name`. The original test `TestPluginJSONEmailOnlyKeepsAuthor` even pinned the buggy behaviour as intended. Fix: scaffold now emits the author object only when `Name != ""`. `Author.Email` grew an `omitempty` json tag so a name-only author serializes as `{"name":"X"}` cleanly. Email remains optional per the schema; name remains required.
- **Regression guard for the broken-symlink Lstat fix:** added `TestEnsureTargetWithBrokenSymlink` so a future refactor that flips `Lstat` back to `Stat` fails CI. Non-Windows build tag (os.Symlink on Windows needs admin).

### Changed

- `yamlQuoteString` fallback branch changed from silent `""` return to an explicit panic. `json.Marshal` of a plain Go string cannot fail per stdlib contract; the panic documents the impossibility rather than papering over a broken build.

### Added

- `SampleSkill` field in `scaffold.Options` is now kebab-case validated (`ErrSampleSkillInvalid`). Defensive: SampleSkill is not a CLI flag today, but if it ever becomes one, non-kebab input would flow into YAML `name:` frontmatter in `sample-skill.md.tmpl` and reopen the T2-2 injection class. Guarded pre-emptively.

## [0.1.3] - 2026-04-19

### Fixed

- **Empty author broke hanko validation on fresh machines (Round 2, T2-1):** A user with no `git config user.name` and no `--author` produced a `plugin.json` with `"author": {"name":"","email":""}` that hanko rejected with a `HANKO-SCHEMA` error on `minLength: 1`. `pluginManifest.Author` is now `*Author` with `omitempty`; the object is omitted entirely when both fields are empty, which downgrades the hanko signal to the `HANKO003` warning (non-blocking). The "passes hanko on first try" promise now holds for fresh-machine users.
- **YAML injection in standalone SKILL.md frontmatter (Round 2, T2-2):** `description: {{.Description}}` dropped raw user strings into YAML. A `--description "# comment"` was parsed as `null`; newlines broke the document; `[a, b]` was parsed as a flow sequence. Fix: added a `yamlQuoteString` helper that routes through `json.Marshal` (YAML is a superset of JSON for scalars) and changed the template to emit `description: "..."` with proper escaping.
- **Broken symlinks confused `ensureTarget` (Round 2, T2-4):** `os.Stat` follows symlinks, so a dangling link at the target path returned `fs.ErrNotExist` and the code took the "nothing there" branch. Switched to `os.Lstat` so broken links are seen as "exists" and `--force` cleans them via `RemoveAll`.
- `filepath.IsLocal` replaces a hand-rolled path-escape check that would have false-matched legitimate names like `..hidden.md` (Round 2, T3-5).
- `CONTRIBUTING.md` stopped referencing a non-existent `testdata/` layout (Round 2, T2-3).

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

[Unreleased]: https://github.com/RoninForge/tsuba/compare/v0.1.5...HEAD
[0.1.5]: https://github.com/RoninForge/tsuba/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/RoninForge/tsuba/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/RoninForge/tsuba/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/RoninForge/tsuba/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/RoninForge/tsuba/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/RoninForge/tsuba/releases/tag/v0.1.0
