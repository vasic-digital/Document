package document

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	assert.Equal(t, "/path/file.md", doc.Path)
	assert.Equal(t, "file", doc.Title)
	assert.Equal(t, "md", doc.Extension)
	assert.Equal(t, FormatUnknown, doc.Format)
	assert.Equal(t, int64(-1), doc.ModTime)
	assert.Equal(t, int64(-1), doc.TouchTime)
}

func TestFilename(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	assert.Equal(t, "file.md", doc.Filename())
}

func TestFilenameNoExtension(t *testing.T) {
	doc := New("/path/file", "file", "")
	assert.Equal(t, "file", doc.Filename())
}

func TestDetectFormatByExtension(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	doc.DetectFormatByExtension()
	assert.Equal(t, FormatMarkdown, doc.Format)
}

func TestDetectFormatByExtensionLatex(t *testing.T) {
	doc := New("/path/paper.tex", "paper", "tex")
	doc.DetectFormatByExtension()
	assert.Equal(t, FormatLaTeX, doc.Format)
}

func TestDetectFormatByContent(t *testing.T) {
	doc := New("/path/file.txt", "file", "txt")
	detected := doc.DetectFormatByContent("# Heading\nSome text")
	assert.True(t, detected)
	assert.Equal(t, FormatMarkdown, doc.Format)
}

func TestDetectFormatByContentLatex(t *testing.T) {
	doc := New("/path/file.txt", "file", "txt")
	detected := doc.DetectFormatByContent("\\documentclass{article}")
	assert.True(t, detected)
	assert.Equal(t, FormatLaTeX, doc.Format)
}

func TestDetectFormatByContentNoMatch(t *testing.T) {
	doc := New("/path/file.txt", "file", "txt")
	detected := doc.DetectFormatByContent("Just plain text")
	assert.False(t, detected)
	assert.Equal(t, FormatUnknown, doc.Format)
}

func TestDetectFormatByContentEmpty(t *testing.T) {
	doc := New("/path/file.txt", "file", "txt")
	detected := doc.DetectFormatByContent("")
	assert.False(t, detected)
}

func TestHasChangedInitial(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	assert.True(t, doc.HasChanged())
}

func TestResetChangeTracking(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	doc.ModTime = 1000
	doc.TouchTime = 2000
	doc.ResetChangeTracking()
	assert.Equal(t, int64(-1), doc.ModTime)
	assert.Equal(t, int64(-1), doc.TouchTime)
}

func TestTouch(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	before := time.Now().UnixMilli()
	doc.Touch()
	after := time.Now().UnixMilli()
	assert.GreaterOrEqual(t, doc.TouchTime, before)
	assert.LessOrEqual(t, doc.TouchTime, after)
}

func TestFileExists(t *testing.T) {
	doc := New("/definitely/nonexistent/file.txt", "file", "txt")
	assert.False(t, doc.FileExists())
}

func TestFileExistsReal(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)
	doc := New(path, "test", "txt")
	assert.True(t, doc.FileExists())
}

func TestGetFileModTime(t *testing.T) {
	doc := New("/nonexistent/file.txt", "file", "txt")
	assert.Equal(t, int64(-1), doc.GetFileModTime())
}

func TestGetFileSize(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	os.WriteFile(path, []byte("hello"), 0644)
	doc := New(path, "test", "txt")
	assert.Equal(t, int64(5), doc.GetFileSize())
}

func TestCreateDocument(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "README.md")
	os.WriteFile(path, []byte("# Test"), 0644)
	doc := CreateDocument(path)
	require.NotNil(t, doc)
	assert.Equal(t, "README", doc.Title)
	assert.Equal(t, "md", doc.Extension)
	assert.Equal(t, FormatMarkdown, doc.Format)
}

func TestCreateDocumentNonexistent(t *testing.T) {
	doc := CreateDocument("/nonexistent/file.txt")
	assert.Nil(t, doc)
}

