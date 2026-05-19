# Test Coverage — digital.vasic.document (round-254)

> Verbatim 2026-05-19 operator mandate: *"all existing tests and Challenges do work in anti-bluff manner - they MUST confirm that all tested codebase really works as expected! We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completition and full usability by end users of the product!"*

CONST-050(B) symbol-to-test ledger. Every exported symbol in `pkg/document` is cross-referenced to the test name(s) that exercise it AND to the round-254 Challenge runner section that exercises it against real OS facilities (`os.MkdirTemp` files + real `os.Stat` / `os.WriteFile`). No metadata-only PASS — every entry below names the production code path and the runtime evidence channel that proves it works.

## Anti-bluff posture (round-254)

- **Real on-disk transport.** `challenges/runner/main.go` Section 1 + 1b creates files in `os.MkdirTemp` and round-trips bilingual (en/sr-Cyrl/ja/ar/zh-CN) titles through `CreateDocument`, `FileExists`, `GetFileModTime`, `GetFileSize`, `HasChanged`, and `Touch`. Byte-length of the round-tripped Title is captured per locale in the PASS line.
- **Real content-pattern detection.** Section 2 drives `DetectByContent` against LaTeX, Markdown, OrgMode, Wikitext, TodoTxt, and CSV payloads carrying non-ASCII text — proving regex-driven detection survives multi-byte UTF-8.
- **Real JSON wire-format.** Section 3 `ToJSON`/`FromJSON` per locale; non-ASCII Title/Author/Tags byte-preserved; `ModTime`/`TouchTime` correctly reset to `-1` after deserialization (since they carry `json:"-"`).
- **Extension-map exhaustiveness.** Section 4 asserts `GetAllExtensions` returns `.`-prefixed entries for every extension key AND that `DetectByExtension` is case-insensitive across `MD`/`Md`/`md`.
- **Equal-by-tuple discrimination.** Section 5 builds three near-clones diverging on Path, Title, and Format respectively; each MUST compare unequal — bluff-proofs the "Equal silently merges divergent documents" failure mode.
- **Paired mutation.** `document_describe_challenge.sh --anti-bluff-mutate` plants a `CreateDocument -> CreateBogus_MUTATED` rename in a tmp ledger copy and asserts the cross-reference gate exits 99. Proves the gate catches ledger-vs-source drift instead of rubber-stamping it.

## pkg/document — types

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `type Document` | `TestNew`, `TestFilename`, `TestFilenameNoExtension`, `TestToJSON`, `TestFromJSON`, `TestEqual`, `TestNotEqual`, `TestEqualNil` | Section 1 / 1b / 3 / 5 |

## pkg/document — constructors

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `func New(path, title, extension string) *Document` | `TestNew`, `TestFilename`, `TestFilenameNoExtension`, `TestDetectFormatByExtension`, `TestDetectFormatByExtensionLatex`, `TestDetectFormatByContent`, `TestDetectFormatByContentLatex`, `TestDetectFormatByContentNoMatch`, `TestDetectFormatByContentEmpty`, `TestHasChangedInitial`, `TestResetChangeTracking`, `TestTouch`, `TestFileExists`, `TestFileExistsReal`, `TestGetFileModTime`, `TestGetFileSize`, `TestEqual`, `TestNotEqual`, `TestEqualNil`, `TestToJSON` | Section 1 (constructor exercised per locale), Section 5 (divergence cases) |
| `func CreateDocument(path string) *Document` | `TestCreateDocument`, `TestCreateDocumentNonexistent` | Section 1 (real on-disk filename → Document round-trip per locale) |
| `func FromJSON(data []byte) (*Document, error)` | `TestFromJSON` | Section 3 (per-locale wire-format round-trip) |

## pkg/document — methods on `*Document`

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `func (d *Document) Filename() string` | `TestFilename`, `TestFilenameNoExtension` | Section 1 |
| `func (d *Document) HasChanged() bool` | `TestHasChangedInitial` | Section 1b (real-file mod-time delta) |
| `func (d *Document) ResetChangeTracking()` | `TestResetChangeTracking` | Section 1b |
| `func (d *Document) Touch()` | `TestTouch` | Section 1b (monotonic time delta) |
| `func (d *Document) DetectFormatByExtension()` | `TestDetectFormatByExtension`, `TestDetectFormatByExtensionLatex` | Section 1 (per-locale Document) |
| `func (d *Document) DetectFormatByContent(content string) bool` | `TestDetectFormatByContent`, `TestDetectFormatByContentLatex`, `TestDetectFormatByContentNoMatch`, `TestDetectFormatByContentEmpty` | Section 2 (per-locale per-format detection) |
| `func (d *Document) GetFileModTime() int64` | `TestGetFileModTime`, `TestGetFileSize` (indirectly via `HasChanged`) | Section 1b (real `os.Stat`) |
| `func (d *Document) GetFileSize() int64` | `TestGetFileSize` | Section 1b (real `os.Stat`) |
| `func (d *Document) FileExists() bool` | `TestFileExists`, `TestFileExistsReal` | Section 1 / 1b |
| `func (d *Document) ToJSON() ([]byte, error)` | `TestToJSON` | Section 3 |
| `func (d *Document) Equal(other *Document) bool` | `TestEqual`, `TestNotEqual`, `TestEqualNil` | Section 5 (three divergence cases) |

