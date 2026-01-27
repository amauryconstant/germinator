---
name: agents-md-manager
description: Manage AGENTS.md file quality and content in dotfiles repository. Use when creating, updating, validating, or refactoring AGENTS.md files. Ensures compliance with best practices including conciseness, specificity, token efficiency, and proper hierarchy. Triggers on tasks like "validate AGENTS.md", "improve documentation quality", "check for anti-patterns", or "refactor AGENTS.md structure".
---

# AGENTS.md Manager

Manage AGENTS.md file quality and content for this dotfiles repository based on best practices.

## When to Use This Skill

- Creating new AGENTS.md files for new components
- Validating existing AGENTS.md files against best practices
- Refactoring AGENTS.md files to improve quality
- Checking for common anti-patterns and token waste
- Ensuring proper location-based hierarchy
- Improving specificity and reducing vagueness
- Splitting overly long files into modular structure

## Core Principles

### 1. Ruthless Conciseness
**Golden rule**: For each line, ask "Would removing this cause opencode to make mistakes?" If not, cut it.

**Why**:
- AGENTS.md loaded every session (costs tokens)
- Bloated files cause opencode to ignore instructions
- Long files dilute important rules

**Target lengths**:
- Ideal: <300 lines (~1000 tokens)
- With MCP servers: <200 lines (~600 tokens)
- Warning threshold: >500 lines (consider splitting)
- Critical threshold: >1000 lines (must split immediately)

### 2. Progressive Disclosure (Most Important)
**Core concept**: Load information on-demand, not all at once.

**Why this matters**:
- Context window fills fast (especially with MCP servers)
- Bloated context causes opencode to enter "dumb zone"
- Most documentation isn't needed for every task
- Token savings of 60-85% possible

**Three-layer approach**:

**Layer 1: Index** (always loaded)
```markdown
## Script Library
50+ scripts in lib/scripts/ organized by 10 categories.
See lib/scripts/AGENTS.md for detailed patterns.
```
Token cost: ~50 tokens

**Layer 2: Topic Details** (loaded when working in that area)
```markdown
# In lib/scripts/AGENTS.md
Categories: core, desktop, system, media, network...
See core/AGENTS.md for UI library patterns.
```
Token cost: ~200 tokens (only when working on scripts)

**Layer 3: Deep Dive** (loaded for specific work)
```markdown
# In lib/scripts/core/AGENTS.md
gum-ui.sh: 35 functions, usage patterns...
```
Token cost: ~400 tokens (only when working on core utilities)

**Total savings**: 600 tokens saved unless specifically needed

**Key pattern - "How to find it" not "Here it is"**:
✅ "See private_dot_config/hypr/AGENTS.md for Hyprland patterns"
❌ [500 lines of Hyprland documentation in root file]

### 3. Include Only What opencode Can't Infer
✅ Project-specific commands opencode can't guess
✅ Non-standard conventions
✅ Environment-specific quirks
✅ Gotchas and non-obvious behaviors

❌ Standard language conventions
❌ Generic advice ("write clean code")
❌ Self-evident information
❌ File-by-file codebase descriptions

### 4. Be Specific, Not Vague
✅ "Use 2-space indentation for TypeScript"
❌ "Format code properly"

✅ "`npm run typecheck` after code changes"
❌ "Run the typechecker"

### 5. Structure for Scannability
- Use markdown headings, lists, tables
- Front-load quick reference sections
- Group related content
- Keep sections focused (<50 lines)

## Token Budget Management

### Understanding Token Costs

**Context window**: ~200k tokens

**Rough estimation**:
- 1 line of markdown ≈ 3-5 tokens (average: 4 tokens)
- 300-line AGENTS.md ≈ 1200 tokens (0.6% of context)
- 1000-line AGENTS.md ≈ 4000 tokens (2% of context)
- 1250-line AGENTS.md ≈ 5000 tokens (2.5% of context)

