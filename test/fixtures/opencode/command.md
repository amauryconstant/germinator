---
name: git-release
description: Create consistent releases and changelogs
allowed-tools:
  - bash
context: fork
agent: general-purpose
argument-hint: [version-number]
model: claude-sonnet-4-5-20250929
---
## What I do
- Draft release notes from merged PRs
- Propose a version bump
- Provide a copy-pasteable `gh release create` command

## When to use me
Use this when you are preparing a tagged release.
Ask clarifying questions if target versioning scheme is unclear.
