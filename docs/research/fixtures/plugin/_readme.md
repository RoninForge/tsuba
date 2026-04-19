# Plugin fixtures

Full plugin directories (manifest + minimal supporting files) demonstrating the three live layouts.

| Directory | Source | Pattern |
|---|---|---|
| `example-plugin-anthropic/` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/tree/main/plugins/example-plugin) | Anthropic reference plugin. Minimal 3-field manifest. |
| `hookify-anthropic/` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official/tree/main/plugins/hookify) | Hooks-focused plugin, demonstrates auto-discovered `hooks/hooks.json` without manifest reference. |
| `agents-development-architecture-bwc/` | [davepoon/buildwithclaude](https://github.com/davepoon/buildwithclaude/tree/main/plugins/agents-development-architecture) | buildwithclaude typed-category bundle. |

Full directory trees available in the source repos; only `plugin.json` saved here because Tsuba's output is primarily manifest + directory structure, not transcribed component content.
