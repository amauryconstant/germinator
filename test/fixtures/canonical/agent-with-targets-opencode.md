---
name: agent-with-opencode-targets
description: Agent demonstrating OpenCode-specific targets
tools:
  - bash
  - grep
  - read
  - edit
model: anthropic/claude-sonnet-4-20250514
permissionPolicy: balanced
behavior:
  mode: primary
  temperature: 0.5
  steps: 15
targets:
  opencode:
    temperature: 0.8
    steps: 25
    prompt: "Override prompt for OpenCode"
---
This agent demonstrates using targets section for OpenCode-specific configuration.

When rendered for Claude Code:
- targets.opencode section is omitted
- Uses default behavior.temperature: 0.5 and behavior.steps: 15

When rendered for OpenCode:
- Uses targets.opencode.temperature: 0.8 instead of behavior.temperature
- Uses targets.opencode.steps: 25 instead of behavior.steps
- Uses targets.opencode.prompt as override
