# Claude Code Documentation

Platform documentation for AI coding assistant configuration

## Official Documentation Sources

- [Overview](https://code.claude.com/docs/en/overview) - General platform overview
- [Skills](https://code.claude.com/docs/en/skills.md) - Custom skills with SKILL.md format
- [Memory](https://code.claude.com/docs/en/memory.md) - Memory management with CLAUDE.md
- [Sub-agents](https://code.claude.com/docs/en/sub-agents.md) - Custom subagents
- [Settings](https://code.claude.com/docs/en/settings.md) - Configuration and permissions
- [CLI Reference](https://code.claude.com/docs/en/cli-reference.md) - Command-line interface
- [Interactive Mode](https://code.claude.com/docs/en/interactive-mode.md) - Built-in commands and shortcuts

---

## Document Types

### Skills

**File:** `.claude/skills/<name>/SKILL.md`

**Frontmatter Fields:**

| Field                    | Type          | Required    | Description                                                                  |
| ------------------------ | ------------- | ----------- | ---------------------------------------------------------------------------- |
| name                     | string        | No          | Display name for skill, lowercase, max 64 chars (defaults to directory name) |
| description              | string        | Recommended | What skill does and when to use it                                           |
| argument-hint            | string        | No          | Hint shown during autocomplete, e.g., `[issue-number]`                       |
| disable-model-invocation | boolean       | No          | Prevent Claude from auto-loading (default: `false`)                          |
| user-invocable           | boolean       | No          | Hide from `/` menu (default: `true`)                                         |
| allowed-tools            | array[string] | No          | Tools Claude can use without approval                                        |
| model                    | string        | No          | Model to use when skill is active                                            |
| context                  | string        | No          | Set to `fork` to run in subagent context                                     |
| agent                    | string        | No          | Which subagent type when `context: fork`                                     |
| hooks                    | object        | No          | Lifecycle hooks scoped to skill                                              |

**Body:** Markdown content with skill instructions

**String Substitutions:**

- `$ARGUMENTS` - All arguments passed
- `$ARGUMENTS[N]` or `$N` - Specific argument by position
- `${CLAUDE_SESSION_ID}` - Current session ID

---

### Subagents (Agents)

**File:** `.claude/agents/<name>.md`

**Frontmatter Fields:**

| Field           | Type          | Required | Description                                                      |
| --------------- | ------------- | -------- | ---------------------------------------------------------------- |
| name            | string        | Yes      | Unique identifier, lowercase with hyphens                        |
| description     | string        | Yes      | When Claude should delegate to this subagent                     |
| tools           | array[string] | No       | Tools subagent can use (PascalCase)                              |
| disallowedTools | array[string] | No       | Tools to deny (PascalCase)                                       |
| model           | string        | No       | Model: `sonnet`, `opus`, `haiku`, or `inherit` (default)         |
| permissionMode  | string        | No       | `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, `plan` |
| skills          | array[string] | No       | Skills to preload into subagent's context                        |
| hooks           | object        | No       | Lifecycle hooks scoped to subagent                               |

**Body:** Markdown content as system prompt

---

### Memory

**Files:**

- `CLAUDE.md` (multiple locations)
- `.claude/CLAUDE.md`
- `.claude/rules/*.md`

**Frontmatter:** None - Pure markdown content with optional file references

**Features:**

- `@path/to/file` syntax for importing files
- Recursive imports supported (max 5 hops)
- Path-specific rules in `.claude/rules/` with `paths` frontmatter field

**Locations (by precedence):**

1. Managed policy: System-level CLAUDE.md
2. Project: `./CLAUDE.md` or `./.claude/CLAUDE.md`
3. Project rules: `./.claude/rules/*.md`
4. User: `~/.claude/CLAUDE.md`
5. Local: `./CLAUDE.local.md`

**Frontmatter for `.claude/rules/*.md`:**

| Field | Type          | Required | Description                         |
| ----- | ------------- | -------- | ----------------------------------- |
| paths | array[string] | No       | Glob patterns for conditional rules |

---

### Settings

**File:** `settings.json` (multiple scopes)

**Permission Settings:**

| Key         | Type          | Description                                  |
| ----------- | ------------- | -------------------------------------------- |
| allow       | array[string] | Permission rules to allow                    |
| ask         | array[string] | Permission rules to ask for confirmation     |
| deny        | array[string] | Permission rules to deny                     |
| defaultMode | string        | Default permission mode: `acceptEdits`, etc. |

**Other Settings:**

| Key            | Type   | Description                          |
| -------------- | ------ | ------------------------------------ |
| permissions    | object | Permission configuration (see above) |
| env            | object | Environment variables                |
| attribution    | object | Git commit/PR attribution settings   |
| hooks          | object | Pre/post tool hooks                  |
| model          | string | Override default model               |
| statusLine     | object | Custom status line config            |
| fileSuggestion | object | Custom file autocomplete             |
| outputStyle    | string | Output style configuration           |
| language       | string | Preferred response language          |

---

## Permission System

### Rule Syntax

`Tool` or `Tool(specifier)`

### Evaluation Order

1. Deny rules (checked first)
2. Ask rules (checked second)
3. Allow rules (checked last)

### Permission Modes

| Mode              | Description                                                  |
| ----------------- | ------------------------------------------------------------ |
| default           | Standard permission checking with prompts                    |
| acceptEdits       | Auto-accept file edits                                       |
| dontAsk           | Auto-deny permission prompts (explicitly allowed tools work) |
| bypassPermissions | Skip all permission checks                                   |
| plan              | Plan mode (read-only exploration)                            |

### Tool Names (PascalCase)

- `Bash` - Execute shell commands
- `Read` - Read file contents
- `Write` - Create/overwrite files
- `Edit` - Modify files (exact string replacement)
- `Grep` - Search file contents
- `Glob` - Find files by pattern
- `List` - List directory contents
- `Patch` - Apply patch files
- `WebFetch` - Fetch web content
- `Task` - Launch subagents
- `Skill` - Load skills
- `TodoRead`, `TodoWrite` - Manage todo lists
- `Question` - Ask user questions

### Specifier Examples

- `Bash` - All bash commands
- `Bash(npm run build)` - Exact command match
- `Bash(*)` - Equivalent to Bash
- `Bash(git *)` - Git commands with any arguments
- `Read(./.env)` - Specific file
- `WebFetch(domain:example.com)` - Specific domain

### Wildcards

- `*` matches zero or more characters
- `?` matches exactly one character

---

## Tool Configuration

### Configuration

- Via `tools` field in subagent
- Via `allowedTools` in skill frontmatter
- Via `--tools` CLI flag
- Via `--allowedTools` CLI flag

---

## Model Identifiers

### Format

Full model name or alias

### Aliases

- `sonnet` - Use Sonnet model
- `opus` - Use Opus model
- `haiku` - Use Haiku model
- `inherit` - Inherit default model (default)

### Full Names

- `claude-sonnet-4-5-20250929` - Specific version

### Examples

```yaml
model: sonnet              # Alias
model: opus                # Alias
model: claude-sonnet-4-5-20250929  # Full name
model: inherit             # Inherit default
```

---

## Validation Constraints

### Skill Names

- Lowercase letters, numbers, hyphens only
- Max 64 characters
- No consecutive hyphens
- No starting/ending with hyphen

### Subagent Names

- Lowercase letters and hyphens
- Must be unique

### Permissions

- Rule evaluation is order-dependent
- Deny rules always take precedence
- Patterns are simple wildcards (`*`, `?`)

### File Paths

- Imports support recursive loading (max 5 hops)
- Home directory expansion: `~` supported

---

## YAML Examples

### Skill Example

```yaml
---
name: git-release
description: Create consistent releases and changelogs
disable-model-invocation: true
allowed-tools: Bash(gh *)
context: fork
agent: general-purpose
---

## What I do
- Draft release notes from merged PRs
- Propose a version bump
- Provide a copy-pasteable `gh release create` command

## When to use me
Use this when you are preparing a tagged release.
Ask clarifying questions if target versioning scheme is unclear.
```

### Subagent Example

```yaml
---
name: code-reviewer
description: Expert code review specialist. Proactively reviews code for quality, security, and maintainability.
tools: Read, Grep, Glob, Bash
model: sonnet
permissionMode: default
---

You are a senior code reviewer ensuring high standards of code quality and security.

When invoked:
1. Run git diff to see recent changes
2. Focus on modified files
3. Begin review immediately

Review checklist:
- Code is clear and readable
- Functions and variables are well-named
- No duplicated code
- Proper error handling
- No exposed secrets or API keys
- Input validation implemented
- Good test coverage
- Performance considerations addressed

Provide feedback organized by priority:
- Critical issues (must fix)
- Warnings (should fix)
- Suggestions (consider improving)

Include specific examples of how to fix issues.
```

### Settings Example

```json
{
  "permissions": {
    "allow": ["Bash(npm run lint)", "Bash(npm run test *)", "Read(~/.zshrc)"],
    "deny": [
      "Bash(curl *)",
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)"
    ]
  },
  "env": {
    "FOO": "bar"
  }
}
```


