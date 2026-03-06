# Document

Go document model with format detection and change tracking.

## Features

- **Document** - Struct with path, title, extension, format, content, author, tags
- **FormatDetector** - Extension and content-based format detection for 18 formats
- **Change Tracking** - HasChanged(), Touch(), ResetChangeTracking()
- **File Operations** - FileExists(), GetFileModTime(), GetFileSize()
- **JSON Serialization** - ToJSON(), FromJSON()

## Usage

```go
import "digital.vasic.document/pkg/document"

// Create a document
doc := document.New("/path/to/README.md", "README", "md")
doc.DetectFormatByExtension() // format = "markdown"

// Content detection
doc.DetectFormatByContent("\\documentclass{article}") // detects LaTeX

// Change tracking
doc.Touch()
if doc.HasChanged() { /* reload */ }

// Create from file path
doc = document.CreateDocument("/path/to/file.md")
```

## License

Apache-2.0
