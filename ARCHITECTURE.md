# Architecture

This document contains technical reference material for Germinator developers.

## Field Mappings

### Agent

| Germinator Field | Claude Code | OpenCode                                         |
| ---------------- | ----------- | ------------------------------------------------ |
| name             | ✓           | ⚠ omitted (uses filename as identifier)          |
| description      | ✓           | ✓                                                |
| model            | ✓           | ✓ (full provider-prefixed ID)                    |
| tools            | ✓           | ✓ (converted to lowercase)                       |
| disallowedTools  | ✓           | ✓ (converted to lowercase, set false)            |
| permissionMode   | ✓           | → Permission object (nested with ask/allow/deny) |
| skills           | ✓           | ⚠ (skipped - not supported)                      |
| mode             | -           | ✓ (primary/subagent/all, defaults to all)        |
| temperature      | -           | ✓ (\*float64 pointer, omits when nil)            |
| maxSteps         | -           | ✓                                                |
| hidden           | -           | ✓ (omits when false)                             |
| prompt           | -           | ✓                                                |
| disable          | -           | ✓ (omits when false)                             |

### Command

| Germinator Field         | Claude Code | OpenCode                      |
| ------------------------ | ----------- | ----------------------------- |
| name                     | ✓           | ✓                             |
| description              | ✓           | ✓                             |
| allowed-tools            | ✓           | ⚠ (skipped - not supported)   |
| disallowed-tools         | ✓           | ⚠ (skipped - not supported)   |
| subtask                  | ✓           | ✓                             |
| argument-hint            | ✓           | ⚠ (skipped - not supported)   |
| context                  | ✓ (fork)    | ✓ (fork)                      |
| agent                    | ✓           | ✓                             |
| model                    | ✓           | ✓ (full provider-prefixed ID) |
| disable-model-invocation | ✓           | ⚠ (skipped - not supported)   |

### Skill

| Germinator Field | Claude Code | OpenCode                      |
| ---------------- | ----------- | ----------------------------- |
| name             | ✓           | ✓                             |
| description      | ✓           | ✓                             |
| allowed-tools    | ✓           | ⚠ (skipped - not supported)   |
| disallowed-tools | ✓           | ⚠ (skipped - not supported)   |
| license          | ✓           | ✓                             |
| compatibility    | ✓           | ✓                             |
| metadata         | ✓           | ✓                             |
| hooks            | ✓           | ✓                             |
| model            | ✓           | ✓ (full provider-prefixed ID) |
| context          | ✓ (fork)    | ✓ (fork)                      |
| agent            | ✓           | ✓                             |
| user-invocable   | ✓           | ⚠ (skipped - not supported)   |

### Memory

| Germinator Field | Claude Code | OpenCode                             |
| ---------------- | ----------- | ------------------------------------ |
| paths            | ✓           | → @ file references (one per line)   |
| content          | ✓           | → Narrative context (rendered as-is) |

## Known Limitations

### Permission Mode Transformation

The transformation from Claude Code's `permissionMode` enum to OpenCode's permission object is a basic approximation:

- `default` → `{edit: {"*": "ask"}, bash: {"*": "ask"}}`
- `acceptEdits` → `{edit: {"*": "allow"}, bash: {"*": "ask"}}`
- `dontAsk` → `{edit: {"*": "allow"}, bash: {"*": "allow"}}`
- `bypassPermissions` → `{edit: {"*": "allow"}, bash: {"*": "allow"}}`
- `plan` → `{edit: {"*": "deny"}, bash: {"*": "deny"}}`

Only `edit` and `bash` tools are mapped (7+ other OpenCode permissionable tools remain undefined). Command-level permission rules are not supported.

### Skipped Fields

The following fields are not supported in OpenCode and are silently skipped:

- **Agent**: `skills`
- **Command**: `disableModelInvocation`, `argumentHint`, `allowedTools`, `disallowedTools`
- **Skill**: `userInvocable`, `allowedTools`, `disallowedTools`

### DisallowedTools Forward Compatibility

OpenCode does not support `disallowedTools` in agents. Fields are included for forward compatibility but not used in current transformations.

### Unidirectional Transformation

Transformation is one-way: Germinator format → target platform only. There is no support for bidirectional conversion.
