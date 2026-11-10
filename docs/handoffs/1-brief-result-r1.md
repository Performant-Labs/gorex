## Brief Review (Round 1)

### BLOCK findings
None.

### Needs-verification findings
**[NV-1]** §4/§8 claim that `run` returns when `a.win == nil` except for New Tab (`commands.go:214-226`), and that `a.do` "changes the state on the main thread, from a menu, and draws" (`commands.go:204`) — the commands.go excerpt is truncated at line 153, so neither is shown. Settle by showing `commands.go:200-230`.
**[NV-2]** §8 claim that `shortcuts()` handles commands only on Windows by looping over the `commands` table (`commands.go:331-344`, quoted in §8) — the excerpt does not reach those lines. Settle by showing `commands.go:321-344` and confirming it matches the quoted text.
**[NV-3]** §8 claims about the pinned mygo — `SetChecked`/`postMain` returning headless, `onMain` blocking, `Focus`/`Close` calling `onMain`, `Invalidate` safe on content-less windows, `MenuItemClicked` using `FocusedWindow()`, `RoleClose` closing the clicked window, `prepareItem` only filling a role submenu when none is set, `ui.KeyComma` existing and `,` being a valid accelerator key — none of the mygo source or the probe output is in this prompt. Settle by reading `docs/handoffs/0001-settings-window/evidence/mygoprobe/output.txt` and the cited mygo file:lines (menu.go, loop.go, window.go, content.go, app.go, ui/keys.go).
**[NV-4]** §2's justification that Ctrl+, sends nothing in legacy terminal encoding (the README exception) — the brief itself concedes this is not verified: Ghostty's legacy encode tables are not in the Go module. Settle with a Ghostty source excerpt or a live encoding test of Ctrl+, in a legacy-mode terminal.
**[NV-5]** §6 claim that `nopConn.Read` blocks forever and `terminal.New` starts a reader goroutine that `Close` cannot end (`plugins/terminal/terminal.go:214-217, :250-253, :603-605`; `state.go` excerpt shows `nopConn` but not the terminal plugin) — settle by showing those terminal.go lines.

### WARN findings
**[W-1]** Esc's behaviour in Settings is specified only for Windows (criterion 9). On macOS and Linux, whether Esc closes the Settings window is unspecified; a user pressing Esc on macOS gets nothing. Recommend stating the intent (Esc closes Settings everywhere, or only on Windows) so the diff gate can check it.
**[W-2]** §1 says "a file without `version` is v1" but does not say what an explicit `"version": 1` does — migrate or reject? Recommend one sentence pinning it (presumably migrate like an unversioned file).
**[W-3]** The brief prescribes implementation shape in unusual detail (the Settings key-handling loop on Windows, the app-wide field, the exact `OnActivate` rewrite, test seams). Much of this is diff-gate material; a brief this prescriptive risks the diff being judged against prose rather than behaviour. Acceptable given the cited evidence, but the diff gate should treat the acceptance criteria, not the prose design, as binding.

### NIT findings
**[NIT-1]** §2 describes the slider as "8–32 pt in 1 pt steps" while also requiring a thumb at 11.6 "between steps"; the two sentences read oddly together even though the behaviour is coherent.
**[NIT-2]** Criterion 5 defers to criterion 6 and criterion 6 back-references criterion 5; a single statement of the clamp/missing-value rules would be less circular.

### Verdict
PASS — no BLOCK findings; implementation may proceed.