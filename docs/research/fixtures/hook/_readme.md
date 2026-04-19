# Hook fixtures

Two distinct hook formats are live in production. Both are valid.

## Format 1 - Anthropic JSON + script pair (hookify)

Found in every Anthropic-authored plugin.

**Pairing convention:** `hooks/hooks.json` declares matchers + event routing and points to scripts at `${CLAUDE_PLUGIN_ROOT}/hooks/<event>.<ext>`.

| File | Source | Shape |
|---|---|---|
| `hookify-hooks.json` | [anthropics hookify](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/hookify/hooks/hooks.json) | 4 events routed, each to one Python script, 10s timeout. |
| `hookify-pretooluse.py` | same plugin | Reads JSON from stdin, evaluates local rules, writes JSON decision to stdout. Uses exit 0 for success, emits `{"hookSpecificOutput": {"permissionDecision": "deny"...}}` to block. |
| `hookify-posttooluse.py` | same | Similar shape. |
| `hookify-stop.py` | same | Stop-event handler. |
| `hookify-userpromptsubmit.py` | same | Prompt-submit handler. |

## Format 2 - buildwithclaude Markdown wrapper

Found in buildwithclaude `plugins/hooks-<category>/hooks/<name>.md`.

| File | Source | Shape |
|---|---|---|
| `bwc-file-protection.md` | [buildwithclaude hooks-security](https://github.com/davepoon/buildwithclaude/blob/main/plugins/hooks-security/hooks/file-protection.md) | Markdown wrapper with frontmatter declaring `event`, `matcher`, `language`, `version`, `category`. Script is embedded as a fenced code block inside the body. Their buildsystem extracts it. |

**What Tsuba needs to know:**
- `tsuba new hook <name>` default: Anthropic JSON-pair format (universal, works with `--plugin-dir` out of the box).
- `tsuba new hook --marketplace=buildwithclaude <name>`: emit the markdown-wrapper format with bwc-specific frontmatter.
- Both formats follow the same stdin/stdout/exit-code contract documented in Hanko research Section 4 and the Claude Code hooks reference.

## Exit-code contract (from Hanko phase-1 research + Claude Code hooks ref)

| Exit | Meaning |
|---|---|
| 0 | Success. stdout JSON parsed for `{decision, hookSpecificOutput, continue, systemMessage, ...}`. |
| 2 | Blocking error. stderr is shown; action blocked; stdout JSON ignored. |
| 1, 3+ | Non-blocking error. stderr shown to user/Claude. Action proceeds. |
| any non-zero (WorktreeCreate only) | Failure. |

## Stdin shape

All events receive a JSON body with: `session_id`, `transcript_path`, `cwd`, `permission_mode`, `hook_event_name`, plus event-specific fields (`tool_name`, `tool_input`, `tool_use_id` for tool events).
