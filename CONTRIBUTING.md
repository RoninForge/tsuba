# Contributing to tsuba

Thanks for considering a contribution. tsuba has a sharp scope: **scaffold marketplace-ready Claude Code skills, plugins, hooks, and agents**. Validation is delegated to [Hanko](https://github.com/RoninForge/hanko). Contributions that stay in the scaffolding lane are welcome. Anything that grows into a full validator, a Claude Code plugin manager, or a hosted service should be discussed in an issue first.

## Ground rules

- **Honest positioning.** No marketing-speak in code, comments, or docs. No emoji in copy or UI.
- **Stay boring.** Prefer stdlib `text/template` over templating libraries. Prefer plain data structures over abstractions.
- **No network in the default path.** Scaffolding is offline. Only `tsuba publish` makes outbound calls, and only to user-specified marketplace URLs.

## Development setup

Requires Go 1.25 or later.

```sh
git clone https://github.com/RoninForge/tsuba.git
cd tsuba
make build     # compile ./bin/tsuba
make test      # go test ./... with race detector
make lint      # golangci-lint run
make fmt       # gofmt
```

## Running tests

```sh
go test ./...                     # everything
go test ./internal/scaffold       # single package
go test -race -cover ./...        # what CI runs
```

Every new template should ship with:

1. An end-to-end test in `scaffold_test.go` that generates the template and asserts the output shape.
2. If the template involves a structured format (JSON, YAML, TOML), a round-trip test that re-parses the output to guarantee the format stays valid when user-supplied fields contain tricky characters (quotes, backslashes, newlines). The `TestPluginJSONInjection` table is the reference pattern.

## Commit style

Conventional Commits, low ceremony:

```
feat(scaffold): add agent template
fix(cli): reject non-kebab-case plugin names
docs: clarify skill vs standalone-skill layout
```

## Pull requests

1. Open an issue first for anything larger than a typo.
2. Write tests.
3. Run `make check` locally and make sure CI is green.
4. Describe the behavior change in the PR body.

## Code layout

```
cmd/tsuba/         thin main() entrypoint
internal/cli/      cobra command tree
internal/version/  build-time version metadata
internal/scaffold/ template rendering + directory writing (+ round-trip tests)
internal/templates/ //go:embed'd text/template sources
internal/hanko/    shell-out adapter to the hanko binary
internal/gitctx/   git config read helpers (user.name, user.email)
action/            composite GitHub Action wrapper
scripts/           install script and helpers
docs/research/     phase-1 research artefacts (not shipped in the binary)
```

Scaffold round-trip fixtures live inline in `internal/scaffold/scaffold_test.go`
(see `TestPluginJSONInjection`). There is no separate `testdata/` directory.

## Reporting security issues

See [SECURITY.md](SECURITY.md). Do not file public issues for security bugs.
