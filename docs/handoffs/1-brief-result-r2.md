## Brief Review (Round 1)

### BLOCK findings
None.

### Resolved findings
**[R-1]** "[B-1] Acceptance criterion 6 contradicts the clamping rule. AC6 states 'A v1 file migrates to v2 on first load: appearance and text size are unchanged' … for a v1 file with, say, `fontSize: 40`, the text size is *not* unchanged …" — the amended brief now carries the exception in both places: §1 ("A user's appearance and text size look exactly the same after the migration, except that a text size outside 8–32 is clamped") and AC6 ("unchanged (`""` becomes `"system"`), except that a text size outside 8–32 is clamped per criterion 5 and a missing text size becomes 11.6"). The contradiction is gone; AC6 is now testable as written.

### Needs-verification findings
**[NV-1]** "`SetChecked` … touches the backend only when the item has a `uid` … whether the post is safe under `go test` after `menu()` ran is not verified here, hence the guard in §3" — mygo `menu.go:219-226`, `:247-264`, `:302-308` and `loop.go:18-31`; the brief itself flags this as unverified, so the guard's necessity/sufficiency (does `backend().Signal()` panic headless?) must be confirmed against those excerpts.
**[NV-2]** "mygo's darwin surface does not override `performKeyEquivalent` … the macOS Window menu role has no Close item" (§4, underpinning AC10's macOS claims) — mygo `menu.go:672-681`, `:575-576`, `:507-508`, and `internal/darwin` (only `testhooks.go:58` mentions the selector); the grep result is described but the darwin source is not shown here.
**[NV-3]** "`prepareItem` fills a role's submenu only when none is set" (§5 Menu placement) — mygo `menu.go:325-327`, quoted in §8; consistent with the quoted text, but the surrounding construction path is not shown.
**[NV-4]** "`NewWindow` off the main thread calls `App.waitReady()`" and "`WindowOptions.StateKey string`" — mygo `window.go:294-295`, `:153-160`.
**[NV-5]** "In legacy terminal encoding Ctrl+, sends nothing" (the README exception rationale) — a terminal control-character table showing 0x2C maps to no C0 code.
**[NV-6]** `ui.KeyComma` exists in mygo's key constants (§5 adds `ui.KeyComma: ","` to `shortcut.String()`'s map, whose shown excerpt at `commands.go:116-121` does not include it) — mygo key-constant definitions.
**[NV-7]** `ui.NewTester`'s `Key`/frame semantics suffice for the planned `settingsView` headless tests — the mygo tester source; `app_test.go:303-337` shows the pattern works for the palette, not for sliders.

### WARN findings
**[W-1]** (carried from round 1, unresolved) The 8–32 slider with 1 pt steps cannot represent the shipped default 11.6 pt. The brief still only says "a value that is not a whole number stays as it is until it is changed" — it specifies the label, not where the thumb sits at 11.6 nor which whole value the first drag/click commits. AC1 and AC4 are testable only if this is pinned down; recommend specifying thumb position (floor or round) and first-commit value in §2 "Default page".
**[W-2]** (carried, partly addressed) AC9 covers Esc/Ctrl+W on Windows; the Ctrl+Shift+W-vs-Ctrl+W distinction in §4 still has no explicit Linux coverage beyond AC10's wording. Low risk; recommend one clause in AC9 naming the platforms each shortcut applies on.

### NIT findings
**[NIT-1]** §1's v2 example could note inline that a non-whole `fontSize` (11.6) persists until the user changes it, since the slider's whole-point steps will surprise the implementer otherwise.
**[NIT-2]** AC11's Dock-icon behavior is macOS-only and untestable in `go test` — the criterion's "tested" scope should reference criterion 12's headless boundary explicitly.

### Verdict
PASS — no BLOCK findings; implementation may proceed.