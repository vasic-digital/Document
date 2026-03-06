# Document

## Project Overview

Go document model library. Package: `digital.vasic.document/pkg/document`.

## Build Commands

```bash
go build ./...
go vet ./...
go test ./... -count=1
```

## Architecture

- `document.go` - Document struct, New(), Filename(), change tracking, file ops, JSON serialization
- `detector.go` - DetectByExtension(), DetectByContent(), GetAllExtensions()

## Key Patterns

- 18 format constants matching Document-KMP
- Document equality uses Path + Title + Format
- ModTime/TouchTime excluded from JSON via `json:"-"`
