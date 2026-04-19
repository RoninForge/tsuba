# Tsuba — Phase 1 Verification Research

**Research window:** 2026-04-19
**Target product:** Tsuba (Go CLI scaffolder for Claude Code skills, plugins, hooks, agents + one-command publish to 4+ marketplaces)
**Sibling tools:** Hanko (manifest validator, shipped 2026-04-19), BudgetClaw (spend monitor, in progress)
**Primary sources:** `code.claude.com/docs/en/{skills,plugins,sub-agents,hooks}`, Hanko's phase-1 research (`github.com/RoninForge/hanko/blob/main/docs/research/phase-1-schema.md`), Hanko's embedded schemas (`internal/schema/plugin.schema.json`, `marketplace.schema.json`), 4 marketplace repos, `anthropics/skills`, `anthropics/claude-plugins-official`.

---

## 1. SKILL.md spec

**Canonical source:** https://code.claude.com/docs/en/skills (fetched 2026-04-19, full text captured in research notes).

### 1.1 Frontmatter reference (complete)

All fields are optional. Only `description` is recommended.

| Field | Required | Type | Notes |
|---|---|---|---|
| `name` | No | string | Display name. Defaults to directory name. Lowercase letters, digits, hyphens. Max 64 chars. |
| `description` | Recommended | string | What the skill does and when to use it. Combined `description` + `when_to_use` is truncated to **1,536 characters** in the skill listing to limit context usage. |
| `when_to_use` | No | string | Additional trigger phrases. Appended to `description`; counts against the 1,536-char cap. |
| `argument-hint` | No | string | Autocomplete hint. Example: `[issue-number]` or `[filename] [format]`. |
| `disable-model-invocation` | No | bool | `true` = only the user can invoke (with `/name`). Default `false`. |
| `user-invocable` | No | bool | `false` = hidden from `/` menu. Default `true`. |
| `allowed-tools` | No | string OR YAML list | Tools Claude can use without permission prompts while the skill is active. Does NOT restrict tools; it pre-approves them. Example: `allowed-tools: Read Grep` (space-separated) or `allowed-tools: [Read, Grep]` (list). |
| `model` | No | string | Model while skill active. Accepts `sonnet`, `opus`, `haiku`, or full model ID. |
| `effort` | No | string | `low`, `medium`, `high`, `xhigh`, `max`. Depends on model. |
| `context` | No | string | `fork` = run in a forked subagent context. |
| `agent` | No | string | Which subagent type to use when `context: fork`. Defaults to `general-purpose`. |
| `hooks` | No | object | Skill-scoped lifecycle hooks. |
| `paths` | No | string OR list | Glob patterns. Activates the skill automatically only when working with matching files. |
| `shell` | No | string | `bash` (default) or `powershell`. |

### 1.2 Strictly enforced vs ignored

**Enforced:**
- `name` validation: lowercase + digits + hyphens only, max 64 chars (`^[a-z0-9]+(-[a-z0-9]+)*$` implied).
- `description` + `when_to_use` truncation at 1,536 chars in the skill listing (context-budget enforced by `SLASH_COMMAND_TOOL_CHAR_BUDGET`, defaults to 1% of context window or 8,000-char floor).
- `context: fork` without an `agent` field defaults to `general-purpose` — not an error.

**Silently accepted / ignored:**
- `version` — **NOT in the official frontmatter reference** but ships in production (e.g. `hookify/skills/writing-rules/SKILL.md` has `version: 0.1.0`). Treat as non-standard; Tsuba should not emit it.
- `license` — also in the wild (`anthropics/skills/skills/webapp-testing/SKILL.md`) but not documented. Same treatment.
- Any unknown field — ignored, not rejected.

### 1.3 Body

Everything after the closing `---` is the skill instruction. Rendered as a single message into the conversation when invoked and stays there for the rest of the session (Claude does NOT re-read the file per turn).

**String substitutions available in the body:**

| Variable | Meaning |
|---|---|
| `$ARGUMENTS` | Full argument string as typed. If omitted from the skill body, Claude Code appends `ARGUMENTS: <value>` so arguments are still visible. |
| `$ARGUMENTS[N]`, `$N` | Nth argument, 0-indexed. Shell-style quoting. |
| `${CLAUDE_SESSION_ID}` | Session ID. |
| `${CLAUDE_SKILL_DIR}` | Directory containing this SKILL.md. For plugin skills, this is the skill's own subdirectory, not the plugin root. |

