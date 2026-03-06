package document

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Format constants matching Document-KMP.
const (
	FormatUnknown           = "unknown"
	FormatPlaintext         = "plaintext"
	FormatMarkdown          = "markdown"
	FormatTodoTxt           = "todotxt"
	FormatCSV               = "csv"
	FormatWikitext          = "wikitext"
	FormatKeyValue          = "keyvalue"
	FormatAsciiDoc          = "asciidoc"
	FormatOrgMode           = "orgmode"
	FormatLaTeX             = "latex"
	FormatReStructuredText  = "restructuredtext"
	FormatTaskPaper         = "taskpaper"
	FormatTextile           = "textile"
	FormatCreole            = "creole"
	FormatTiddlyWiki        = "tiddlywiki"
	FormatJupyter           = "jupyter"
	FormatRMarkdown         = "rmarkdown"
	FormatBinary            = "binary"
)

// Document represents a text document with metadata.
type Document struct {
	ID        string   `json:"id"`
	Path      string   `json:"path"`
	Title     string   `json:"title"`
	Extension string   `json:"extension"`
	Content   string   `json:"content,omitempty"`
	Author    string   `json:"author,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Format    string   `json:"format"`
	ModTime   int64    `json:"-"`
	TouchTime int64    `json:"-"`
}

// New creates a Document with the given path, title, and extension.
func New(path, title, extension string) *Document {
	return &Document{
		Path:      path,
		Title:     title,
		Extension: extension,
		Format:    FormatUnknown,
		ModTime:   -1,
		TouchTime: -1,
	}
}

// Filename returns the title with extension.
func (d *Document) Filename() string {
	if d.Extension != "" {
		return d.Title + "." + d.Extension
	}
	return d.Title
}

// HasChanged reports whether the document's file has been modified since last tracking.
func (d *Document) HasChanged() bool {
	if d.ModTime < 0 || d.TouchTime < 0 {
		return true
	}
	fileModTime := d.GetFileModTime()
	return fileModTime > d.ModTime
}

// ResetChangeTracking resets modification tracking.
func (d *Document) ResetChangeTracking() {
	d.ModTime = -1
	d.TouchTime = -1
}

// Touch updates the touch time to now.
func (d *Document) Touch() {
	d.TouchTime = time.Now().UnixMilli()
}

// DetectFormatByExtension sets the format based on the file extension.
func (d *Document) DetectFormatByExtension() {
	d.Format = DetectByExtension(d.Extension)
}

// DetectFormatByContent sets the format based on content analysis.
func (d *Document) DetectFormatByContent(content string) bool {
	detected := DetectByContent(content)
	if detected != "" {
		d.Format = detected
		return true
	}
	return false
}

// GetFileModTime returns the file's modification time in milliseconds.
func (d *Document) GetFileModTime() int64 {
	info, err := os.Stat(d.Path)
	if err != nil {
		return -1
	}
	return info.ModTime().UnixMilli()
}

// GetFileSize returns the file size in bytes.
func (d *Document) GetFileSize() int64 {
	info, err := os.Stat(d.Path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// FileExists checks if the document file exists.
func (d *Document) FileExists() bool {
	_, err := os.Stat(d.Path)
	return err == nil
}

// ToJSON serializes the document to JSON.
func (d *Document) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// FromJSON deserializes a document from JSON.
func FromJSON(data []byte) (*Document, error) {
	var doc Document
	err := json.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}
	doc.ModTime = -1
	doc.TouchTime = -1
	return &doc, nil
}

// CreateDocument creates a Document from a file path.
func CreateDocument(path string) *Document {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	_ = info

	base := filepath.Base(path)
	ext := filepath.Ext(base)
	title := strings.TrimSuffix(base, ext)
	ext = strings.TrimPrefix(ext, ".")

	doc := New(path, title, ext)
	doc.DetectFormatByExtension()
	return doc
}

// Equal checks if two documents are equal (by path, title, format).
func (d *Document) Equal(other *Document) bool {
	if other == nil {
		return false
	}
	return d.Path == other.Path && d.Title == other.Title && d.Format == other.Format
}
