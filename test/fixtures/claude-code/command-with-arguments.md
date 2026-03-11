---
name: create-component
description: Creates a new component with tests
tools:
  - Bash
  - Read
  - Write
execution:
  context: fork
  agent: general-purpose
arguments:
  hint: "[component-name] [--with-styles] [--with-story]"
model: claude-sonnet-4-5-20250929
---
Component creation workflow for $ARGUMENTS:

## Usage
```
/create-component Button --with-styles --with-story
```

## Steps
1. Parse component name and flags from $ARGUMENTS
2. Create component file in src/components/[Name]/[Name].tsx
3. If --with-styles: Create [Name].module.css
4. If --with-story: Create [Name].stories.tsx
5. Create test file [Name].test.tsx
6. Update index exports
