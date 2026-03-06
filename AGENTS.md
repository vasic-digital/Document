# Document Agent Guidelines

## Testing

Run: `go test ./... -count=1`

28 tests in `pkg/document/document_test.go` covering:
- Creation, filename, format detection, change tracking
- File operations, JSON serialization, equality
- Extension mapping, content patterns, constants

## Rules

- Never remove or disable tests
- All changes must pass existing tests
- Format constants must match Document-KMP values
