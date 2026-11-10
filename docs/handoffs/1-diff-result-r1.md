## Implementation Review (Round 1)

### BLOCK findings
None.

### Needs-verification findings
- **[NV-1]** preferences.go:76-92 (`settingsKeys`) — the claim that on Windows the loop over `commands` is the only path for Ctrl+, and that Esc/Ctrl+W/Ctrl+Shift+W reliably reach `settingsView`'s key handling before the terminal/main-window handling — would be settled by running the app (or at least `TestSettingsKeysOnWindows`) on a Windows host; the handoff itself lists this test as never run.
- **[NV-2]** preferences.go:160-174 (`appearanceChoice`) — the claim that `ui.RadioBase` with a local `selected` copy reports `Changed()` on click and drives selection correctly — mygo's `ui` widget source (`ui/base.go` or equivalent for `RadioBase`) would settle it; no excerpt in this prompt shows `RadioBase`.
- **[NV-3]** preferences.go:193-196 — the claim that `ui.Slider(...).Step(1)` snaps clicks to whole points while the thumb renders at 11.6 — the Evidence Appendix cites `ui/base.go:122-124` for `Step` semantics, but the excerpt is doc-comment only and does not show thumb rendering at a non-step value; a mygo slider source excerpt or probe would settle it.
- **[NV-4]** README.md:61-64 — the claim that "in the terminal's legacy encoding Ctrl+, sends nothing" — the brief itself notes this is not verified against source (Ghostty's legacy tables are not in the Go module); it is issue #1's assertion only.
- **[NV-5]** commands.go:250-262 (`run`) — the claim that `focusWindow`/`mygo.RunOnMain(a.open)` called from inside another window's frame (palette "Settings…", Ctrl+Shift+T in Settings on Windows with no main window) does not deadlock or misbehave at runtime — the handoff lists it as a known unverified runtime behaviour; a live run would settle it.

### WARN findings
- **[W-1]** settings.go (`saveSettings`) — a failed write of `settings.json` (disk full, permissions changed mid-run) is ignored, `readOnly` stays "", and the Settings footer still says "Changes apply and save as you make them." This is in tension with §2's closing rule ("The user is never left believing a change was saved when it was not"), even though the brief names only the three read-only reasons. Recommend checking the `os.WriteFile` error in `saveSettings` and setting a read-only reason for the rest of the run.
- **[W-2]** preferences.go:76-82 — on Windows, `settingsKeys`' command loop can reach `cmdNewTab`, whose `a.run` path with `a.win == nil` calls `mygo.RunOnMain(a.open)` and waits; if a main loop ever serves Settings' frame off the main thread, this could stall a frame. The brief mandates this behaviour, and the handoff flags it, but it deserves live Windows verification before ship (see NV-1/NV-5).
- **[W-3]** preferences.go:198-210 — the preview terminal is `Resize(60, 4)` once and never resized again; the preview box width is window-relative, so at other window widths the preview content no longer matches the box (clipped or under-filled). Recommend re-resizing the preview terminal to the box's column count each frame, or fixing the preview box width.

### NIT findings
- **[NIT-1]** commands.go:145 — the per-Show-Tab `command{}` values are heap-allocated per `menu()` call and matched only by the closure; harmless, but a comment noting they are deliberately not in `commands` (so `runsIn` treats them as main-window-only) would help.
- **[NIT-2]** preferences.go:70 — `a.settingsPage` persists after the Settings window closes; reopening Settings returns to the last page. Probably desirable (macOS convention), but unstated in the brief either way.

### Verdict
PASS — no BLOCK findings; testing may proceed.