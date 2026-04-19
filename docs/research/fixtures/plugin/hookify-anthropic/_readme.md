# hookify (Anthropic official)

**Source:** https://github.com/anthropics/claude-plugins-official/tree/main/plugins/hookify

**Type:** Full plugin with hooks, agents, and skills.

**Shape notes:**
- Manifest lacks `version` (same as example-plugin).
- Has `hooks/hooks.json` listing 4 event types (PreToolUse, PostToolUse, Stop, UserPromptSubmit). Each points to a Python script via `${CLAUDE_PLUGIN_ROOT}/hooks/<event>.py`.
- Shows that the manifest does NOT list `hooks` — the `hooks/hooks.json` is auto-loaded per the v2.1+ convention. Per Hanko research Section 4.5, listing it explicitly in the manifest would cause a duplicate-hooks error.
- Uses `${CLAUDE_PLUGIN_ROOT}` for portable hook paths.
- Agent uses `tools: ["Read", "Grep"]` YAML array, `model: inherit`, `color: yellow`.
- Skill uses `version: 0.1.0` in frontmatter (non-standard — not in official skill frontmatter reference).

**What Tsuba should copy:** the event coverage pattern (one script per event), `${CLAUDE_PLUGIN_ROOT}` everywhere, avoid manifest-level `hooks` reference.