### Recommended Budget Allocation

**Without MCP servers**:
- AGENTS.md files: <5% of context (<10,000 tokens combined)
- Root file: <1000 tokens (300 lines)
- Working context: Reserve 80%+ for code and conversation

**With MCP servers**:
- MCP tools can consume 66,000+ tokens (33% of context!)
- AGENTS.md budget: <3% of context (<6,000 tokens)
- Root file: <600 tokens (200 lines)
- Critical: Monitor with `/context` command

### Warning Thresholds

**Per-file limits**:
```bash
# Estimate tokens
wc -l AGENTS.md
# Example: 1250 lines × 4 tokens ≈ 5000 tokens

# Thresholds
# <1000 tokens: Excellent
# 1000-2000 tokens: Good (consider optimizing)
# 2000-4000 tokens: Warning (split recommended)
# >4000 tokens: Critical (opencode may ignore rules)
```

**Repository-wide**:
- Calculate total: All AGENTS.md files in hierarchy
- With location-based loading: Only active files count
- Progressive disclosure reduces total loaded tokens

### Measuring Your Token Budget

```bash
# Find all AGENTS.md files with line counts
find . -name "AGENTS.md" -exec wc -l {} \; | sort -n

# Calculate estimated total tokens
find . -name "AGENTS.md" -exec wc -l {} \; | \
  awk '{sum+=$1} END {print "Total lines:", sum, "| Est. tokens:", sum*4}'

# Identify largest files (optimization targets)
find . -name "AGENTS.md" -exec wc -l {} \; | sort -rn | head -5
```

### Token Savings Examples

**Case Study 1: Monorepo refactoring**
- Before: 2400-line root (9600 tokens always loaded)
- After: 300-line root + topic files (1200 tokens + on-demand loading)
- Savings: 8400 tokens (87.5%) when not working on specific topics
- Impact: opencode follows rules more consistently

**Case Study 2: Progressive disclosure**
- Before: All docs in root (5000 tokens)
- After: Index in root, details in subdirectories (800 tokens root, 4200 on-demand)
- Savings: 4200 tokens (84%) when not working in those areas
- Impact: Faster responses, better accuracy

**Case Study 3: Table optimization**
- Before: Verbose file descriptions (147 tokens per section)
- After: Table format (47 tokens per section)
- Savings: 100 tokens per section (68%)
- Scales: 10 sections = 1000 tokens saved

## MCP Context Management

### MCP Server Token Impact

**Problem**: MCP tools consume massive context before conversation starts