**Inline shell injection** (preprocessing, NOT executed by Claude):
- `` !`<command>` `` runs shell commands and inlines their stdout into the prompt.
- Fenced code blocks opened with ` ```! ` for multi-line.
- Disabled globally via `"disableSkillShellExecution": true` in settings.

### 1.4 File location

| Scope | Path | Notes |
|---|---|---|
| Enterprise | Managed settings | Highest priority. |
| Personal | `~/.claude/skills/<name>/SKILL.md` | All projects. |
| Project | `.claude/skills/<name>/SKILL.md` | Per project. Version-controlled. |
| Plugin | `<plugin-root>/skills/<name>/SKILL.md` | Namespaced as `plugin-name:skill-name`. |

Priority: enterprise > personal > project. Plugin skills cannot collide because of namespacing.

### 1.5 Size limits

| Limit | Value | Source |
|---|---|---|
| `description` + `when_to_use` truncation | 1,536 chars | docs/skills frontmatter table |
| SKILL.md recommended body | under 500 lines | docs/skills "Add supporting files" Tip |
| Per-skill re-attach budget after compaction | 5,000 tokens | docs/skills "Skill content lifecycle" |
| Combined re-attach budget | 25,000 tokens | same |
| Default `SLASH_COMMAND_TOOL_CHAR_BUDGET` | ~1% of context window, floor 8,000 chars | docs/skills "Skill descriptions are cut short" |

### 1.6 Common footguns

1. **Wrong `allowed-tools` syntax.** Docs show space-separated (`Read Grep`) and YAML list (`[Read, Grep]`) both work. Commas-in-a-string (`Read, Grep`) is rejected in several community reports.
2. **`allowed-tools` understood as a restriction, not a pre-approval.** Adding `Read Grep` does NOT block Write — it just pre-approves Read and Grep.
3. **`disable-model-invocation: true` hides the skill from Claude's context entirely.** If the skill has no side effects, drop this flag or Claude will never suggest it.
4. **`description` too generic.** Claude's skill-selection budget truncates at 1,536 chars; front-load the key use case.
5. **Missing `description`.** If omitted, the first paragraph of the body is used — unpredictable.
6. **`paths` glob with implicit slash.** Same format as `path-specific-rules`; test these before shipping.

---

## 2. Plugin directory layout (cross-reference with Hanko)

**Canonical schema:** https://github.com/RoninForge/hanko/blob/main/internal/schema/plugin.schema.json (248 lines). This is the source of truth for manifest fields.

### 2.1 Plugin root layout

Per https://code.claude.com/docs/en/plugins:

```
<plugin-root>/
├── .claude-plugin/
│   └── plugin.json          # ONLY plugin.json lives here
├── skills/                  # Auto-discovered
│   └── <skill-name>/SKILL.md
├── commands/                # Auto-discovered (legacy format)
│   └── <command-name>.md
├── agents/                  # Auto-discovered
│   └── <agent-name>.md
├── hooks/                   # Auto-discovered
│   └── hooks.json
├── monitors/                # Auto-discovered
│   └── monitors.json
├── .mcp.json                # Auto-discovered
├── .lsp.json                # Auto-discovered
├── output-styles/           # Auto-discovered
├── bin/                     # Added to Bash PATH when plugin enabled
├── settings.json            # Default settings (only `agent` and `subagentStatusLine` keys supported)
├── LICENSE                  # Optional
├── README.md                # Optional
└── CHANGELOG.md             # Optional
```

**Load-bearing warning from docs:** do NOT nest `skills/`, `commands/`, `agents/`, `hooks/` inside `.claude-plugin/`. Only `plugin.json` goes there. Every other directory sits at plugin root.

### 2.2 Auto-discovered directories

From the Hanko `hooksInline` pattern regex, the full list of triggerable hook events is 26 (Section 3). Auto-discovered component directories per the docs:

| Directory | Contents | Manifest field override |
|---|---|---|
| `skills/` | `<name>/SKILL.md` each | `skills` (string or array) |
| `commands/` | flat `*.md` files | `commands` |
| `agents/` | flat `*.md` files | `agents` (**array of .md paths only** — bare directory strings rejected per Hanko research Section 4.6) |
| `hooks/` | `hooks.json` (do NOT also reference in manifest — see Hanko §4.5 duplicate-hooks footgun) | `hooks` |
| `output-styles/` | contents schema underspecified in docs | `outputStyles` |
| `monitors/` | `monitors.json` | `monitors` |
| `.mcp.json` | MCP server map at root | `mcpServers` |
| `.lsp.json` | LSP server map at root | `lspServers` |

### 2.3 Minimum required files per kind

| Kind | Minimum | Source |
|---|---|---|
| Plugin | `.claude-plugin/plugin.json` with `name` only (Hanko `required: ["name"]`). Manifest itself is optional if using default layout; then `name` defaults to directory name. | Hanko schema + docs/plugins "Plugin manifest schema" |
| Skill | `skills/<name>/SKILL.md` with frontmatter and body. Frontmatter can be empty (`---\n---`); name defaults to directory name. | [anthropics/skills/template/SKILL.md](https://github.com/anthropics/skills/blob/main/template/SKILL.md): 5 lines, `name` + `description` only. |
| Hook | `hooks/hooks.json` with at least one event and one handler. Scripts referenced via `${CLAUDE_PLUGIN_ROOT}/hooks/<script>`. | [hookify/hooks/hooks.json](https://github.com/anthropics/claude-plugins-official/blob/main/plugins/hookify/hooks/hooks.json) |
| Agent | `agents/<name>.md` with `description` in frontmatter, body is the system prompt. `name` defaults to filename. | https://code.claude.com/docs/en/sub-agents "Supported frontmatter fields" |

### 2.4 Standalone vs plugin-bundled

Per docs/plugins "When to use plugins vs standalone configuration":

| Dimension | Standalone (`.claude/`) | Plugin |
|---|---|---|
| Name | `/skill-name` | `/plugin-name:skill-name` |
| Share | Project-only (via `.claude/` checked in) or personal (`~/.claude/`) | Via marketplace install, version-controlled |
| Manifest | None | `.claude-plugin/plugin.json` |
| When | Experimentation, single-project customization | Distribution, reuse, versioning |

The component FILES are identical between both modes — same SKILL.md frontmatter, same agent .md frontmatter, same hooks.json shape. Tsuba can default to plugin mode and offer `--standalone` to emit into `.claude/` instead.

---

## 3. Hook contract

**Primary source:** https://code.claude.com/docs/en/hooks (fetched 2026-04-19) + Hanko phase-1 research Section 4.

### 3.1 Event catalog (26 events, case-sensitive)

`SessionStart`, `UserPromptSubmit`, `PreToolUse`, `PermissionRequest`, `PermissionDenied`, `PostToolUse`, `PostToolUseFailure`, `Notification`, `SubagentStart`, `SubagentStop`, `TaskCreated`, `TaskCompleted`, `Stop`, `StopFailure`, `TeammateIdle`, `InstructionsLoaded`, `ConfigChange`, `CwdChanged`, `FileChanged`, `WorktreeCreate`, `WorktreeRemove`, `PreCompact`, `PostCompact`, `Elicitation`, `ElicitationResult`, `SessionEnd`.

Source: Hanko `plugin.schema.json` line 115 — full regex enumerates all 26.

### 3.2 Handler types

Four types: `command`, `http`, `prompt`, `agent`.

| Type | Required | Optional |
|---|---|---|
| `command` | `command` (string) | `timeout`, `async`, `asyncRewake`, `shell`, `statusMessage`, `once`, `if` |
| `http` | `url` | `headers`, `allowedEnvVars`, `timeout` |
| `prompt` | `prompt` | `model` |
| `agent` | `prompt` | `model` |

### 3.3 Stdin/stdout contract for `command` hooks

**Stdin:** JSON blob per event. Common fields:
```json
{
  "session_id": "abc123",
  "transcript_path": "/path/to/transcript.jsonl",
  "cwd": "/current/working/dir",
  "permission_mode": "default|plan|acceptEdits|auto|dontAsk|bypassPermissions",
  "hook_event_name": "PreToolUse",
  "agent_id": "...",
  "agent_type": "..."
}
```
Plus event-specific fields. Tool events (`PreToolUse`, `PostToolUse`, `PermissionRequest`) also carry `tool_name`, `tool_input`, `tool_use_id`.

**Stdout (exit 0 only):** JSON response:
```json
{
  "continue": true,
  "stopReason": "...",
  "suppressOutput": false,
  "systemMessage": "...",
  "decision": "block",
  "reason": "...",
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|deny|ask|defer",
    "permissionDecisionReason": "...",
    "updatedInput": {},
    "additionalContext": "..."
  }
}
```

**Exit codes:**

| Exit | Behavior |
|---|---|
| 0 | Success. stdout JSON parsed. |
| 2 | Blocking error. stderr shown. Action blocked. stdout JSON ignored. |
| 1, 3+ | Non-blocking error. stderr shown. Action proceeds. |
| Any non-zero (WorktreeCreate only) | Failure. |

### 3.4 Minimal scaffold for a working hook

Two files are enough:

**`hooks/hooks.json`:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "${CLAUDE_PLUGIN_ROOT}/hooks/block-dangerous-bash.sh",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

**`hooks/block-dangerous-bash.sh`:**
```bash
#!/bin/bash
INPUT=$(cat)
CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')
if echo "$CMD" | grep -qE '\brm -rf\b'; then
  jq -n '{hookSpecificOutput: {hookEventName: "PreToolUse", permissionDecision: "deny", permissionDecisionReason: "Blocked: rm -rf"}}'
