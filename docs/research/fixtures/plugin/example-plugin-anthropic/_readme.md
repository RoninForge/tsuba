# example-plugin (Anthropic official)

**Source:** https://github.com/anthropics/claude-plugins-official/tree/main/plugins/example-plugin

**Type:** Reference plugin, Anthropic-authored. Demonstrates every extension surface (commands, skills, MCP).

**Shape notes:**
- Minimal manifest: only `name`, `description`, `author`. No `version` field, contradicting cc-marketplace's stricter schema.
- Layout: `.claude-plugin/plugin.json`, `commands/example-command.md`, `skills/example-skill/SKILL.md`, `skills/example-command/SKILL.md`, `.mcp.json`.
- Ships BOTH legacy flat `commands/*.md` and new `skills/<name>/SKILL.md` to demonstrate that a single plugin can carry both layouts.
- Uses `[Read, Glob, Grep, Bash]` YAML array form for `allowed-tools` (other plugins use space-separated string).

**What Tsuba should copy:** the minimal manifest pattern (3 fields) as the "RECOMMENDED" default. Tsuba's generated plugins will bias toward `skills/<name>/SKILL.md` (not flat `commands/`).
