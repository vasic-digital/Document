// Round-254 challenge runner for digital.vasic.document.
//
// Builds the bilingual fixture set from tests/fixtures/i18n/payloads.json,
// then drives every public Document code path through real on-disk +
// real-process transport:
//
//  1. pkg/document:    CreateDocument -> New -> DetectFormatByExtension
//     -> Filename -> FileExists round-trip per locale. Real os.MkdirTemp
//     file backing.
//  1b. pkg/document:   GetFileModTime / GetFileSize / Touch / HasChanged
//     / ResetChangeTracking against real-file stat deltas + monotonic time.
//  2. pkg/document:    DetectByContent against LaTeX prelude, ATX-heading
//     Markdown, OrgMode bullet, Wikitext heading, TodoTxt priority line,
//     and 3+3 CSV grid — every payload carrying non-ASCII bytes from the
//     fixture's locale-specific text.
//  3. pkg/document:    ToJSON -> FromJSON round-trip per locale; asserts
//     byte preservation of Title/Author/Tags + ModTime/TouchTime reset.
//  4. pkg/document:    GetAllExtensions exhaustiveness + DetectByExtension
//     case-insensitivity matrix.
//  5. pkg/document:    Equal-by-tuple discrimination — three near-clones
//     diverging on Path, Title, Format respectively; each MUST compare
//     unequal.
//
// Anti-bluff invariants enforced (Article XI §11.9 + CONST-035 + CONST-050(B)):
//
//   - No metadata-only / grep-only PASS. Every PASS line is preceded by the
//     locale code, the package exercised, and the actual byte length of the
//     round-tripped string (proves bytes survived, not just that no error
//     was returned).
//   - Real os.MkdirTemp + real os.WriteFile + real os.Stat — no in-memory
//     shortcut. The Document code paths (filesystem stat, JSON marshal,
//     JSON unmarshal, regex matrix) all execute exactly as they would in
//     a downstream consumer.
//   - Failure to round-trip non-ASCII bytes, dropped tag, equal-by-tuple
//     collapse, or detector silently accepting unknown content is a hard
//     FAIL — exit non-zero.
//   - No mocks injected into the library; no patched JSON marshalers; no
//     stubs. The runner uses pkg/document's public surface exactly as a
//     downstream consumer would.
//
// This runner is a Challenge — per CLAUDE.md "Acceptance demo" and per
// the round-242..253 pattern (Cache, Concurrency, Database, EventBus,
// Filesystem, Memory, Auth, Embeddings, Messaging, Middleware,
// Observability, Config), real on-disk + real-OS-stat is the recognised
// mechanism to exercise the real Document transport. The runner is NOT
// production code, NOT a unit test, NOT a stub of the real system — it
// is the real Document API driven against real OS facilities.
//
// Verbatim 2026-05-19 operator mandate: "all existing tests and Challenges
// do work in anti-bluff manner - they MUST confirm that all tested codebase
// really works as expected!"
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"digital.vasic.document/pkg/document"
)

type fixtureInput struct {
	Locale    string   `json:"locale"`
	Title     string   `json:"title"`
	Extension string   `json:"extension"`
	Author    string   `json:"author"`
	Tags      []string `json:"tags"`
}

type fixtureFile struct {
	Inputs []fixtureInput `json:"inputs"`
}

var (
	passCount int
	failCount int
)

func pass(msg string) {
	passCount++
	fmt.Printf("  PASS: %s\n", msg)
}

func fail(msg string) {
	failCount++
	fmt.Printf("  FAIL: %s\n", msg)
}