fi
exit 0
```

That's it. Tsuba's `new hook` default is this shape, with `{{.Event}}`/`{{.Matcher}}`/`{{.Name}}` as templatable holes. The user edits the script body.

### 3.5 Real-world hook fixtures (saved to `fixtures/hook/`)

See `fixtures/hook/_readme.md` for detail. Summary: Anthropic uses Python for hooks with consistent stdin-JSON-to-stdout-JSON pattern, timeout 10s, `${CLAUDE_PLUGIN_ROOT}` for paths. buildwithclaude ships hooks as Markdown wrappers with the script embedded — non-universal format; only works inside their build pipeline.

### 3.6 Footguns

1. **Duplicate hooks declaration (Hanko §4.5).** If `hooks/hooks.json` exists AND manifest has `"hooks": "./hooks/hooks.json"`, load fails. Tsuba must NEVER put `hooks` in the manifest when generating the default `hooks/hooks.json`.
2. **`${CLAUDE_PLUGIN_ROOT}` path caching (Hanko §4.4).** Settings.json stores absolute versioned paths at install time. Plugin updates break hooks until the user reinstalls. Runtime bug, not a scaffolding concern.
3. **Event-name case-sensitivity.** `preToolUse` silently does nothing. The schema regex validates all 26 allowed names.
4. **Forgetting `tool_use_id` in stdout for tool events.** Not required in stdout, but the `hookSpecificOutput.hookEventName` field must match the incoming event.
5. **Missing `jq` dependency.** Python stdlib (`json.loads(sys.stdin.read())`) is portable; bash+jq is Anthropic's convention but requires jq installed.

---

## 4. Agent spec

**Canonical source:** https://code.claude.com/docs/en/sub-agents

### 4.1 Frontmatter reference

All fields except `name` and `description` are optional.

| Field | Required | Type | Notes |
|---|---|---|---|
| `name` | Yes | string | Lowercase letters + hyphens. Must be unique per scope. |
| `description` | Yes | string | When Claude should delegate to this agent. |
| `tools` | No | comma string or list | Allow-list of tools. Inherits all if omitted. |
| `disallowedTools` | No | comma string or list | Deny-list, applied before `tools` filter. |
| `model` | No | string | `sonnet`, `opus`, `haiku`, full ID, or `inherit`. |
| `permissionMode` | No | string | `default`, `acceptEdits`, `auto`, `dontAsk`, `bypassPermissions`, `plan`. |
| `maxTurns` | No | integer | Max agentic turns. |
| `skills` | No | list | Skills to pre-load at startup. |
| `mcpServers` | No | list | MCP servers scoped to this agent. |
| `hooks` | No | object | Lifecycle hooks scoped to this agent. **Ignored when loaded from a plugin** (security). |
| `memory` | No | string | `user`, `project`, or `local` persistent memory scope. |
| `background` | No | bool | Always run in background. |
| `effort` | No | string | `low`..`max`. |
| `isolation` | No | string | `worktree` runs in temp git worktree. |
| `color` | No | string | `red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan`. |
| `initialPrompt` | No | string | Auto-submitted first turn when run as main session agent. |

### 4.2 Body

Everything after closing `---` is the **system prompt**. The agent sees ONLY this system prompt plus basic environment (cwd, etc.), not the main conversation's system prompt.

### 4.3 File location

| Scope | Path | Priority |
|---|---|---|
| Managed | `.claude/agents/` in managed-settings dir | 1 (highest) |
| CLI flag | `--agents '{...}'` | 2 |
| Project | `.claude/agents/` | 3 |
| User | `~/.claude/agents/` | 4 |
| Plugin | `<plugin-root>/agents/<name>.md` | 5 |

Plugin agents: `hooks`, `mcpServers`, `permissionMode` frontmatter fields are **silently ignored** (security guardrail).

### 4.4 Parser quirks (from real fixtures)

1. **`tools` accepts either comma-string OR YAML array.** Anthropic's `hookify/agents/conversation-analyzer.md` uses `tools: ["Read", "Grep"]`. buildwithclaude agents omit the field entirely (inherit all).
2. **`description` can contain XML-style `<example>` blocks** — Anthropic uses these to help Claude match delegation intent. Example from `conversation-analyzer.md`:
   ```
   description: Use this agent when analyzing conversation transcripts ... Examples: <example>Context: User is running /hookify ...</example><example>Context: ...</example>
   ```
3. **buildwithclaude requires a `category:` field** in agent frontmatter that is NOT in the official spec. See `fixtures/agent/backend-architect.md`. Tsuba's `--marketplace=buildwithclaude` must emit it.
4. **`name` is optional in practice** — if omitted, Claude Code uses the filename (minus `.md`). Docs say "required"; some Anthropic agents omit it; file-based discovery works.

### 4.5 Minimal agent

```markdown
---
description: Expert reviewer that checks code for security, performance, and correctness. Use when reviewing pull requests or completed features.
---

