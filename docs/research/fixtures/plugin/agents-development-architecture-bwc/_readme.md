# agents-development-architecture (buildwithclaude)

**Source:** https://github.com/davepoon/buildwithclaude/tree/main/plugins/agents-development-architecture

**Type:** Typed-category agent bundle per buildwithclaude's convention.

**Shape notes:**
- Directory NAME must start with `agents-<category>` per buildwithclaude CONTRIBUTING.md.
- Manifest has `version` (buildwithclaude expects it — cc-marketplace-style strictness).
- `author` uses `{name, url}` form, no email.
- `keywords` lists each agent name (not semantic keywords) — maybe an indexing hack.
- Agent frontmatter requires `category:` field matching one of buildwithclaude's allowed categories (e.g. `development-architecture`).
- Agents are flat `.md` files in `agents/` directory (no subdirectories).

**What Tsuba should copy per --marketplace=buildwithclaude:** directory-name prefix `<kind>-<category>`, mandatory `version`, `category` field in agent/command frontmatter, bundle convention (one plugin = one typed category of agents).
