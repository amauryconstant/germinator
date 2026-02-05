---
name: agent-with-claude-code-targets
description: Agent demonstrating Claude Code-specific targets
tools:
  - bash
  - grep
  - read
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: balanced
behavior:
  mode: primary
  temperature: 0.7
  steps: 20
targets:
  claude-code:
    skills:
      - coding
      - git-workflow
    disable-model-invocation: false
---
This agent demonstrates using targets section for Claude Code-specific configuration.

When rendered for Claude Code:
- Skills array is rendered from targets.claude-code.skills
- disable-model-invocation field is rendered from targets.claude-code

When rendered for OpenCode:
- targets.claude-code section is omitted
- Only common fields are used
