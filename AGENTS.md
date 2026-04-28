# AGENTS.md - Tsuba

> Guidance for AI coding agents working on this repository. Human contributors should read CONTRIBUTING.md first; this file exists to give agents the same context without spelunking.

## What this is

Tsuba (鍔) is a **scaffolder for Claude Code plugins and standalone SKILL.md files**. One command generates a valid plugin directory: `.claude-plugin/plugin.json` (correct schema, author fields auto-populated from `git config`), a sample skill, LICENSE, and README. The output passes Hanko validation on the first run.

Single static Go binary. MIT licensed. Part of the [RoninForge](https://roninforge.org) toolkit.

## Trust pledge (load-bearing - never break this)

1. **Zero network calls in the binary by default.** Embedded `text/template` sources for every generator are vendored. Scaffolding works fully offline. If a feature genuinely needs the network (publishing, fetching the latest schema), it must be behind an explicit subcommand, never on the default path.
2. **No silent overwrites.** Tsuba writes files. If a target path exists, refuse unless the user passed `--force`.

## Validation is delegated to Hanko

`tsuba validate` shells out to [Hanko](https://github.com/RoninForge/hanko) via `os/exec`. We do not duplicate the validation logic.

- Hanko's JSON output schema (`hanko validate --json`) is the contract. If it changes, our parser breaks. Coordinate via the Hanko repo.
- Hanko's exit codes are the contract. 0 = pass, 1 = validation failure, 2 = usage error. We surface them.
- `tsuba doctor` reports whether Hanko is on PATH. Do not bundle Hanko; users install it separately.

## Layout

    cmd/tsuba/             Cobra entrypoint
    internal/scaffold/     Generators for plugin and skill directories
    internal/templates/    Embedded text/template sources (go:embed)
    internal/validate/     Shells out to Hanko
    internal/doctor/       Environment probe (Hanko, git config, PATH)
    action.yml             Composite GitHub Action for CI-time validation
    testdata/              Scaffold output golden files

## Build, test, lint

    make check    # fmt + vet + lint + test (run before any commit)
    make build    # → ./bin/tsuba
    make test     # race detector + coverage
    make snapshot # local goreleaser dry-run, no publish

## What Tsuba scaffolds today

- **Full plugin directory**: `.claude-plugin/plugin.json`, sample SKILL.md, LICENSE (MIT), README.
- **Standalone SKILL.md** under `skills/<name>/`.

Hooks and agents scaffolding lands in v0.2 - if you implement, follow the existing `internal/scaffold/` pattern (one file per generator + golden test fixture).

## Style

- Go 1.22+. `errors.Is` / `errors.As`, not string matching.
- No emoji in code, comments, or generated output. Dev-tool audience.
- No em dashes in user-facing strings (AI-detection giveaway).
- Generated content (`internal/templates/*.tmpl`) is also user-facing. Same rules apply.

## What you should NOT do

- Do not duplicate Hanko validation. Shell out, don't reimplement.
- Do not silently overwrite existing files. Refuse without `--force`.
- Do not introduce a default-path network call. The scaffold flow must work offline.
- Do not assume Hanko is installed. `tsuba scaffold` works without it; `tsuba validate` and `tsuba doctor` report missing Hanko gracefully.

## Releasing

    git tag v0.X.Y
    git push origin v0.X.Y

goreleaser publishes binaries to GitHub Releases and bumps the Homebrew tap (`roninforge/homebrew-tap`).

## More context

- Site: https://roninforge.org/tsuba
- Markdown digest for AI fetchers: https://roninforge.org/tsuba.md
- Tutorial: https://roninforge.org/tutorials/how-to-create-a-claude-skill-for-my-cv (uses Tsuba + Hanko + BudgetClaw end-to-end)
- Sibling tools: [BudgetClaw](https://github.com/RoninForge/budgetclaw), [Hanko](https://github.com/RoninForge/hanko)