## pkg/document — package functions

| Exported symbol | Unit-test coverage | Runner section |
|-----------------|--------------------|----------------|
| `func DetectByExtension(ext string) string` | `TestDetectByExtensionAll`, `TestDetectByExtensionCaseInsensitive`, `TestDetectByExtensionUnknown` | Section 4 (case-insensitive matrix) |
| `func DetectByContent(content string) string` | `TestDetectFormatByContent`, `TestDetectFormatByContentLatex`, `TestDetectFormatByContentNoMatch`, `TestDetectFormatByContentEmpty` | Section 2 (6 content classes × 5 locales) |
| `func GetAllExtensions() []string` | `TestGetAllExtensions` | Section 4 (exhaustiveness + dot-prefix invariant) |

## pkg/document — format constants

All 18 covered by `TestFormatConstants` (unit) AND referenced by Section 1 / 2 / 4 of the runner indirectly through extension-map + content-pattern resolution.

| Constant | String value |
|----------|--------------|
| `FormatUnknown` | `unknown` |
| `FormatPlaintext` | `plaintext` |
| `FormatMarkdown` | `markdown` |
| `FormatTodoTxt` | `todotxt` |
| `FormatCSV` | `csv` |
| `FormatWikitext` | `wikitext` |
| `FormatKeyValue` | `keyvalue` |
| `FormatAsciiDoc` | `asciidoc` |
| `FormatOrgMode` | `orgmode` |
| `FormatLaTeX` | `latex` |
| `FormatReStructuredText` | `restructuredtext` |
| `FormatTaskPaper` | `taskpaper` |
| `FormatTextile` | `textile` |
| `FormatCreole` | `creole` |
| `FormatTiddlyWiki` | `tiddlywiki` |
| `FormatJupyter` | `jupyter` |
| `FormatRMarkdown` | `rmarkdown` |
| `FormatBinary` | `binary` |

## Runner section index

| Section | Production code path exercised | Real OS facility |
|---------|--------------------------------|------------------|
| 1 — Filesystem constructor | `CreateDocument` → `New` → `DetectFormatByExtension` → `Filename` → `FileExists` | `os.MkdirTemp` + `os.WriteFile` + `os.Stat` |
| 1b — Mod-time + Touch tracking | `GetFileModTime` → `GetFileSize` → `Touch` → `HasChanged` → `ResetChangeTracking` | real-file stat delta + monotonic time |
| 2 — Bilingual content detection | `DetectFormatByContent` / `DetectByContent` | regex matrix vs UTF-8 input |
| 3 — JSON wire-format round-trip | `ToJSON` → `FromJSON` per locale | `encoding/json` real marshal/unmarshal |
| 4 — Extension map exhaustiveness | `GetAllExtensions` → `DetectByExtension` (case-insensitive) | dot-prefix audit + case-folding |
| 5 — Equal-by-tuple divergence | `Equal` against Path/Title/Format-divergent clones | n/a (pure-function discrimination) |

## Coverage rollup

- Total exported symbols in `pkg/document`: 16 (1 type + 3 constructors + 9 methods + 3 package functions; plus 18 format constants).
- Unit-test coverage: 100% of exported symbols.
- Runner coverage: 100% of exported symbols (constants exercised transitively via extension-map and content-pattern resolution; explicitly enumerated in Section 4 and Section 2 respectively).
- Paired-mutation gate: planted symbol-rename forces gate to exit 99 — proves ledger-vs-source drift detection works.

## Cross-references

- `CONST-035` Article XI §11.9 — anti-bluff forensic anchor.
- `CONST-050(B)` — 100% test-type coverage.
- `CONST-046` — no hardcoded English-only strings; bilingual fixture covers en/sr-Cyrl/ja/ar/zh-CN.
- `CONST-053` — `.gitignore` hardened; no build artefacts versioned.
