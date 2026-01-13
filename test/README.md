# Test Support

This directory stores test fixtures and golden files for testing.

## Structure

- `fixtures/` - Test fixtures and test data
- `golden/` - Golden files (expected outputs) for regression tests

## When to Add Files

Add files to these directories when writing tests:
- Place test data in `fixtures/`
- Store expected outputs in `golden/` for comparison with actual results

## Notes

- Subdirectory READMEs will be added when content is added
- Golden files are used for snapshot testing and regression validation