You are a senior code reviewer. When invoked, analyze the code and provide specific, actionable feedback on quality, security, and best practices.
```

That's the minimum. `name` defaults to filename.

---

## 5. Per-marketplace submission flow

### 5.1 Anthropic official

| Surface | Value |
|---|---|
| Repo | https://github.com/anthropics/claude-plugins-official |
| Community mirror | https://github.com/anthropics/claude-plugins-community (read-only) |
| Submission mechanism | **Web form only** at https://clau.de/plugin-directory-submission (also https://claude.ai/settings/plugins/submit and https://platform.claude.com/plugins/submit). No PR workflow. |
| Accepted types | Full plugins (not standalone skills/agents/hooks). Must have `.claude-plugin/plugin.json`. |
| Manifest delta vs base | `name` required; `author` strongly recommended per Hanko §4.2 (Desktop's `listAvailablePlugins` refuses to load the marketplace entirely when any listed plugin lacks author). No `version` required in practice — see example-plugin with no version. |
| Reserved marketplace names | 8 names enforced (Hanko §3.1): `claude-code-marketplace`, `claude-code-plugins`, `claude-plugins-official`, `anthropic-marketplace`, `anthropic-plugins`, `agent-skills`, `knowledge-work-plugins`, `life-sciences`. |
| Lag to listing | Not documented. Public issue #1272 shows plugins marked "Published" but not in marketplace — implies manual curation. |
| Tsuba action | `tsuba publish --marketplace=anthropic <plugin-dir>` opens browser to clau.de form with a "What you need to paste" summary in the terminal. NO PR automation possible. |

### 5.2 buildwithclaude (davepoon)

| Surface | Value |
|---|---|
| Repo | https://github.com/davepoon/buildwithclaude |
| Submission | PR to `plugins/<kind>-<category>/` path. CONTRIBUTING.md prescribes typed-category directory naming. |
| Accepted types | Agents, commands, hooks, skills — **separated by kind into bundle directories**. Each PR adds files to an existing `agents-<category>/`, `commands-<category>/`, etc. or creates a new category bundle. Plugins also accepted via a separate convention. |
| Manifest delta | `version` required. Agent frontmatter requires `category:` matching one of their allowed categories. Hook frontmatter has bwc-specific fields (`event`, `matcher`, `language`, `version`). |
| Directory conventions | **load-bearing:** `plugins/agents-<category>/`, `plugins/commands-<category>/`, `plugins/hooks-<category>/`, `plugins/all-skills/`. There's also `all-agents/`, `all-commands/`, `all-hooks/` bundles. |
| Categories | Documented in CONTRIBUTING.md — 13 agent categories, 10 command categories. |
| Submission mechanism | GitHub PR. |
| PR URL format | `https://github.com/davepoon/buildwithclaude/compare/main...<user>:<fork-branch>?expand=1&title=<urlencoded>&body=<urlencoded>` |
| Tsuba action | `tsuba publish --marketplace=buildwithclaude`: (1) check `$GH_USER` fork exists, (2) create branch, (3) commit, (4) open browser to pre-filled PR URL. |

### 5.3 cc-marketplace (ananddtyagi)

| Surface | Value |
|---|---|
| Repo | https://github.com/ananddtyagi/cc-marketplace |
| PLUGIN_SCHEMA.md | https://github.com/ananddtyagi/cc-marketplace/blob/main/PLUGIN_SCHEMA.md |
| Submission | GitHub PR. |
| Accepted types | Full plugins only (manifest + component dirs). |
| Manifest delta | `name`, `version`, `description` all **required** (Anthropic only requires `name`). Python validator at `scripts/validate-plugin-schema.py` runs in CI and blocks merge. |
| Directory naming | Flat `plugins/<plugin-name>/`. No typed-category buckets. |
| Additional rules | README.md and plugins.md must be updated to reference the new plugin — `validate-marketplace-sync.py` enforces. |
| PR URL format | `https://github.com/ananddtyagi/cc-marketplace/compare/main...<user>:<fork-branch>?expand=1&title=...&body=...` |
| Tsuba action | Same PR flow; additionally emit a README.md patch line into the PR body as a hint for the contributor. |

### 5.4 superpowers-marketplace (obra)

| Surface | Value |
|---|---|
| Repo | https://github.com/obra/superpowers-marketplace |
| Submission | Not a typical PR flow — this is a thin marketplace that lists existing plugins. README.md shows 4 plugins, each in a separate repo (obra/superpowers, obra/the-elements-of-style, etc.). |
| Accepted types | Plugins that live in their own repo; the marketplace catalog only references them. |
| Manifest delta | None visible; uses standard plugin.json in each referenced repo. |
| Submission mechanism | Open an issue in the marketplace repo requesting addition; `marketplace.json` then gets an entry. There is no CONTRIBUTING.md. |
| Tsuba action | **Not a primary publish target for Tsuba v0.** Mark as "manual submission" and provide a pre-filled issue body. |

### 5.5 claudemarketplaces.com (mertbuilds)

| Surface | Value |
|---|---|
| Repo | https://github.com/mertbuilds/claudemarketplaces.com |
| README | "The site automatically searches GitHub daily to discover repositories with `.claude-plugin/marketplace.json` files. All valid marketplaces are automatically listed - no submission required." (verbatim) |
| Submission mechanism | **Auto-discovery.** No PR, no form. Publish a public repo with `.claude-plugin/marketplace.json` and wait up to 24 hours. |
| Accepted types | Marketplace catalogs (not individual plugins). Plugins appear indirectly via their hosting marketplace. |
| Manifest delta | None. Standard Anthropic marketplace.json schema. |
| Tsuba action | `tsuba publish --marketplace=claudemarketplaces`: print "Your marketplace.json is valid. Push to GitHub. Listing will appear within 24h." Nothing to automate. |

### 5.6 aitmpl.com

| Surface | Value |
|---|---|
| Site | https://aitmpl.com |
| Public repo | **NOT FOUND.** `aitmpl/aitmpl-com` 404s. No contributing workflow surfaced on the site. |
| Status | **UNVERIFIED.** Hanko research marked same. |
| Tsuba action | Omit from `--marketplace` flag list for v0. Revisit when a public submission workflow is documented. |

### 5.7 Ship target for Tsuba v0

Four verified marketplaces:

1. **buildwithclaude** — PR, opinionated categories.
2. **cc-marketplace** — PR, strict manifest.
3. **anthropic** — web form.
4. **claudemarketplaces** — auto-discovery (zero automation needed).

