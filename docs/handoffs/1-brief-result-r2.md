## Brief Review (Round 1)

### BLOCK findings
None.

### Resolved findings
**[R-1]** "[B-1] Criterion 10 requires Ctrl+Shift+W to close Settings on Windows, but the brief provides no mechanism for it there — and its own premises rule one out." The amended §2 now specifies: "The Settings window needs its own Esc, Ctrl+W and **Ctrl+Shift+W** handling to close it… on Windows no menu shortcut runs in a native-UI window, so Settings' own key handling must match it and close the window through `closeWindow` (§6)", and criterion 9/10 match ("On Windows, Esc, Ctrl+W and Ctrl+Shift+W close the Settings window"). The mechanism (Settings' own key handling on Windows, exactly four keys) is now specified and consistent with the shown premise that native-UI windows on Windows run no menu shortcuts.

**[R-2]** "[B-2] Criterion 10 and §4 contradict each other about ⌘T from Settings." The amended criterion 10 now exempts New Tab explicitly: "⌘T (Ctrl+Shift+T on Linux) never acts on a hidden main window either: New Tab is app-wide (§4), so it first opens the main window if it is closed or brings it to the front, and only then opens the tab there", and §4 matches ("New Tab… brings the main window to the front before it opens the tab, so from Settings it never acts on a hidden main window"). The contradiction is gone; ⌘D, ⌘K and ⌘1–⌘9 remain correctly blocked by the window check.

### Needs-verification findings
**[NV-1]** The mygo behavioral claims in §8 — `SetChecked` safe headless after `NewMenu` (probe ran on darwin only, extrapolated to Linux/Windows), `Focus`/`Close`/`NewWindow`/`Theme.SetSource` blocking headless via `onMain`, `Invalidate`'s early return on empty content, `OnClosed` being a method on `*Window`, `app.go:591` `FocusedWindow()` for macOS menu clicks, `menu.go:85-88` window-of-menu-bar off darwin, `RoleClose` closing the clicked window, off-darwin `SetMenu` on every window, `OnActivate` firing only from `applicationShouldHandleReopen` — the mygo source lines are cited but the excerpts are paraphrases, not shown verbatim in this prompt. Settling evidence: the quoted lines from mygo `v0.3.5-0.20261008155018-bad73d119dc8` (`menu.go:243-264, 325-327, 507-508, 575-576, 672-681`, `loop.go:19-31, 87-100`, `window.go:513-519/534/607`, `content.go:145-148`, `app.go:319-323/591`, `internal/darwin/app.go:305-306`) and the committed `docs/handoffs/0001-settings-window/evidence/mygoprobe/output.txt`.

**[NV-2]** `ui.KeyComma` exists in mygo's `ui/keys.go:45` and `,` parses as an accelerator key — settle with the quoted `ui/keys.go` lines and `internal/accelerator`'s acceptance of `,`.

**[NV-3]** "In legacy terminal encoding Ctrl+, sends nothing" — the brief itself concedes this is unverified (issue #1's text; the legacy tables live in Ghostty's native library, not the Go module). Settle with issue #1 or the terminal plugin's legacy encoding table.

**[NV-4]** `a.win` is set to nil when the main window closes — the `OnActivate` change's behavior analysis ("Main window open… `a.win != nil`, so it opens nothing") rests on it; the shown `main.go` excerpt (truncated at line 110, and `reset()` at `:155-159` clearing only per-window state) does not show where `a.win` is cleared. Settle with the main window's `OnClosed` handler in `main.go`.

**[NV-5]** "In legacy terminal encoding Ctrl+, sends nothing" aside, the claim that `shortcuts()` runs only in the main window's view (so Settings on Windows must handle Ctrl+, itself) is asserted from `commands.go`/`view.go:37`; the shown excerpts show `a.shortcuts(c)` in `a.view` and the Windows loop in the cited-but-not-shown `commands.go:331-338`, consistent but not conclusive that no other surface dispatches shortcuts. Settle with the full `shortcuts()` and its callers.

### WARN findings
**[W-1]** (carried, unaddressed) Migration retry vs an existing `settings.v1.json` from a previous failed run: whether the retry overwrites the backup, keeps it, or treats migration as done is unspecified. Recommend stating it (overwrite is natural, since the v1 file was left unchanged).

**[W-2]** (carried, unaddressed) When the v2 file is first written after a successful migration (at load, or only at the next `saveSettings`) is unspecified; a crash in between repeats the migration. Benign but the criterion-12 tests should pin it.

**[W-3]** (carried, unaddressed) The failed-backup and unparseable/newer-version paths make the app silently read-only for the whole run with no user-visible signal. Recommend stating that the silence is intended, or surfacing it in the window's footer.

**[W-4]** (carried, unaddressed) The headless-`SetChecked` probe ran on darwin only; CI runs the criterion-12 tests on all platforms. If a backend diverges, tests hang rather than fail. Recommend a timeout guard on those tests.

### NIT findings
**[NIT-1]** (carried) Title: "General and Default's text size" — stray apostrophe; should be "Default text size".

**[NIT-2]** (carried) The slider is "8–32 pt in 1 pt steps" while the default is 11.6; the brief never says what the thumb shows for a non-integral value (between steps, or snapped for display). One sentence would remove the ambiguity.

### Verdict
PASS — no BLOCK findings; implementation may proceed.