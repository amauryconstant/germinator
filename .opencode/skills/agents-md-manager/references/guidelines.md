# AGENTS.md Guidelines

Based on best practices for AGENTS.md files, applicable to OpenCode.

## Core Principle: Ruthless Conciseness

**Golden rule**: For each line, ask "Would removing this cause opencode to make mistakes?" If not, cut it.

**Why this matters**:
- AGENTS.md is loaded every session and consumes context
- Bloated files cause opencode to ignore your actual instructions
- Long files dilute important rules in noise
- Every line costs tokens; context window is precious

## What to Include

✅ **Bash commands opencode can't guess**
- Non-standard build commands
- Project-specific scripts
- Custom tooling workflows

✅ **Code style rules that differ from defaults**
- Non-standard indentation
- Import conventions
- Naming patterns unique to your project

✅ **Testing instructions and preferred test runners**
- How to run tests
- Preferred testing patterns
- Coverage requirements

✅ **Repository etiquette**
- Branch naming conventions
- PR conventions
- Commit message format

✅ **Architectural decisions specific to your project**
- Non-obvious design patterns
- Technology choices and rationale
- Integration patterns

✅ **Developer environment quirks**
- Required environment variables
- Local setup gotchas
- Tool-specific configurations

✅ **Common gotchas or non-obvious behaviors**
- Edge cases in the codebase
- Known limitations
- Workarounds for specific issues

## What to Exclude

❌ **Anything opencode can figure out by reading code**
- Standard library usage
- Common patterns
- Self-evident code organization

❌ **Standard language conventions opencode already knows**
- Basic Python PEP 8
- JavaScript/TypeScript defaults
- Standard REST conventions

❌ **Detailed API documentation (link to docs instead)**
- Full API references
- Complete library documentation
- External service APIs

❌ **Information that changes frequently**
- Specific version numbers
- Temporary workarounds
- Short-lived feature flags

❌ **Long explanations or tutorials**
- How-to guides
- Step-by-step tutorials
- Educational content

❌ **File-by-file descriptions of the codebase**
- Directory walkthroughs
- Module-by-module explanations
- Detailed file listings

❌ **Self-evident practices**
- "Write clean code"
- "Use meaningful variable names"
- "Add comments where needed"

## Structure Guidelines

### Markdown Formatting
- Use headings for hierarchy
- Use lists for scannable content
- Use code blocks for commands/examples
- Keep sections focused and short

### Emphasis for Critical Rules
If opencode keeps ignoring a rule, add emphasis:
- **IMPORTANT**
- **YOU MUST**
- **NEVER**
- Bold or capitalization for critical constraints

### Import Syntax
Use `@path/to/file` for modularity:

```markdown
# Architecture
@docs/architecture.md

# Coding Standards
@docs/typescript-conventions.md
@docs/react-patterns.md

# Personal Overrides
@~/.config/opencode/my-project-instructions.md
```

### Location-Based Hierarchy

Files are loaded in hierarchical order:
1. User memory (`~/.config/opencode/AGENTS.md`)
2. Project root (`./AGENTS.md` or `./.opencode/AGENTS.md`)
3. Parent directories (walks up from cwd)
4. Child directories (loaded on-demand when working in subdirectories)

## File Locations

| Location | Purpose | Shared With |
|----------|---------|-------------|
| `~/.config/opencode/AGENTS.md` | Personal preferences (all projects) | Just you |
| `./AGENTS.md` | Project memory (team-shared) | Team via git |
| `./.opencode/AGENTS.md` | Project memory (alternate location) | Team via git |
| `./AGENTS.local.md` | Private project preferences | Just you (gitignored) |
| `./.opencode/rules/*.md` | Modular project rules | Team via git |

## Modular Rules (`.opencode/rules/`)

For larger projects, split instructions into focused files:

```
.opencode/rules/
├── code-style.md       # Code style guidelines
├── testing.md          # Testing conventions
├── security.md         # Security requirements
├── frontend/           # Frontend-specific
│   ├── react.md
│   └── styles.md
└── backend/            # Backend-specific
    ├── api.md
    └── database.md
```

### Path-Specific Rules

Use YAML frontmatter to scope rules to specific files:

```markdown
---
paths:
  - "src/api/**/*.ts"
  - "lib/api/**/*.ts"
---

# API Development Rules

- All API endpoints must include input validation
- Use the standard error response format
```

**Glob patterns supported**:
- `**/*.ts` - All TypeScript files
- `src/**/*` - All files under src/
- `*.md` - Markdown in root
- `src/**/*.{ts,tsx}` - Multiple extensions

## Quality Maintenance

### Treat Like Code
- Review when things go wrong
- Prune regularly
- Test changes by observing opencode's behavior
- Check into git for team contribution

### Track Violations
If opencode ignores a rule despite it being documented:
- The file is probably too long (rule getting lost)
- The phrasing might be ambiguous
- Consider adding emphasis or moving to a hook

### Length Guidelines
- No hard limit, but shorter is better
- If approaching 1000 lines, split into modular rules
- If opencode asks questions answered in AGENTS.md, simplify phrasing
- If opencode ignores documented rules, reduce total length

## Common Anti-Patterns

### Kitchen Sink Session
Context filled with unrelated conversation, files, commands.
**Fix**: Start fresh session between unrelated tasks

### Over-Specified AGENTS.md
File too long, important rules get lost.
**Fix**: Ruthlessly prune; if opencode does it correctly without the instruction, delete it

### Trust-Then-Verify Gap
opencode produces plausible code without handling edge cases.
**Fix**: Always provide verification (tests, scripts, screenshots)

### Redundant Information
Same information in multiple locations.
**Fix**: Use imports (`@path`) or cross-references instead of duplication

## Examples

### Good AGENTS.md

```markdown
# Code Style
- Use ES modules (import/export), not CommonJS (require)
- Destructure imports: `import { foo } from 'bar'`
- Max function length: 50 lines

# Workflow
- Typecheck after code changes: `npm run typecheck`
- Run single tests, not full suite: `npm test -- <test-name>`
- Never commit to main; use feature branches

# Environment
- Required env var: `DATABASE_URL` (see .env.example)
- Docker required for integration tests
```

### Bad AGENTS.md

```markdown
# Introduction
This is our project. It uses TypeScript and React. We follow best practices.

# File Structure
- src/ contains the source code
  - components/ has React components
    - Button.tsx is a button component
    - Input.tsx is an input component
  - utils/ has utility functions
    - date.ts has date utilities
    - string.ts has string utilities

# Coding Standards
- Write clean code
- Use meaningful variable names
- Add comments where appropriate
- Follow TypeScript best practices
- Test your code thoroughly
```

The bad example wastes tokens on:
- Self-evident information (opencode knows TypeScript/React)
- File-by-file descriptions (opencode can read the codebase)
- Generic advice opencode already follows
- Redundant explanations
