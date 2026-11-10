## Brief Review (Round 1)

### BLOCK findings
**[B-1]** Acceptance criterion 6 contradicts the clamping rule. AC6 states "A v1 file migrates to v2 on first load: appearance and text size are unchanged", and §1 states "A user's appearance and text size look exactly the same after the migration." But §1 also states "Today loading accepts 6–40 while `setFontSize` clamps to 8–32, so a v1 value outside 8–32 is clamped on load" — for a v1 file with, say, `fontSize: 40`, the text size is *not* unchanged and does not look the same. AC5 reinforces the clamping. Two parts of the brief disagree, making AC6 untestable as written. Fix: AC6 should read "unchanged, except that a value outside 8–32 is clamped per criterion 5" (and §1's "exactly the same" softened likewise), or the brief must pick one behavior.

### Needs-verification findings
**[NV-1]** "Today loading accepts 6–40 while `setFontSize` clamps to 8–32" — the load-time validation range in `settings.go` (the current load code) and the clamp in `setFontSize`, shown as file:line excerpts.
**[NV-2]** "Native-UI windows run no menu shortcuts on Windows (`shortcuts()` in `commands.go`)" and "Ctrl+, reaches the main window through the existing `command` table loop in `shortcuts()`" — the shown body of `shortcuts()` in `commands.go`, confirming it iterates the `command` table so `cmdSettings` is picked up with no hand-written case.
**[NV-3]** "mygo's darwin surface does not override `performKeyEquivalent`, and the macOS Window menu role has no Close item" (§4) — the mygo darwin surface source and the default role menu construction; this underpins all of AC10's macOS claims.
**[NV-4]** "`{Role: mygo.RoleAppMenu}` uses mygo's default submenu, and mygo replaces it only when `Submenu` is set" (§5 Menu placement) — mygo's menu-building code for roles.
**[NV-5]** "`MenuItem.SetChecked` … if `SetChecked` cannot run under `go test` (no backend), guard it" — the mygo `SetChecked` implementation, confirming whether it panics or no-ops headless; the guard may be unnecessary or insufficient.
**[NV-6]** "`mygo.NewWindow` off the main thread waits for `App.waitReady`" and that `StateKey "settings"` is a supported `WindowOptions` field — mygo's `NewWindow` source and the `WindowOptions` struct definition.
**[NV-7]** "In legacy terminal encoding Ctrl+, sends nothing" (the README exception rationale) — a terminal-encoding reference (e.g. the control-character table showing 0x2C maps to no C0 control code) confirming Ctrl+, is unrepresentable.
**[NV-8]** "`OnActivate` reopens the main window only when `!hasVisibleWindows && a.win == nil`" and "`App.run`'s 'no main window' exception … covers only New Tab" — the shown excerpts of `main.go` / `run` matching these descriptions, since AC10–11 depend on the described current behavior.
**[NV-9]** "⌘W in Settings would run Close Pane … and ⌘T, ⌘D and ⌘K would act on the hidden main window" — `item()`'s `Click` signature in `commands.go` shown ignoring the `*mygo.Window`, confirming the hazard is real.
**[NV-10]** `attach`'s display-only terminal (`terminal.New`, `nopConn{}`, `Feed`) and `ui.NewTester`/`TestPaletteScrolls` exist and behave as "REUSE" assumes — shown excerpts from the relevant files.

### WARN findings
**[W-1]** The 8–32 slider with 1 pt steps cannot represent the shipped default 11.6 pt. The brief's answer ("a value that is not a whole number stays as it is until it is changed") tells the label what to show but not the slider thumb what to display at 11.6, nor which whole value a click/drag commits first. Recommend specifying: the thumb sits at floor/round(11.6), and the first drag commits the slider's whole value. Otherwise AC1's "Default" page renders in an under-specified state.
**[W-2]** AC9 covers Esc and Ctrl+W on Windows only; macOS Esc behavior for the Settings window (and whether Cmd+W is the close path per §4, which it is) is fine, but the non-macOS Ctrl+Shift+W "Close Pane" item vs Ctrl+W on Settings is asserted in §4 without an acceptance criterion covering Linux. AC10 partially covers it; recommend making it explicit that Ctrl+W closing Settings is tested on all platforms where it applies.
**[W-3]** §6 Write gate says an invalid file makes the app "run on defaults", while §1 Robustness says an invalid file "must not destroy the user's file". If the app runs on defaults and any subsequent setting change calls `saveSettings()`, the read-only gate must hold for the *session*, not just the load. Recommend the brief state explicitly that read-only status persists until a successful load of a writable file, and that AC8's "never loses the user's file" is tested with a post-load settings mutation.

### NIT findings
**[NIT-1]** The example v2 JSON shows `"fontSize": 11.6` while the slider is whole-point stepped; a comment noting non-whole values persist until changed would help implementers.
**[NIT-2]** "`+ New Profile…` footer is delivery 2: leave it out" vs wireframe 1 which presumably shows it — call out explicitly that wireframe 1 should be read with the footer cropped/ignored, so the implementer doesn't ship it from the image.

### Verdict
BLOCK — 1 blocking finding; must resolve before implementation starts.