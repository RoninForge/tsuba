# Changelog

All notable changes to tsuba are documented here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), versions follow [SemVer](https://semver.org/).

## [Unreleased]

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
