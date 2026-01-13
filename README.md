# Germinator

A configuration adapter that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. It uses Claude Code's document standard as the source format and adapts it for other platforms.

## Overview

Germinator enables users who test different AI coding assistants regularly to:
- Maintain **one source of truth** for their coding assistant setup
- Quickly **switch platforms** without rewriting their configuration
- **Adapt** their setup to new projects easily

## Installation

### From Source

```bash
go build -o germinator ./cmd
```

### Prerequisites

- Go 1.25.5 or later

## Usage

```bash
./germinator --help
```

### Commands

Germinator provides two main commands:

- **validate** - Validate a document file
- **adapt** - Transform a document to another platform

### Examples

```bash
# Validate a document
./germinator validate path/to/document.md

# Adapt a document for Claude Code
./germinator adapt path/to/document.md output.md --platform claude-code
```

## Document Types

Germinator supports four types of AI coding assistant documents:

- **Agents** - Specialized AI agents with capabilities and selection criteria
- **Commands** - CLI commands with tool references and templates
- **Memory** - Context and guidelines for AI assistants
- **Skills** - Specialized skills and techniques

## License

[Add your license here]
