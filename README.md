# digital.vasic.document

A Go document-model library providing format detection (18 formats), change tracking, file-stat helpers, and JSON serialization. Module path: `digital.vasic.document` (Go 1.24). Standalone — no consuming-project context leaks; safe to incorporate at any owning project's root per CONST-051(C).

## Packages

| Package | Purpose |
|---------|---------|
| `pkg/document` | `Document` struct + `New` / `CreateDocument` constructors, `Filename`, change tracking (`HasChanged`/`Touch`/`ResetChangeTracking`), file ops (`FileExists`/`GetFileModTime`/`GetFileSize`), `Equal`, `ToJSON`/`FromJSON`, format detection (`DetectFormatByExtension`/`DetectFormatByContent`, `DetectByExtension`/`DetectByContent`/`GetAllExtensions`), and 18 format constants matching the Document-KMP reference implementation. |

### pkg/document usage

```go
import "digital.vasic.document/pkg/document"

// Create with explicit fields
doc := document.New("/tmp/README.md", "README", "md")
doc.DetectFormatByExtension() // -> "markdown"

// Construct from filesystem
doc = document.CreateDocument("/tmp/README.md")

// Content-based detection (overrides format if matched)
if doc.DetectFormatByContent("\\documentclass{article}") {
    // doc.Format == "latex"
}

// Change tracking
doc.Touch()
if doc.HasChanged() {
    // reload from disk
}

// File stat helpers
exists := doc.FileExists()
size := doc.GetFileSize()
mod := doc.GetFileModTime()

// JSON round-trip
b, _ := doc.ToJSON()
clone, _ := document.FromJSON(b)
```

## Supported formats (extension map)

`markdown` (md/markdown/mdown/mkd/mkdn), `plaintext` (txt/text), `csv` (csv/tsv), `latex` (tex/latex), `orgmode` (org), `wikitext` (wiki/mediawiki), `restructuredtext` (rst/rest), `asciidoc` (adoc/asciidoc/asc), `taskpaper`, `textile`, `creole`, `tiddlywiki` (tid/tiddlywiki), `jupyter` (ipynb), `rmarkdown` (rmd/rmarkdown), `keyvalue` (ini/cfg/conf/properties/toml/yaml/yml/json), `binary` (bin/exe/dll/so/dylib). Unknown extensions fall through to `plaintext`.

Content-pattern detectors (highest-priority first): `\documentclass{` / `\begin{document}` → LaTeX, `(N) ` priority lines & `x YYYY-MM-DD` completion lines → TodoTxt, `#`..`######` ATX headings → Markdown, leading `*`/`**`/`***` → OrgMode, `==..==` headings → Wikitext, 3+3 comma matrix → CSV.

## Build

```bash
go build ./...
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
```

## Anti-bluff guarantees (round-254)

This module ships positive runtime evidence for every claim, not metadata-only PASS. Article XI §11.9 forbids "tests are green / feature is broken" outcomes, so the round-254 enrichment proves end-to-end usability through real OS facilities:

- **Real on-disk transport.** `challenges/runner/main.go` Section 1 + 1b uses `os.MkdirTemp` to create real files, writes them via `os.WriteFile`, then drives `document.CreateDocument`, `Document.GetFileModTime`, `Document.GetFileSize`, `Document.FileExists`, `Document.HasChanged`, and `Document.Touch` against those real files. Bytes round-tripped per locale appear in the PASS line — no "test passed but no proof" outcome possible.
- **Bilingual content detection.** Section 2 runs `DetectByContent` against `tests/fixtures/i18n/payloads.json` payloads for 5 locales (en, sr-Cyrl, ja, ar, zh-CN). LaTeX prelude, ATX-heading Markdown, OrgMode bullet, Wikitext heading, TodoTxt priority line, and CSV grid each prove the regex matrices survive non-ASCII bytes.
- **JSON wire-format preservation.** Section 3 `ToJSON` + `FromJSON` round-trip every locale's non-ASCII Title/Author/Tags through `encoding/json` and asserts byte length matches input — proves the wire stays intact across the serialization boundary.
- **Extension map completeness.** Section 4 asserts `GetAllExtensions()` returns a `.`-prefixed entry for every key in the map and that `DetectByExtension` is case-insensitive (covers `MD`, `Md`, `md`).
- **Equal-by-tuple invariant.** Section 5 plants a Path-only divergence, a Title-only divergence, and a Format-only divergence to prove `Equal` is sensitive to all three components — bluff-proofs the "two near-identical Documents compare equal" failure mode.
- **Paired mutation.** `challenges/document_describe_challenge.sh --anti-bluff-mutate` plants a deliberate symbol-rename mutation in a tmp copy of `docs/test-coverage.md` (e.g. `CreateDocument` → `CreateBogus_MUTATED`) and asserts the cross-reference gate exits 99. Proves the gate catches ledger-vs-source drift rather than rubber-stamping it. Clean tree → exit 0; planted mutation → exit 99; real failure → exit 1. CONST-035 + CONST-050(B) compliant.

> Verbatim 2026-05-19 operator mandate: "all existing tests and Challenges do work in anti-bluff manner - they MUST confirm that all tested codebase really works as expected! We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completition and full usability by end users of the product!"

## Running the round-254 Challenge

```bash
# Clean mode — gate PASS expected (exit 0)
bash challenges/document_describe_challenge.sh

# Paired-mutation mode — gate FAIL expected on the planted symbol-rename (exit 99)
bash challenges/document_describe_challenge.sh --anti-bluff-mutate
```

The runner alone (without the wrapper gate) can also be exercised:

```bash
go run ./challenges/runner -fixtures ./tests/fixtures/i18n/payloads.json
```

## Symbol-to-test ledger

See [`docs/test-coverage.md`](docs/test-coverage.md) for the CONST-050(B) symbol × test cross-reference of every exported `pkg/document` symbol. Drift between the ledger and the source tree is mechanically detected by the round-254 Challenge.

## License

Apache-2.0