func TestEqual(t *testing.T) {
	doc1 := New("/path/file.md", "file", "md")
	doc1.Format = FormatMarkdown
	doc2 := New("/path/file.md", "file", "md")
	doc2.Format = FormatMarkdown
	assert.True(t, doc1.Equal(doc2))
}

func TestNotEqual(t *testing.T) {
	doc1 := New("/path/file1.md", "file1", "md")
	doc2 := New("/path/file2.md", "file2", "md")
	assert.False(t, doc1.Equal(doc2))
}

func TestEqualNil(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	assert.False(t, doc.Equal(nil))
}

func TestToJSON(t *testing.T) {
	doc := New("/path/file.md", "file", "md")
	doc.Format = FormatMarkdown
	data, err := doc.ToJSON()
	require.NoError(t, err)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	assert.Equal(t, "/path/file.md", m["path"])
	assert.Equal(t, "markdown", m["format"])
}

func TestFromJSON(t *testing.T) {
	data := []byte(`{"path":"/path/file.md","title":"file","extension":"md","format":"markdown"}`)
	doc, err := FromJSON(data)
	require.NoError(t, err)
	assert.Equal(t, "/path/file.md", doc.Path)
	assert.Equal(t, FormatMarkdown, doc.Format)
	assert.Equal(t, int64(-1), doc.ModTime)
}

func TestFormatConstants(t *testing.T) {
	assert.Equal(t, "unknown", FormatUnknown)
	assert.Equal(t, "plaintext", FormatPlaintext)
	assert.Equal(t, "markdown", FormatMarkdown)
	assert.Equal(t, "todotxt", FormatTodoTxt)
	assert.Equal(t, "csv", FormatCSV)
	assert.Equal(t, "latex", FormatLaTeX)
	assert.Equal(t, "orgmode", FormatOrgMode)
	assert.Equal(t, "wikitext", FormatWikitext)
	assert.Equal(t, "asciidoc", FormatAsciiDoc)
	assert.Equal(t, "restructuredtext", FormatReStructuredText)
	assert.Equal(t, "taskpaper", FormatTaskPaper)
	assert.Equal(t, "textile", FormatTextile)
	assert.Equal(t, "creole", FormatCreole)
	assert.Equal(t, "tiddlywiki", FormatTiddlyWiki)
	assert.Equal(t, "jupyter", FormatJupyter)
	assert.Equal(t, "rmarkdown", FormatRMarkdown)
	assert.Equal(t, "keyvalue", FormatKeyValue)
	assert.Equal(t, "binary", FormatBinary)
}

func TestDetectByExtensionAll(t *testing.T) {
	tests := map[string]string{
		"md": FormatMarkdown, "txt": FormatPlaintext, "csv": FormatCSV,
		"tex": FormatLaTeX, "org": FormatOrgMode, "wiki": FormatWikitext,
		"rst": FormatReStructuredText, "adoc": FormatAsciiDoc,
		"ipynb": FormatJupyter, "rmd": FormatRMarkdown, "ini": FormatKeyValue,
	}
	for ext, expected := range tests {
		assert.Equal(t, expected, DetectByExtension(ext), "extension: %s", ext)
	}
}

func TestDetectByExtensionCaseInsensitive(t *testing.T) {
	assert.Equal(t, FormatMarkdown, DetectByExtension("MD"))
	assert.Equal(t, FormatMarkdown, DetectByExtension("Md"))
}

func TestDetectByExtensionUnknown(t *testing.T) {
	assert.Equal(t, FormatPlaintext, DetectByExtension("xyz"))
}

func TestGetAllExtensions(t *testing.T) {
	exts := GetAllExtensions()
	assert.NotEmpty(t, exts)
	for _, ext := range exts {
		assert.True(t, len(ext) > 1 && ext[0] == '.', "extension should start with dot: %s", ext)
	}
}
