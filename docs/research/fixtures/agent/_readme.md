# Agent fixtures

All real `.md` agent files from live plugin repos.

| File | Source | Shape |
|---|---|---|
| `conversation-analyzer.md` | [anthropics hookify](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/hookify/agents/conversation-analyzer.md) | Anthropic-authored. `model: inherit`, `color: yellow`, `tools: ["Read", "Grep"]` YAML array. Long XML-formatted `<example>` blocks in description. Uses no `name` field — derived from filename. |
| `agent-sdk-verifier-py.md` | [anthropics agent-sdk-dev](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/agent-sdk-dev/agents/agent-sdk-verifier-py.md) | Anthropic's reference agent. Demonstrates `name` explicit, long system prompt, detailed instructions. |
| `backend-architect.md` | [buildwithclaude](https://github.com/davepoon/buildwithclaude/blob/main/plugins/agents-development-architecture/agents/backend-architect.md) | buildwithclaude convention: has `category:` field, no `tools` or `model`. |
| `golang-expert.md` | [buildwithclaude](https://github.com/davepoon/buildwithclaude/blob/main/plugins/agents-language-specialists/agents/golang-expert.md) | Same bwc pattern. `category: language-specialists`. No tools restriction. |
| `python-pro.md` | [buildwithclaude](https://github.com/davepoon/buildwithclaude/blob/main/plugins/agents-language-specialists/agents/python-pro.md) | Same bwc pattern. |

**Patterns:**
- Minimum: `description` field. `name` defaults to filename.
- Anthropic uses `tools` as YAML list (`["Read", "Grep"]`); buildwithclaude usually omits.
- `model` is optional: `sonnet`, `opus`, `haiku`, `inherit`, or omit.
- `color` is Anthropic-specific ornamentation.
- `category` is **buildwithclaude-specific** — must match their allowed list if submitting there.
- Body is plain markdown, starts with "You are a..." per bwc convention.
