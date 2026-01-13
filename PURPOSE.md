# Purpose

## What This Tool Is

A **configuration adapter** that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. It uses Claude Code's document standard as the source format and adapts it for other platforms.

## Use Case

Users who test different AI coding assistants regularly. The tool enables them to:

1. Maintain **one source of truth** for their coding assistant setup
2. Quickly **switch platforms** without rewriting their configuration
3. **Adapt** their setup to new projects easily

## How Users Plan to Use It

```bash
cli action input_file target_platform [options]

```

## Key Constraints

- **No predefined directory structure** - works with any input/output paths
- **Platform differences handled** - tool names, permissions, conventions mapped appropriately
- **Source content preserved** - only adapted/enriched for target platform
- **If platform doesn't support a feature** → it's not supported (no forced compatibility)

## Assessment

**What the tool provides**: A platform adapter that solves the **configuration lock-in problem** for AI coding assistants. It's essentially a **config converter** that enables portable coding assistant setups.

**Why it matters**: As the AI coding assistant landscape matures, developers need to switch tools without losing their customized configurations. This tool provides that portability.