**Typical impact**:
- Small MCP setup: 10,000-20,000 tokens
- Medium setup: 30,000-50,000 tokens
- Large setup: 66,000+ tokens (33% of Sonnet's window!)

**Example**: Filesystem + Git + Database MCPs = 45,000 tokens
- Your AGENTS.md budget: Only 10,000 tokens remaining for <5% target
- With 1250-line root: Already at 5000 tokens (half your budget!)

### Context Monitoring

Monitor context usage regularly:
- When performance degrades
- Before major refactoring
- After adding MCP servers
- When opencode seems to "forget" instructions

### AGENTS.md Strategy with MCP

**Adjust targets**:
- Root file: <200 lines (<800 tokens) instead of 300
- Combined files: <6000 tokens instead of 10,000
- More aggressive progressive disclosure
- Eliminate all non-essential content

**Alternative: CLI tools**:
```markdown
# Instead of MCP server for git (saves ~5000 tokens)
Use `gh` CLI for GitHub operations
Use `git` commands directly

# Instead of MCP server for package management
Use `paru`, `pacman` commands directly
```

**When MCP is worth it**:
- Complex APIs (database queries, cloud services)
- Real-time data (monitoring, logs)
- Stateful interactions (auth, sessions)

**When CLI is better**:
- Simple commands (git, package managers)
- One-shot operations
- Standard Unix tools

## How This Skill Demonstrates Progressive Disclosure

This skill itself uses progressive disclosure principles:

### SKILL.md (This File)
**Always loaded when skill activates**
- Core principles (5 key concepts)
- Token budget overview
- Quick validation workflow
- Common task patterns
- ~500 lines

**Token cost**: ~2000 tokens when skill active

### Reference Files (On-Demand)
**Loaded only when needed**:

**anthropic-guidelines.md** (~800 lines)
- Load when: Checking compliance with official guidelines
- Load when: Resolving disagreements about best practices
- Load when: Need detailed include/exclude lists
- Token cost: ~3200 tokens (only when loaded)

**quality-checklist.md** (~300 lines)
- Load when: Performing manual quality review
- Load when: Need comprehensive validation checklist
- Token cost: ~1200 tokens (only when loaded)

**dotfiles-patterns.md** (~600 lines)
- Load when: Working on dotfiles-specific AGENTS.md files
- Load when: Need chezmoi/template patterns
- Token cost: ~2400 tokens (only when loaded)

### Total Token Savings

**Without progressive disclosure**:
- All content in SKILL.md: ~1900 lines
- Token cost: ~7600 tokens every time skill used

**With progressive disclosure**:
- SKILL.md: ~500 lines = ~2000 tokens (always)
- References: ~1700 lines = ~6800 tokens (on-demand)
- Average use: 2000 tokens (no references needed)
- Deep work: 4000-5000 tokens (1-2 references loaded)

**Typical savings**: 5600 tokens (73%) for routine work

### Pattern to Replicate

Apply this same structure to your repository:

**Root AGENTS.md** (always loaded):
- Architecture overview
- Critical safety protocols
- Quick reference to subdirectories
- ~300 lines = ~1200 tokens

**Subdirectory AGENTS.md** (loaded when working in that area):
- Component-specific patterns
- Integration points
- Testing procedures
- ~200 lines each = ~800 tokens each

**Result**: 1200 tokens always loaded, additional context only when needed

## Validation Workflow

### Step 1: Run Automated Validation

```bash
# Validate single file
python3 .opencode/skills/agents-md-manager/scripts/validate_agents_md.py AGENTS.md

# Validate all AGENTS.md files in repository
python3 .opencode/skills/agents-md-manager/scripts/validate_agents_md.py .
```

**Checks**:
- File length (warns >500 lines, errors >1000)
- Token estimation (calculates approximate token cost)
- Common anti-patterns (generic advice, historical background, tutorials)
- Vague language (properly, correctly, as needed)
- Markdown structure (code blocks, excessive lists)
- Progressive disclosure usage (import patterns, cross-references)

### Step 2: Manual Quality Review

Use the quality checklist: See [quality-checklist.md](references/quality-checklist.md)

**Key checks**:
- [ ] Conciseness test (every line justified)
- [ ] Specificity test (no vague instructions)
- [ ] Value test (opencode can't infer from code)
- [ ] Structure quality (scannable, organized)
- [ ] Token efficiency (tables, cross-refs, no duplication)

### Step 3: Anthropic Guidelines Compliance

Refer to official guidelines: See [anthropic-guidelines.md](references/anthropic-guidelines.md)

**Validate against**:
- Include/exclude lists
- Structure recommendations
- Import syntax usage
- Location hierarchy appropriateness

### Step 4: Dotfiles-Specific Patterns

For dotfiles repositories: See [dotfiles-patterns.md](references/dotfiles-patterns.md)

**Check**:
- Configuration management patterns documented
- Security constraints clear (encryption, secrets)
- Tool integration points specified
- Validation workflows provided
- No standard tool documentation repetition

## Common Tasks

### Creating New AGENTS.md

**Workflow**:
1. Determine appropriate location (root, subdirectory, rules/)
2. Start with minimal template:
    ```markdown
    # [Component] - OpenCode Reference

    **Location**: `/path/to/directory`
    **Parent**: See `../AGENTS.md` for overview
    **Root**: See `/root/AGENTS.md` for core standards

   **CRITICAL**: Be concise. Sacrifice grammar for token-efficiency.

   ## Quick Reference
   - Purpose: [Brief description]
   - Key files: [Main files]
   - Integration: [How it connects]

   ## [Main Content Sections]
   ```
3. Add only essential, non-inferable content
4. Validate with script
5. Manual review against checklist

### Refactoring Existing AGENTS.md

**When to refactor**:
- File >500 lines
- opencode ignores documented rules
- opencode asks questions answered in file
- Duplicate content across files
- Stale or outdated sections

**Refactoring strategies**:

**Strategy 1: Split into modular rules**
```
# Before: Single 800-line AGENTS.md
AGENTS.md (800 lines)

# After: Modular structure
.opencode/
├── AGENTS.md (200 lines - core only)
└── rules/
    ├── scripts.md
    ├── templates.md
    ├── security.md
    └── workflows.md
```

**Strategy 2: Use imports**
```markdown
# In main AGENTS.md
## Script Standards
@.opencode/rules/scripts.md

## Template System
@.opencode/rules/templates.md
```

**Strategy 3: Move to subdirectories**
```
# Before: Everything in root AGENTS.md
AGENTS.md (1200 lines covering all components)

# After: Location-based hierarchy
AGENTS.md (300 lines - architecture overview)
private_dot_config/
├── AGENTS.md (XDG overview)
├── hypr/AGENTS.md (Hyprland-specific)
└── shell/AGENTS.md (Shell-specific)
```

### Improving Specificity

**Pattern: Vague → Specific**

**Before**:
```markdown
Test your code properly before committing.
Run the appropriate validation commands.
```

**After**:
```markdown
# Pre-Commit Validation
1. `chezmoi diff` - See what will change
2. `chezmoi apply --dry-run` - Test without applying
3. `bash -n script.sh.tmpl` - Validate shell syntax
```

**Pattern: Generic → Concrete**

**Before**:
```markdown
Follow coding best practices.
Write clean, maintainable code.
```

**After**:
```markdown
# Script Standards
- Use `set -euo pipefail` for multi-step scripts
- Source UI library: `. "$UI_LIB"`
- POSIX sh for utilities, bash for complex logic
```

### Reducing Token Waste

**Use tables instead of verbose lists**:

**Before** (147 tokens):
```markdown
- monitor.conf.tmpl: This file handles display settings including
  resolution, scaling, and position. It's a template file that uses
  Go template syntax.
- environment.conf: This file sets up environment variables for
  NVIDIA, Qt/GTK, and XDG. It's not a template file.
```

**After** (47 tokens):
```markdown
| File | Purpose | Template? |
|------|---------|-----------|
| monitor.conf.tmpl | Display settings | ✅ Yes |
| environment.conf | Env vars | ❌ No |
```

**Use cross-references instead of duplication**:

**Before**:
```markdown
# In multiple files: 400 lines of UI library docs duplicated
```

**After**:
```markdown
# In each file:
**UI Library**: See `core/AGENTS.md` for gum-ui.sh reference
```

**Front-load essentials**:

**Before**:
```markdown
[500 lines of detailed content]
## Summary
Key points buried at end
```

**After**:
```markdown
## Quick Reference
- Purpose: [Brief]
- Reload: [Command]
- Files: [Count]

[Detailed content follows]
```

## Location Hierarchy Guidelines

### Root AGENTS.md
**Scope**: Repository-wide standards, architecture, core patterns

**Include**:
- Critical safety protocols
- Architecture overview (high-level only)
- Script/template standards (patterns, not exhaustive)
- Quality requirements
- Cross-references to subdirectories (progressive disclosure)

**Length targets**:
- Ideal: 200-300 lines (~800-1200 tokens)
- With MCP: <200 lines (<800 tokens)
- Maximum: 500 lines (requires refactoring if exceeded)

**Token budget**: <5% of total context (< 1200 tokens ideal)

### Subdirectory AGENTS.md
**Scope**: Component-specific implementation details

**Include**:
- Component purpose and structure
- Configuration syntax (specific patterns only)
- Integration points
- Testing/reload procedures
- Cross-references to parent/root

**Length targets**:
- Ideal: 150-250 lines (~600-1000 tokens)
- Maximum: 400 lines (split into sub-subdirectories if exceeded)

**Pattern**:
```markdown
**Location**: `/path/to/directory`
**Parent**: See `../AGENTS.md` for overview
**Root**: See `/root/AGENTS.md` for core standards

**CRITICAL**: Be concise. Sacrifice grammar for token-efficiency.
```

### Modular Rules (`.opencode/rules/`)
**Scope**: Topic-specific instructions

**When to use**:
- File >500 lines
- Distinct topics that can be separated
- Path-specific rules needed

**Structure**:
```
.opencode/rules/
├── topic1.md
├── topic2.md
└── category/
    ├── subtopic1.md
    └── subtopic2.md
```

## Anti-Patterns to Avoid

### Over-Documentation
```markdown
# BAD: Repeating official docs
Git is a version control system. Commands:
- git add <file>
- git commit -m "message"
[... 50 lines of git tutorial ...]
```
**Fix**: Link to docs, document only your specific patterns

### Listing Every File
```markdown
# BAD: File-by-file descriptions
~/.config/hypr/
├── hyprland.conf (main config file)
├── monitor.conf (monitor configuration)
[... 20 files with obvious names ...]
```
**Fix**: Tree structure without verbose descriptions

### Generic Advice
```markdown
# BAD: Self-evident practices
- Write clean code
- Use meaningful variable names
- Test thoroughly
```
**Fix**: Delete (opencode already knows this)

### Historical Background
```markdown
# BAD: Unnecessary history
Hyprland was created in 2021 by vaxry. It's a dynamic
tiling Wayland compositor written in C++.
```
**Fix**: Delete (opencode knows this, not relevant to your config)

## Maintenance Schedule

### Weekly
- Note when opencode violates/ignores rules
- Track questions opencode asks that are documented

### Monthly
- Review sections for continued relevance
- Prune content that hasn't been referenced
- Update changed workflows

### Quarterly
- Full validation against checklist
- Reorganize if file grew >500 lines
- Validate imports and cross-references
- Check effectiveness indicators

## Effectiveness Indicators

### Positive Signs (Keep)
✅ opencode uses documented commands regularly
✅ Rules prevent recurring mistakes
✅ Conventions differ from defaults (need documentation)
✅ opencode references sections when working

### Negative Signs (Remove/Fix)
❌ opencode asks questions answered in file → Unclear or too long
❌ opencode ignores documented rules → File too long or needs emphasis
❌ Section never referenced in months → Probably not needed
❌ You find yourself re-explaining documented items → Improve clarity

## Reference Materials

**For validation**: See [quality-checklist.md](references/quality-checklist.md)
**For guidelines**: See [anthropic-guidelines.md](references/anthropic-guidelines.md)
**For dotfiles patterns**: See [dotfiles-patterns.md](references/dotfiles-patterns.md)

## Quick Commands

```bash
# Validate all AGENTS.md files
python3 .opencode/skills/agents-md-manager/scripts/validate_agents_md.py .

# Check specific file
python3 .opencode/skills/agents-md-manager/scripts/validate_agents_md.py path/to/AGENTS.md

# Find all AGENTS.md files
find . -name "AGENTS.md" -type f

# Check file lengths
find . -name "AGENTS.md" -exec wc -l {} \; | sort -n

# Check for vague language
grep -n "properly\|correctly\|appropriately" AGENTS.md

# Check for generic advice
grep -ni "clean code\|best practices\|good code" AGENTS.md
```