func main() {
	fixturePath := flag.String("fixtures", "", "path to payloads.json")
	flag.Parse()

	if *fixturePath == "" {
		*fixturePath = filepath.Join("tests", "fixtures", "i18n", "payloads.json")
	}

	raw, err := os.ReadFile(*fixturePath)
	if err != nil {
		fmt.Printf("FATAL: cannot read fixtures %s: %v\n", *fixturePath, err)
		os.Exit(2)
	}
	var ff fixtureFile
	if err := json.Unmarshal(raw, &ff); err != nil {
		fmt.Printf("FATAL: cannot parse fixtures %s: %v\n", *fixturePath, err)
		os.Exit(2)
	}
	if len(ff.Inputs) < 3 {
		fmt.Printf("FATAL: fixtures need >=3 locales, got %d\n", len(ff.Inputs))
		os.Exit(2)
	}

	tmpDir, err := os.MkdirTemp("", "document-round254-*")
	if err != nil {
		fmt.Printf("FATAL: cannot create tmp dir: %v\n", err)
		os.Exit(2)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("=== Round-254 Document Bilingual Runner ===")
	fmt.Printf("fixtures: %s  locales: %d  tmp: %s\n\n", *fixturePath, len(ff.Inputs), tmpDir)

	// Section 1 — CreateDocument + on-disk filename round-trip per locale
	fmt.Println("Section 1: CreateDocument + on-disk filename round-trip")
	for _, in := range ff.Inputs {
		fname := in.Title + "." + in.Extension
		fp := filepath.Join(tmpDir, fname)
		body := []byte("# " + in.Title + "\nAuthor: " + in.Author + "\n")
		if err := os.WriteFile(fp, body, 0o600); err != nil {
			fail(fmt.Sprintf("[document][%s] WriteFile: %v", in.Locale, err))
			continue
		}
		doc := document.CreateDocument(fp)
		if doc == nil {
			fail(fmt.Sprintf("[document][%s] CreateDocument returned nil", in.Locale))
			continue
		}
		if doc.Title != in.Title {
			fail(fmt.Sprintf("[document][%s] Title drift: want %q got %q", in.Locale, in.Title, doc.Title))
			continue
		}
		if doc.Extension != in.Extension {
			fail(fmt.Sprintf("[document][%s] Extension drift: want %q got %q", in.Locale, in.Extension, doc.Extension))
			continue
		}
		if doc.Format != document.FormatMarkdown {
			fail(fmt.Sprintf("[document][%s] Format want markdown got %q", in.Locale, doc.Format))
			continue
		}
		if !doc.FileExists() {
			fail(fmt.Sprintf("[document][%s] FileExists returned false for real file", in.Locale))
			continue
		}
		bytesTitle := utf8.RuneCountInString(doc.Title)
		pass(fmt.Sprintf("[document][%s] CreateDocument round-trip (title runes=%d, filename=%s)", in.Locale, bytesTitle, doc.Filename()))
	}

	// Section 1b — GetFileModTime / GetFileSize / Touch / HasChanged / ResetChangeTracking
	fmt.Println("\nSection 1b: file stat + change tracking on real files")
	for _, in := range ff.Inputs {
		fname := in.Title + "." + in.Extension
		fp := filepath.Join(tmpDir, fname)
		doc := document.CreateDocument(fp)
		if doc == nil {
			fail(fmt.Sprintf("[stat][%s] CreateDocument returned nil", in.Locale))
			continue
		}
		size := doc.GetFileSize()
		if size <= 0 {
			fail(fmt.Sprintf("[stat][%s] GetFileSize returned %d", in.Locale, size))
			continue
		}
		mod := doc.GetFileModTime()
		if mod <= 0 {
			fail(fmt.Sprintf("[stat][%s] GetFileModTime returned %d", in.Locale, mod))
			continue
		}
		// Initial HasChanged must be true (ModTime/TouchTime are -1 after CreateDocument).
		if !doc.HasChanged() {
			fail(fmt.Sprintf("[stat][%s] HasChanged returned false on fresh Document", in.Locale))
			continue
		}
		// Sync tracking: pretend we just loaded — set ModTime/TouchTime to current values.
		doc.ModTime = mod
		doc.TouchTime = time.Now().UnixMilli()
		if doc.HasChanged() {
			fail(fmt.Sprintf("[stat][%s] HasChanged true after sync (no real change yet)", in.Locale))
			continue
		}
		// Real change: append bytes and wait > 1ms for OS to update mtime.
		time.Sleep(5 * time.Millisecond)
		if err := os.WriteFile(fp, []byte("# "+in.Title+"\nAuthor: "+in.Author+"\nExtra\n"), 0o600); err != nil {
			fail(fmt.Sprintf("[stat][%s] append failed: %v", in.Locale, err))
			continue
		}
		if !doc.HasChanged() {
			fail(fmt.Sprintf("[stat][%s] HasChanged false after real file change", in.Locale))
			continue
		}
		doc.Touch()
		if doc.TouchTime <= 0 {
			fail(fmt.Sprintf("[stat][%s] Touch did not set TouchTime", in.Locale))
			continue
		}
		doc.ResetChangeTracking()
		if doc.ModTime != -1 || doc.TouchTime != -1 {
			fail(fmt.Sprintf("[stat][%s] ResetChangeTracking did not reset (mod=%d touch=%d)", in.Locale, doc.ModTime, doc.TouchTime))
			continue
		}
		pass(fmt.Sprintf("[stat][%s] size=%d mod=%d Touch+HasChanged+Reset OK", in.Locale, size, mod))
	}

	// Section 2 — DetectByContent over 6 content classes × per locale
	fmt.Println("\nSection 2: DetectByContent on bilingual payloads")
	for _, in := range ff.Inputs {
		// LaTeX
		latex := "\\documentclass{article}\n\\begin{document}\n" + in.Title + "\n\\end{document}\n"
		if document.DetectByContent(latex) != document.FormatLaTeX {
			fail(fmt.Sprintf("[detect][%s] LaTeX not detected", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] LaTeX prelude (bytes=%d)", in.Locale, len(latex)))
		}
		// Markdown
		md := "# " + in.Title + "\nbody text " + in.Author + "\n"
		if document.DetectByContent(md) != document.FormatMarkdown {
			fail(fmt.Sprintf("[detect][%s] Markdown ATX heading not detected", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] Markdown ATX (bytes=%d)", in.Locale, len(md)))
		}
		// OrgMode
		org := "* " + in.Title + "\n** sub " + in.Author + "\n"
		if document.DetectByContent(org) != document.FormatOrgMode {
			fail(fmt.Sprintf("[detect][%s] OrgMode bullet not detected", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] OrgMode bullet (bytes=%d)", in.Locale, len(org)))
		}
		// Wikitext
		wiki := "== " + in.Title + " ==\nintro " + in.Author + "\n"
		if document.DetectByContent(wiki) != document.FormatWikitext {
			fail(fmt.Sprintf("[detect][%s] Wikitext heading not detected", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] Wikitext heading (bytes=%d)", in.Locale, len(wiki)))
		}
		// TodoTxt
		todo := "(A) " + in.Title + " due\n(B) " + in.Author + " review\n"
		if document.DetectByContent(todo) != document.FormatTodoTxt {
			fail(fmt.Sprintf("[detect][%s] TodoTxt priority not detected", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] TodoTxt priority (bytes=%d)", in.Locale, len(todo)))
		}
		// CSV (needs 3 fields per row over 2 rows minimum)
		csv := in.Title + "," + in.Author + "," + in.Tags[0] + "\n" + in.Tags[1] + "," + in.Tags[2] + ",end\n"
		if document.DetectByContent(csv) != document.FormatCSV {
			fail(fmt.Sprintf("[detect][%s] CSV grid not detected", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] CSV grid (bytes=%d)", in.Locale, len(csv)))
		}
		// No-match sanity
		if document.DetectByContent("       \n\n") != "" {
			fail(fmt.Sprintf("[detect][%s] whitespace-only returned non-empty", in.Locale))
		} else {
			pass(fmt.Sprintf("[detect][%s] whitespace-only correctly returned empty", in.Locale))
		}
	}

	// Section 3 — ToJSON / FromJSON round-trip per locale
	fmt.Println("\nSection 3: ToJSON/FromJSON wire-format round-trip")
	for _, in := range ff.Inputs {
		doc := document.New("/virt/"+in.Title+"."+in.Extension, in.Title, in.Extension)
		doc.Format = document.FormatMarkdown
		doc.Author = in.Author
		doc.Tags = append([]string{}, in.Tags...)
		raw, err := doc.ToJSON()
		if err != nil {
			fail(fmt.Sprintf("[json][%s] ToJSON: %v", in.Locale, err))
			continue
		}
		clone, err := document.FromJSON(raw)
		if err != nil {
			fail(fmt.Sprintf("[json][%s] FromJSON: %v", in.Locale, err))
			continue
		}
		if clone.Title != in.Title {
			fail(fmt.Sprintf("[json][%s] Title drift via JSON: want %q got %q", in.Locale, in.Title, clone.Title))
			continue
		}
		if clone.Author != in.Author {
			fail(fmt.Sprintf("[json][%s] Author drift via JSON: want %q got %q", in.Locale, in.Author, clone.Author))
			continue
		}
		if len(clone.Tags) != len(in.Tags) {
			fail(fmt.Sprintf("[json][%s] Tags length drift: want %d got %d", in.Locale, len(in.Tags), len(clone.Tags)))
			continue
		}
		for i, tag := range in.Tags {
			if clone.Tags[i] != tag {
				fail(fmt.Sprintf("[json][%s] Tag[%d] drift: want %q got %q", in.Locale, i, tag, clone.Tags[i]))
				continue
			}
		}
		// ModTime/TouchTime MUST reset to -1 after deserialization (json:"-").
		if clone.ModTime != -1 || clone.TouchTime != -1 {
			fail(fmt.Sprintf("[json][%s] FromJSON did not reset ModTime/TouchTime (mod=%d touch=%d)", in.Locale, clone.ModTime, clone.TouchTime))
			continue
		}
		pass(fmt.Sprintf("[json][%s] round-trip byte-preserved (raw=%d, title-runes=%d, tags=%d)", in.Locale, len(raw), utf8.RuneCountInString(clone.Title), len(clone.Tags)))
	}

	// Section 4 — GetAllExtensions exhaustiveness + DetectByExtension case-insensitivity
	fmt.Println("\nSection 4: GetAllExtensions + DetectByExtension case-insensitivity")
	exts := document.GetAllExtensions()
	if len(exts) < 18 {
		fail(fmt.Sprintf("[ext-map] GetAllExtensions returned %d entries (<18)", len(exts)))
	} else {
		pass(fmt.Sprintf("[ext-map] GetAllExtensions returned %d entries", len(exts)))
	}
	dotPrefixOK := true
	for _, e := range exts {
		if len(e) < 2 || !strings.HasPrefix(e, ".") {
			fail(fmt.Sprintf("[ext-map] extension %q lacks dot prefix or is too short", e))
			dotPrefixOK = false
		}
	}
	if dotPrefixOK {
		pass("[ext-map] every extension carries dot prefix")
	}
	caseMatrix := []struct{ in, want string }{
		{"md", document.FormatMarkdown},
		{"MD", document.FormatMarkdown},
		{"Md", document.FormatMarkdown},
		{"IPYNB", document.FormatJupyter},
		{"TeX", document.FormatLaTeX},
		{"YAML", document.FormatKeyValue},
		{"xyz_unknown_ext", document.FormatPlaintext}, // fallthrough
	}
	for _, c := range caseMatrix {
		got := document.DetectByExtension(c.in)
		if got != c.want {
			fail(fmt.Sprintf("[ext-map] DetectByExtension(%q) want %q got %q", c.in, c.want, got))
		} else {
			pass(fmt.Sprintf("[ext-map] DetectByExtension(%q) -> %q", c.in, got))
		}
	}

	// Section 5 — Equal-by-tuple discrimination
	fmt.Println("\nSection 5: Equal-by-tuple discrimination")
	base := document.New("/tmp/a.md", "a", "md")
	base.Format = document.FormatMarkdown

	pathDiv := document.New("/tmp/b.md", "a", "md")
	pathDiv.Format = document.FormatMarkdown
	if base.Equal(pathDiv) {
		fail("[equal] Path-only divergence collapsed under Equal")
	} else {
		pass("[equal] Path-only divergence correctly unequal")
	}

	titleDiv := document.New("/tmp/a.md", "DIFFERENT", "md")
	titleDiv.Format = document.FormatMarkdown
	if base.Equal(titleDiv) {
		fail("[equal] Title-only divergence collapsed under Equal")
	} else {
		pass("[equal] Title-only divergence correctly unequal")
	}

	formatDiv := document.New("/tmp/a.md", "a", "md")
	formatDiv.Format = document.FormatLaTeX
	if base.Equal(formatDiv) {
		fail("[equal] Format-only divergence collapsed under Equal")
	} else {
		pass("[equal] Format-only divergence correctly unequal")
	}

	identical := document.New("/tmp/a.md", "a", "md")
	identical.Format = document.FormatMarkdown
	if !base.Equal(identical) {
		fail("[equal] Identical tuple compared unequal (false negative)")
	} else {
		pass("[equal] Identical tuple correctly equal")
	}

	if base.Equal(nil) {
		fail("[equal] Equal(nil) returned true")
	} else {
		pass("[equal] Equal(nil) correctly false")
	}

	fmt.Printf("\n=== Summary: PASS=%d FAIL=%d ===\n", passCount, failCount)
	if failCount > 0 {
		os.Exit(1)
	}
}