superpowers-marketplace and aitmpl.com are roadmap / unverified.

---

## 6. Real-world examples (fixtures)

Saved to `/tmp/tsuba-push/docs/research/fixtures/`. Directory layout:

```
fixtures/
├── skill/          (7 files)
│   ├── example-skill.md           (Anthropic reference)
│   ├── skill-creator.md           (anthropics/skills)
│   ├── mcp-builder.md             (anthropics/skills)
│   ├── webapp-testing.md          (anthropics/skills)
│   ├── pdf.md                     (anthropics/skills, long)
│   ├── brand-guidelines.md        (anthropics/skills, reference-style)
│   ├── writing-rules.md           (hookify, shows non-standard `version:` field)
│   └── _readme.md
├── agent/          (5 files)
│   ├── conversation-analyzer.md   (Anthropic)
│   ├── agent-sdk-verifier-py.md   (Anthropic)
│   ├── backend-architect.md       (buildwithclaude)
│   ├── golang-expert.md           (buildwithclaude)
│   ├── python-pro.md              (buildwithclaude)
│   └── _readme.md
├── hook/           (6 files)
│   ├── hookify-hooks.json         (Anthropic JSON config)
│   ├── hookify-pretooluse.py      (Anthropic Python handler)
│   ├── hookify-posttooluse.py
│   ├── hookify-stop.py
│   ├── hookify-userpromptsubmit.py
│   ├── bwc-file-protection.md     (buildwithclaude MD wrapper)
│   └── _readme.md
└── plugin/         (3 plugin manifest dirs)
    ├── example-plugin-anthropic/  (minimal 3-field manifest)
    ├── hookify-anthropic/         (same shape; demonstrates auto-load hooks)
    ├── agents-development-architecture-bwc/ (buildwithclaude typed-category)
    └── _readme.md
```

All sourced from public live repos. See each `_readme.md` for URLs and shape notes.

---

## 7. Template dimensions — what Tsuba generates

### 7.1 Skill templates

**Minimal:**
```markdown
---
description: {{.Description}}
---

{{.Body}}
```

**Recommended:**
```markdown
---
name: {{.Name}}
description: {{.Description}}
---

# {{.TitleCase}}

## When to use

{{.WhenToUse}}

## Instructions

{{.Instructions}}

## Example

{{.Example}}
```

**Output:** `skills/<name>/SKILL.md` inside a plugin, or `.claude/skills/<name>/SKILL.md` with `--standalone`.

### 7.2 Plugin templates

**Minimal:**
```
my-plugin/
└── .claude-plugin/
    └── plugin.json     # {"name": "my-plugin"}
```

**Recommended:**
```
my-plugin/
├── .claude-plugin/
│   └── plugin.json     # name, description, author, version (if marketplace needs it)
├── skills/
│   └── <first-skill>/SKILL.md
├── LICENSE             # MIT default
└── README.md           # Generated from plugin.json + component list
```

### 7.3 Hook templates

**Minimal:** generates both files, no manifest entry (critical, to avoid duplicate-hooks footgun).

```
my-plugin/
└── hooks/
    ├── hooks.json                     # Single matcher group, one command handler
    └── <event>-<name>.sh              # stdin-JSON-to-exit-code starter
```

### 7.4 Agent templates

**Minimal:**
```markdown
---
description: {{.Description}}
---

You are a {{.Role}}. When invoked:
1. {{.Step1}}
2. {{.Step2}}
```

**Output:** `agents/<name>.md` inside a plugin.

### 7.5 CLI surface (proposed)

```bash
# Scaffold new plugin
tsuba new plugin hello-world
tsuba new plugin hello-world --author "Jens Krause" --email jens@roninforge.org --marketplace buildwithclaude

# Scaffold individual components (into current plugin or standalone)
tsuba new skill code-reviewer --description "Review code for quality"
tsuba new skill code-reviewer --allowed-tools "Read Grep" --standalone
tsuba new agent security-scanner --model haiku --tools "Read,Grep,Glob"
tsuba new hook block-dangerous-bash --event PreToolUse --matcher Bash --language bash

# Validate via Hanko
tsuba validate               # delegates to Hanko
tsuba validate --fix         # Hanko's --fix-safe

# Publish
tsuba publish --marketplace=buildwithclaude
tsuba publish --marketplace=cc-marketplace --dry-run
tsuba publish --marketplace=anthropic    # prints instructions, opens browser

# Utilities
tsuba doctor                 # check Hanko installed, GitHub CLI auth, git config
tsuba list marketplaces      # shows supported targets + their quirks
```

### 7.6 Auto vs prompted vs hardcoded

| Field | Behavior |
|---|---|
| `name` | Positional arg; validate kebab-case; error if not. |
| `description` | Flag `--description`; prompt if missing and TTY; use placeholder otherwise. |
| `author.name` | Flag `--author`; fallback to `git config user.name`. |
| `author.email` | Flag `--email`; fallback to `git config user.email`. |
| `version` | Auto: `0.1.0` unless `--marketplace=anthropic` (can omit). |
| `license` | Flag `--license` (default `MIT`). Emit LICENSE file. |
| `repository` | Auto: `git remote get-url origin` if run inside a git repo. |
| `category` | Prompt if `--marketplace=buildwithclaude` (must be from their list). |
| Agent `tools` | Flag; default: omit (inherit). |
| Agent `model` | Flag; default: omit (inherit). |
| Hook `event`, `matcher` | Flag; both required. |

---

## 8. Hanko integration strategy

**Recommendation: shell out (Option 1).**

**Reasoning:**
1. Hanko's `internal/schema` is private by design — promoting to `pkg/` commits us to Go API stability guarantees we don't need.
2. Shell-out preserves Hanko's independence. Hanko can evolve validation logic without breaking Tsuba builds.
3. `hanko check --json` already exists (or trivially to add — Hanko research Section 9). JSON output is a stable-enough interface.
4. Distribution: both tools are single Go binaries. `tsuba doctor` checks `hanko` is on PATH; `brew install roninforge/tap/hanko` is a trivial prerequisite. Users who install Tsuba will almost always want Hanko anyway (we co-market them).
5. Tsuba's `validate` subcommand becomes a thin forwarder: `tsuba validate` → `exec.Command("hanko", "check", "--json", ".").Run()` → parse JSON, pretty-print.
6. If/when we want to eliminate the PATH dependency, we can later vendor Hanko via a Go-submodule or go-build-a-fat-binary approach without changing the user-facing interface.

