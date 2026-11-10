## Brief Review (Round 1)

### BLOCK findings

**[B-1]** §3 says "**No guard is needed**" for headless `SetChecked` in tests, but §8's verified-source note on `SetChecked` says the opposite: "whether the post is safe under `go test` after `menu()` ran is not verified here, **hence the guard in §3**." — Two parts of the brief directly disagree: §3 instructs tests to build `a.menu()`, keep the items, and call `SetChecked` with no guard, while the evidence note that §3 cites concludes a guard is needed and even claims §3 contains one. Criterion 12 makes these tests mandatory, so the disagreement lands in required test code. — Resolve whether `postMain` after `menu()` is safe under `go test`, state one answer, and make §3 and §8 say the same thing.

**[B-2]** Criterion 11 and §4 "App lifecycle" are contradicted by the shown code. The brief asserts "the reopen depends on `a.win == nil` alone," but `main.go:46-50` (shown) is `if !hasVisibleWindows && a.win == nil { a.open() }` — the reopen today depends on `!hasVisibleWindows` too. With Settings open and the main window closed, `hasVisibleWindows` is true, so today's `OnActivate` would NOT reopen the main window. — Satisfying criterion 11 therefore requires changing `main.go:46-50` to drop the `hasVisibleWindows` condition (or otherwise special-case reopen), but §5 "Reuse & extend-vs-new" lists no such change to `OnActivate`. The brief must state the `OnActivate` change explicitly and confirm it does not regress the "no window at all → reopen on Dock click" and "app quit when all windows closed" paths.

### Needs-verification findings

**[NV-1]** That `ui.KeyComma` exists in mygo's `ui` package and that adding it to `shortcut.String()`'s key-name map produces "," — the ui package's key constants are not excerpted; a `ui` keys source excerpt (or the `plugins/terminal/keys.go` mapping cited in Source material) would settle it.

**[NV-2]** That Ctrl+, in legacy terminal encoding sends nothing on Windows (the claimed basis for the README exception) — evidence would be the same source the existing Ctrl+=/Ctrl− exception cites (README's existing note or the encoding tables it references).

**[NV-3]** The §8 mygo facts: `menu.go:243-264` `SetChecked`/`update` posts without waiting; `app.go:591` macOS menu click window is `FocusedWindow()`; no `performKeyEquivalent` override in `internal/darwin` outside `testhooks.go:58`; `menu.go:325-327` role submenu only filled when unset; `menu.go:507-508` `RoleClose` closes the clicked window; `window.go` Focus/Close/`NewWindow` blocking headless; `loop.go:87-100` `onMain` waits. None of these lines' text is in this prompt (only the repo excerpts are), so each remains a hypothesis at this gate — showing the corresponding mygo source lines would settle them.

### WARN findings

**[W-1]** Ctrl+, with Settings focused on Windows. Criterion 2 requires Ctrl+, on Windows to "open Settings, or bring the open window to the front." On Windows the main window's `shortcuts()` loop only runs while the main window has key focus, and the brief gives the Settings window only Esc and Ctrl+W handling (§2, criterion 9); the menu bar does not run shortcuts in native-UI windows on Windows (§2). So with Settings frontmost on Windows, nothing is described that handles Ctrl+,. — Specify (and add to criterion 9 or a new criterion) that the Settings window's own key handling includes Ctrl+, to bring itself to front.

**[W-2]** Criterion 5 ("clamped to 8–32 everywhere, including when loading") is in tension with the v1-missing-`fontSize` rule (migrates to 11.6, "not clamped up to 8") and with criterion 6's exceptions; the exceptions are stated, but criterion 5 reads absolute. Reword criterion 5 as "clamped to 8–32 everywhere, except as provided by criterion 6" to avoid an unmeetable acceptance test.

**[W-3]** Acceptance has no criterion covering the README change (recording the Ctrl+, Shift-rule exception) or the new source file's README Layout-table entry, both of which the brief mandates in §2 and §5. Add them to the acceptance list or accept they are diff-gate checks.

### NIT findings

**[NIT-1]** "Default's" in the brief's title should be "Default's"→"Default text size" (stray apostrophe) — cosmetic only.

### Verdict
BLOCK — 2 blocking finding(s); must resolve before implementation starts.