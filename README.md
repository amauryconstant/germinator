# Germinator

A configuration adapter that transforms AI coding assistant documents (commands, memory, skills, agents) between platforms. It uses Claude Code's document standard as the source format and adapts it for other platforms.

## Overview

Germinator enables users who test different AI coding assistants regularly to:
- Maintain **one source of truth** for their coding assistant setup
- Quickly **switch platforms** without rewriting their configuration
- **Adapt** their setup to new projects easily

## Directory Structure

```
germinator/
├── cmd/                    # CLI entry point
│   └── root.go            # Main entry point using Cobra
├── internal/              # Private application code
│   ├── core/             # Core interfaces and implementations
│   └── services/         # Business logic services
├── pkg/                   # Public library code
│   └── models/           # Domain models (Document, Agent, Command, etc.)
├── config/               # Configuration files
│   ├── schemas/          # JSON Schema files for validation
│   ├── templates/        # Template files for rendering output
│   └── adapters/         # Platform adapter configurations
├── test/                 # Test artifacts
│   ├── fixtures/         # Test input documents
│   └── golden/           # Expected output files for comparison
└── scripts/              # Utility scripts
```

## Build

### Prerequisites

- Go 1.25.5 or later

### Build Commands

```bash
# Build the CLI
go build -o germinator ./cmd

# Build all packages
go build ./...

# Run tests
go test ./...

# Verify with go vet
go vet ./...
```

## Usage

```bash
./germinator --help
```

## Key Constraints

- **No predefined directory structure** - works with any input/output paths
- **Platform differences handled** - tool names, permissions, conventions mapped appropriately
- **Source content preserved** - only adapted/enriched for target platform
- **If platform doesn't support a feature** → it's not supported (no forced compatibility)

## License

[Add your license here]
