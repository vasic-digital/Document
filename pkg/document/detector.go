package document

import (
	"regexp"
	"strings"
)

var extensionMap = map[string]string{
	"md": FormatMarkdown, "markdown": FormatMarkdown, "mdown": FormatMarkdown,
	"mkd": FormatMarkdown, "mkdn": FormatMarkdown,
	"txt": FormatPlaintext, "text": FormatPlaintext,
	"csv": FormatCSV, "tsv": FormatCSV,
	"tex": FormatLaTeX, "latex": FormatLaTeX,
	"org": FormatOrgMode,
	"wiki": FormatWikitext, "mediawiki": FormatWikitext,
	"rst": FormatReStructuredText, "rest": FormatReStructuredText,
	"adoc": FormatAsciiDoc, "asciidoc": FormatAsciiDoc, "asc": FormatAsciiDoc,
	"taskpaper": FormatTaskPaper,
	"textile": FormatTextile,
	"creole": FormatCreole,
	"tid": FormatTiddlyWiki, "tiddlywiki": FormatTiddlyWiki,
	"ipynb": FormatJupyter,
	"rmd": FormatRMarkdown, "rmarkdown": FormatRMarkdown,
	"ini": FormatKeyValue, "cfg": FormatKeyValue, "conf": FormatKeyValue,
	"properties": FormatKeyValue, "toml": FormatKeyValue,
	"yaml": FormatKeyValue, "yml": FormatKeyValue, "json": FormatKeyValue,
	"bin": FormatBinary, "exe": FormatBinary, "dll": FormatBinary,
	"so": FormatBinary, "dylib": FormatBinary,
}

type contentPattern struct {
	re     *regexp.Regexp
	format string
}

var contentPatterns = []contentPattern{
	{regexp.MustCompile(`\\documentclass\{`), FormatLaTeX},
	{regexp.MustCompile(`\\begin\{document\}`), FormatLaTeX},
	{regexp.MustCompile(`(?m)^\(\w\)\s+`), FormatTodoTxt},
	{regexp.MustCompile(`(?m)^x\s+\d{4}-\d{2}-\d{2}\s+`), FormatTodoTxt},
	{regexp.MustCompile(`(?m)^#{1,6}\s+.+`), FormatMarkdown},
	{regexp.MustCompile(`(?m)^\*{1,3}\s+.+`), FormatOrgMode},
	{regexp.MustCompile(`(?m)^=={1,5}\s+.+\s+=={1,5}`), FormatWikitext},
	{regexp.MustCompile(`(?m)^[^,\n]+,[^,\n]+,[^,\n]+\n[^,\n]+,[^,\n]+,[^,\n]+`), FormatCSV},
}

// DetectByExtension returns the format for a given file extension.
func DetectByExtension(ext string) string {
	if f, ok := extensionMap[strings.ToLower(ext)]; ok {
		return f
	}
	return FormatPlaintext
}

// DetectByContent returns the format detected from content, or empty string if none.
func DetectByContent(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	for _, p := range contentPatterns {
		if p.re.MatchString(content) {
			return p.format
		}
	}
	return ""
}

// GetAllExtensions returns all registered extensions with dot prefix.
func GetAllExtensions() []string {
	exts := make([]string, 0, len(extensionMap))
	for k := range extensionMap {
		exts = append(exts, "."+k)
	}
	return exts
}