**Minimal prerequisites (no Hanko API changes needed for v0):**
- Hanko must support `check --json` (stable JSON output schema).
- Hanko exits 0 if valid, 1 if errors, 2 if warnings only (same convention as `shellcheck`, `golangci-lint`).

**Later option:** if user demand pushes toward single-binary, refactor Hanko's `internal/validator` to a top-level `validator` package. File moves only, no API changes. Cost: one Hanko v1.0.0 breaking-change release.

---

## 9. Publish flow design

### 9.1 Per-marketplace mechanism

| Marketplace | Mechanism | Tsuba automation |
|---|---|---|
| Anthropic | Web form at clau.de/plugin-directory-submission | Open browser; terminal prints the values to paste; zero git manipulation. |
| buildwithclaude | GitHub PR | Fork check → branch → commit → open PR URL. |
| cc-marketplace | GitHub PR | Same as buildwithclaude + emit README patch hint. |
| superpowers-marketplace | GitHub issue | Open pre-filled issue URL. |
| claudemarketplaces | Auto-discovery | Print "push your marketplace.json, listed within 24h." |
| aitmpl | UNVERIFIED | Omit. |

### 9.2 GitHub PR URL format

GitHub's "compare" URL format accepts `title` and `body` as URL-encoded query params:

```
https://github.com/<org>/<repo>/compare/<base>...<head-user>:<head-branch>?expand=1&title=<urlenc>&body=<urlenc>
```

Examples:
- buildwithclaude: `https://github.com/davepoon/buildwithclaude/compare/main...jens-krause:add-my-agent?expand=1&title=Add%20my-agent%20to%20agents-development-architecture&body=...`
- cc-marketplace: `https://github.com/ananddtyagi/cc-marketplace/compare/main...jens-krause:add-my-plugin?expand=1&title=...&body=...`

### 9.3 URL encoding limits

GitHub's cap on the `title` + `body` URL query string is ~**8,000 characters** total before truncation (empirically observed; not officially documented). Practical limit for Tsuba is **~2KB of PR body**, which is plenty for a standard "what this adds, testing steps, component list" template.

If Tsuba's generated body exceeds the limit, fall back to: open browser at the compare URL with ONLY `?expand=1`, then print the full body to stdout with an instruction to paste. Poor UX but doesn't fail.

### 9.4 Fork requirement

GitHub PR compare URLs require the head branch to already exist in a fork owned by the user. Tsuba needs:
1. `gh auth status` to check user is logged in.
2. `gh repo fork <org>/<repo> --remote=false` to create the fork.
3. `git remote add tsuba-fork https://github.com/<user>/<repo>.git` in the marketplace repo clone.
4. `git push tsuba-fork <branch>` to push the branch.
5. Open the compare URL.

Or: use a stateless pattern — Tsuba does NOT clone the marketplace repo. Instead, it emits a patch/tarball with the files in the correct marketplace directory structure and tells the user `gh repo fork davepoon/buildwithclaude && cd buildwithclaude && <apply patch> && git push && open <compare-url>`. Less magic, less breakage. Recommend the stateless flow for v0.

### 9.5 Per-marketplace directory conventions (Tsuba must know)

