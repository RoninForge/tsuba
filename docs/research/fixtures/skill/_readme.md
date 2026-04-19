# Skill fixtures

Each skill is a real, live `SKILL.md` pulled from Anthropic repos or the official plugin marketplace.

| File | Source | Shape |
|---|---|---|
| `example-skill.md` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/example-plugin/skills/example-skill/SKILL.md) | Anthropic's reference; demonstrates frontmatter options, `argument-hint`, `allowed-tools` as YAML array. |
| `skill-creator.md` | [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills/skill-creator) | Anthropic's own skill-creator. Minimal frontmatter (name + description only). Body is a long playbook. |
| `mcp-builder.md` | [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills/mcp-builder) | Skill that builds MCP servers. Long body, only 2 frontmatter fields. |
| `webapp-testing.md` | [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills/webapp-testing) | Testing skill. Name + description + `license` field (uncommon). |
| `pdf.md` | [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills/pdf) | Heavy skill (315 lines). Shows progressive-disclosure pattern. |
| `brand-guidelines.md` | [anthropics/skills](https://github.com/anthropics/skills/tree/main/skills/brand-guidelines) | Reference-style content skill (not a task skill). |
| `writing-rules.md` | [anthropics/claude-plugins-official hookify](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/hookify/skills/writing-rules/SKILL.md) | Has `version: 0.1.0` frontmatter field — NOT in official skill reference, but shipped in production anyway. Tsuba should treat `version` as optional/ignored. |

**Patterns across all:**
- Anthropic's own skills almost never fill more than 2 frontmatter fields (`name`, `description`).
- Long body is the norm, not the exception — skill-creator is ~200 lines of prose.
- `allowed-tools`, `disable-model-invocation`, `argument-hint` are all optional and rarely used in Anthropic's own skills.
- The `/template/SKILL.md` file in `anthropics/skills` is literally 5 lines — Tsuba's minimal template should match.
