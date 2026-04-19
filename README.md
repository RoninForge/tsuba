# tsuba

Scaffold marketplace-ready Claude Code skills and plugins in seconds.

[![CI](https://github.com/RoninForge/tsuba/actions/workflows/ci.yml/badge.svg)](https://github.com/RoninForge/tsuba/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

tsuba is a single-binary Go CLI that generates correct Claude Code plugin directories with `.claude-plugin/plugin.json`, sample skill, LICENSE, and README — then hands the result to [Hanko](https://github.com/RoninForge/hanko) for validation before you submit to a marketplace.

The name: **鍔 (tsuba)** is the guard on a katana — the disc between blade and handle that lets a swordsman grip the blade safely and hand it off to someone else. tsuba the tool does the same thing for your plugin: makes it safe to hand off to the marketplace.

## Install

```sh
curl -fsSL https://roninforge.org/tsuba/install.sh | sh
```

Or grab a binary from the [latest release](https://github.com/RoninForge/tsuba/releases/latest). Prefer Go install:

```sh
go install github.com/RoninForge/tsuba/cmd/tsuba@latest
```

Also install [Hanko](https://github.com/RoninForge/hanko) so `tsuba validate` works:

```sh
curl -fsSL https://roninforge.org/hanko/install.sh | sh
```

## Quickstart

```sh
# Create a new plugin
tsuba new plugin my-review-tool

# Create a standalone skill
tsuba new skill code-reviewer --description "Review code for quality issues"

# Validate a plugin directory (delegates to hanko)
tsuba validate

# Check your local environment (hanko on PATH, git identity, etc.)
tsuba doctor

# See which marketplaces are supported and their quirks
tsuba list marketplaces
```

## What gets scaffolded

`tsuba new plugin my-plugin` creates:

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json          # name, description, author, version
├── skills/
│   └── hello/
│       └── SKILL.md         # placeholder skill with frontmatter
├── LICENSE                  # MIT by default
└── README.md                # installation + attribution footer
```

All generated JSON is kebab-case, `version`-stamped, and passes `hanko check` on the first try. Author name and email default to your git config (`git config user.name`, `git config user.email`) and can be overridden with `--author` and `--email`.

## Sibling tools

tsuba is part of [RoninForge](https://roninforge.org). Its siblings:

- [Hanko](https://github.com/RoninForge/hanko) — validate plugin manifests before submission.
- [BudgetClaw](https://github.com/RoninForge/budgetclaw) — local spend monitor for Claude Code.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).