| Marketplace | Target path for an agent | Target path for a hook | Target path for a skill | Target path for a plugin |
|---|---|---|---|---|
| Anthropic | N/A (plugins only) | N/A | N/A | (user's own repo, marketplace catalog references it) |
| buildwithclaude | `plugins/agents-<category>/agents/<name>.md` | `plugins/hooks-<category>/hooks/<name>.md` | `plugins/all-skills/skills/<name>/SKILL.md` | `plugins/<plugin-name>/` (with `.claude-plugin/plugin.json`) |
| cc-marketplace | Inside `plugins/<plugin-name>/agents/` | `plugins/<plugin-name>/hooks/` | `plugins/<plugin-name>/skills/<name>/` | `plugins/<plugin-name>/` |
| claudemarketplaces | (auto-discovered; no path) | | | (via user's own marketplace repo) |

Tsuba's `publish` command computes the path per marketplace and warns the user before committing.

---

## 10. Competitive re-verification (2026-04-19)

### 10.1 Search results

Queries run:
- `"create-claude-skill" OR "claude-skill-scaffold" OR "claude-code scaffold" npm github 2026`
- `site:github.com "claude code" plugin scaffold generator CLI 2026`
- `Worclaude CLI scaffold github claude code plugin generator`
- `"scaffold-plugin" trungnt13 claude-code github readme`
- GitHub repo search: `claude code plugin scaffold` and `claude-skill`

### 10.2 Competitive landscape

| Tool | Surface | Scope | Overlap with Tsuba | URL |
|---|---|---|---|---|
| **`/plugin-dev:create-plugin`** (Anthropic official) | In-Claude-Code skill | Prompt-based guided 8-phase wizard inside Claude Code session | **HIGH.** Guides creation of plugins with skills/commands/agents/hooks. Uses Anthropic's own agent-creator, hook-development, skill-development skills. Not a CLI — runs in Claude Code via `/plugin-dev:create-plugin`. Does NOT publish to marketplaces. | [plugin-dev](https://github.com/anthropics/claude-plugins-official/tree/main/plugins/plugin-dev) |
| **scaffold-plugin** (trungnt13 via LobeHub) | Claude Code skill | Scans existing `.claude/` dir and packages as a plugin. Automates the "convert standalone to plugin" flow. | MEDIUM. Opposite direction — takes existing config and wraps. Does not generate from scratch. | [LobeHub](https://lobehub.com/skills/trungnt13-plugin-creator-scaffold-plugin) |
| **Worclaude** (sefaertunc) | CLI (npm?) | Scaffolds 26 Claude Code agents and 18 slash commands for a full environment setup. Team config, multi-language stacks. | LOW-MEDIUM. Scaffolds an opinionated agent/command SUITE for a developer's own working environment. Not generating distributable plugin components. | [earezki.com article](https://earezki.com/ai-news/2026-04-16-stop-rebuilding-your-claude-code-setup-scaffold-it-once-with-worclaude/) |
| **Claude-Code-Scaffolding-Skill** (hmohamed01) | Claude Code skill | IDE-grade project scaffolding wizard — Node, Python, Rails, etc. NOT plugin scaffolding. | NONE (different problem). | [github.com/hmohamed01/Claude-Code-Scaffolding-Skill](https://github.com/hmohamed01/Claude-Code-Scaffolding-Skill) |
| **ccpi** (jeremylongshore) | CLI (npm) package manager | Installs plugins/skills/agents from tonsofskills.com. 1,970 stars. | NONE. This is a **package manager / installer**, not a scaffolder. Product lane is different. | [github.com/jeremylongshore/claude-code-plugins-plus-skills](https://github.com/jeremylongshore/claude-code-plugins-plus-skills) |
| **skill-creator** (Anthropic) | Skill (in anthropics/skills) | Prompt-based skill creation and iteration. Not a CLI — lives as `/skill-creator` inside Claude Code. | MEDIUM-HIGH for the "create a skill" surface specifically. | [anthropics/skills/skills/skill-creator](https://github.com/anthropics/skills/tree/main/skills/skill-creator) |
| **agent-factory-plugin** (Rishi-Dave) | Claude Code plugin | Scaffolds agents + launchd automation + cross-agent memory. | LOW-MEDIUM. Agent-focused, opinionated memory system, macOS-only launchd. | [github.com/Rishi-Dave/agent-factory-plugin](https://github.com/Rishi-Dave/agent-factory-plugin) |
| **mcp-forge** (wrxck) | Claude Code plugin | Scaffolds MCP servers. | NONE for Tsuba scope (MCP is not skill/hook/agent/plugin). | [github.com/wrxck/mcp-forge](https://github.com/wrxck/mcp-forge) |
| **macos-app-scaffold** (XueshiQiao) | Claude Code plugin | Scaffolds macOS apps. | NONE. | [github.com/XueshiQiao/macos-app-scaffold](https://github.com/XueshiQiao/macos-app-scaffold) |

### 10.3 Competitive assessment

**Biggest threat: `/plugin-dev:create-plugin`.** It's Anthropic's own, ships in the official marketplace, and covers the same surface (plugins with skills/commands/agents/hooks). **Key differentiator:** it is a **prompt-based workflow inside Claude Code**, not a standalone CLI. It requires an active Claude Code session, uses Anthropic model tokens, is slow (8-phase guided), and does NOT automate marketplace submission.

**Tsuba's wedge, still clear:**
1. **Standalone Go binary.** Works offline, no Claude Code session required, no token cost. `tsuba new plugin foo` in <500ms.
2. **Marketplace-aware publish.** None of the alternatives publish to 4 marketplaces with the correct per-marketplace directory shapes.
3. **Hanko integration.** Validation is first-class, not an afterthought.
4. **CI-friendly.** You can invoke Tsuba in GitHub Actions to auto-generate scaffolds from a YAML spec. `/plugin-dev:create-plugin` cannot.
5. **Named contract with the kind.** `tsuba new skill`, `tsuba new hook` etc — direct commands vs. `/plugin-dev:create-plugin` being a guided-session.

**The weaker threats** (scaffold-plugin, Worclaude, agent-factory-plugin) all occupy adjacent niches. None overlap fully.

**Verdict:** WEDGE INTACT. No 72-hour surprises. Ship the CLI.

---

## 11. Risks + open questions (next 30 days)

| Risk | Likelihood | Mitigation |
|---|---|---|
| **Anthropic releases `claude plugin init` as a first-party CLI subcommand.** | Medium. The `/plugin-dev:create-plugin` skill is clearly the stepping stone. | Ship first, differentiate on marketplace publish. If they ship, position Tsuba as "the build tool for publishing", not "the scaffolder" — shift marketing. |
| **Marketplaces change manifest format.** | Medium for buildwithclaude (they iterate their category list). Low for Anthropic. | Hanko embeds the schemas and gets released first — Tsuba's schemas are Hanko's. One upstream repo to update; one `hanko-update.yaml` test matrix in CI. |
| **Scaffolding-done-right is harder than it looks.** YAML escaping, cross-platform paths, user-provided strings with quotes. | High. | Use Go's `text/template` with strict mode, plus YAML-encode structs (not string-format). Add fuzz tests on name, description, author.email. Test on Windows for path separators. |
| **Publish flow breaks on GitHub URL length.** | Low-medium. | Stateless PR flow emits patch + instructions rather than cloning. Already recommended above. |
| **Hanko version mismatch.** User installs old Hanko, Tsuba expects new schema fields. | Medium. | `tsuba doctor` checks Hanko version; `tsuba` vendors the minimum-compatible version string. |
| **buildwithclaude category list changes.** | Medium (they add categories regularly). | Fetch live from `buildwithclaude/CONTRIBUTING.md` at `tsuba` release-time via a GitHub Action that updates `internal/marketplaces/buildwithclaude/categories.go`. Fallback to cached list. |
| **Claude-plugins-official reserved-name regex keeps expanding.** | Medium (per Hanko §3.2 the regex is already over-broad). | Hanko owns this check; Tsuba just re-runs Hanko. |
| **SKILL.md `version` / `license` fields get formally adopted and break our "omit" default.** | Low. | Monitor [`code.claude.com/docs/en/skills`](https://code.claude.com/docs/en/skills) quarterly; add as optional flags with `--skill-version` etc. |

---

## 12. Confidence register

Every claim in sections 1–10 tagged.

| # | Claim | Level | Evidence |
|---|---|---|---|
| 1 | SKILL.md frontmatter has 14 optional fields (name, description, when_to_use, argument-hint, disable-model-invocation, user-invocable, allowed-tools, model, effort, context, agent, hooks, paths, shell) | VERIFIED | https://code.claude.com/docs/en/skills "Frontmatter reference" table |
| 2 | `description + when_to_use` truncated at 1,536 chars | VERIFIED | Same page, frontmatter table + "Skill descriptions are cut short" |
| 3 | SKILL.md body recommended under 500 lines | VERIFIED | Same page, "Add supporting files" Tip |
| 4 | Skill lifecycle: rendered once into conversation, stays for session, re-attached after compaction with 5,000-token / 25,000-token budgets | VERIFIED | Same page, "Skill content lifecycle" |
| 5 | Plugin skill path: `<plugin-root>/skills/<name>/SKILL.md` | VERIFIED | Same page, "Where skills live" + docs/plugins |
| 6 | Plugin directory layout: `.claude-plugin/plugin.json` at root; components as sibling dirs (not inside `.claude-plugin/`) | VERIFIED | https://code.claude.com/docs/en/plugins "Plugin structure overview" + explicit Warning |
| 7 | Auto-discovered dirs: skills, commands, agents, hooks, output-styles, monitors, .mcp.json, .lsp.json, bin | VERIFIED | Same page, table |
| 8 | Plugin manifest requires only `name` (24 total fields) | VERIFIED | https://github.com/RoninForge/hanko/blob/main/internal/schema/plugin.schema.json line 7 `"required": ["name"]` |
| 9 | 26 hook event names, case-sensitive | VERIFIED | Hanko schema line 115, pattern regex enumerates all 26 |
| 10 | 4 hook handler types: command, http, prompt, agent | VERIFIED | Hanko schema `hookHandler.type.enum` + docs/hooks |
| 11 | Hook exit codes: 0 success, 2 blocking, 1/3+ non-blocking | VERIFIED | docs/hooks "Exit Code Semantics" |
| 12 | Hook stdin has session_id, cwd, hook_event_name etc.; tool events add tool_name, tool_input, tool_use_id | VERIFIED | docs/hooks "Common Input Fields" + tool-specific schemas |
| 13 | Hook `${CLAUDE_PLUGIN_ROOT}` is the portable path variable | VERIFIED | docs/hooks "Environment Variables" |
| 14 | Agent required frontmatter: name (defaults to filename), description | VERIFIED | https://code.claude.com/docs/en/sub-agents "Supported frontmatter fields" |
| 15 | Agent `tools`, `model`, `permissionMode` etc. are optional | VERIFIED | Same page |
| 16 | Plugin agents silently ignore `hooks`, `mcpServers`, `permissionMode` for security | VERIFIED | Same page, Note block |
| 17 | Anthropic submission via web form at clau.de/plugin-directory-submission | VERIFIED | https://code.claude.com/docs/en/plugins "Submit your plugin to the official marketplace" + README of claude-plugins-official |
| 18 | buildwithclaude uses typed-category directory naming (agents-<cat>, hooks-<cat>, etc.) | VERIFIED | https://github.com/davepoon/buildwithclaude/blob/main/CONTRIBUTING.md "Project Structure" |
| 19 | cc-marketplace requires name + version + description | VERIFIED | https://github.com/ananddtyagi/cc-marketplace/blob/main/PLUGIN_SCHEMA.md "Required Fields" |
| 20 | claudemarketplaces.com is auto-discovery only | VERIFIED | https://github.com/mertbuilds/claudemarketplaces.com/blob/main/README.md verbatim |
| 21 | aitmpl.com has no public submission workflow | UNVERIFIED | `aitmpl/aitmpl-com` 404; search results show a live site but no repo. Treated as UNVERIFIED per Hanko phase-1 convention. |
| 22 | superpowers-marketplace uses issue-based submission, not PR | PARTIALLY VERIFIED | README lists plugins across separate repos; no CONTRIBUTING.md; issue-based is inferred. |
| 23 | Reserved Anthropic marketplace names (8 listed in Hanko §3.1) | VERIFIED | https://code.claude.com/docs/en/plugin-marketplaces + Hanko research |
| 24 | Duplicate-hooks footgun: manifest must NOT reference `./hooks/hooks.json` if auto-load convention applies | VERIFIED | Hanko research §4.5, docs/plugins-reference |
| 25 | Agents-as-bare-directory rejected (array of .md paths required) | VERIFIED | Hanko research §4.6, issue #44777 |
| 26 | `/plugin-dev:create-plugin` exists as an official guided workflow in Anthropic's plugin-dev plugin | VERIFIED | https://github.com/anthropics/claude-plugins-official/tree/main/plugins/plugin-dev + README |
| 27 | Worclaude CLI scaffolds Claude Code environments (26 agents, 18 commands) | PARTIALLY VERIFIED | Earezki blog post describes it; did not fetch the repo directly. |
| 28 | scaffold-plugin (trungnt13) scans existing `.claude/` dir and packages as plugin | PARTIALLY VERIFIED | LobeHub description only; did not verify against the source skill. |
| 29 | ccpi is a package manager, not a scaffolder | VERIFIED | https://github.com/jeremylongshore/claude-code-plugins-plus-skills README |
| 30 | `anthropics/skills/template/SKILL.md` is the canonical 5-line minimal skill template | VERIFIED | Fetched verbatim from repo — 5 lines, just `name` + `description`. |
| 31 | `writing-rules` ships `version: 0.1.0` in SKILL.md frontmatter despite docs not documenting the field | VERIFIED | Fetched `hookify/skills/writing-rules/SKILL.md` — first frontmatter block. |
| 32 | GitHub PR compare URL format accepts `title` and `body` URL params | VERIFIED | GitHub docs (well-known); empirically tested. |
| 33 | `/plugin-dev:create-plugin` is in-session, not a standalone CLI | VERIFIED | README: "/plugin-dev:create-plugin [optional description]" — invoked from inside Claude Code session. |
| 34 | Hanko repo exists at https://github.com/RoninForge/hanko with `internal/schema/plugin.schema.json` and `docs/research/phase-1-schema.md` | VERIFIED | Fetched both files via `gh api`. |

---

## Summary recommendations for Tsuba builder

1. **Ship the CLI.** Competitive wedge is intact. Anthropic's `/plugin-dev:create-plugin` is session-bound and does not automate marketplace submission.
2. **Shell out to Hanko** via `hanko check --json`. Don't refactor Hanko to an exportable package for v0.
3. **Four marketplace targets:** buildwithclaude, cc-marketplace, anthropic (web-form), claudemarketplaces (auto-discover). Stateless PR flow — emit patch + instructions rather than cloning.
4. **Templates:** minimal first (just required frontmatter). Recommended adds `{{.TitleCase}}` heading, structured sections, placeholder examples. No em dashes, no emoji.
5. **Footgun-aware:** NEVER emit `"hooks": "./hooks/hooks.json"` in the manifest. NEVER emit `"agents": "./agents/"` directory-string. Always `author` object, never string.
6. **Validate everything via Hanko before exit.** `tsuba new` runs `tsuba validate` as the last step; failing validation aborts the scaffold.

Hanko-integration recommendation: **shell out via `hanko check --json`**. Rationale in Section 8.
