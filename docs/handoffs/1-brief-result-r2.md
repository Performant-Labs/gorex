## Brief Review (Round 1)

### BLOCK findings
None.

### Resolved findings
**[R-1]** [B-1] "§3 says '**No guard is needed**' for headless `SetChecked` in tests, but §8's verified-source note on `SetChecked` says the opposite … Two parts of the brief directly disagree." — The amended brief makes §3 and §8 agree: both now state no guard is needed, supported by a re-run probe whose reported output (`SetChecked(dark): returned`, `after SetChecked: light=false dark=true`) is quoted in §8's excerpts section. The contradiction is gone; the probe's reality remains a needs-verification item (see NV-1).

**[R-2]** [B-2] "Criterion 11 and §4 'App lifecycle' are contradicted by the shown code … §5 'Reuse & extend-vs-new' lists no such change to `OnActivate`." — The brief now states the change explicitly in §4 ("**change `main.go:46-50`** to drop the `hasVisibleWindows` condition"), lists it in §5 ("**EXTEND** the `OnActivate` handler … drop `!hasVisibleWindows &&`"), and walks each non-regression path (no window at all, main window open, `OnWindowAllClosed`, ⌘-Tab). Criterion 11 is now consistent with the shown `main.go:46-50` and with the plan.

### Needs-verification findings
**[NV-1]** The probe results quoted in §8 (`SetChecked` safe headless after `NewMenu`; `Theme.SetSource`, `Focus`, `Close` blocking; `Invalidate` returning) — `scratchpad/mygoprobe/probe_test.go` and its `go test -count=1 -v` output are cited but not shown in this prompt. The probe file's contents and its recorded output would settle it.

**[NV-2]** All the pinned-mygo line claims (menu.go:243-264 update/postMain, app.go:591 FocusedWindow, no `performKeyEquivalent` override outside testhooks.go:58, menu.go:325-327 role submenu, menu.go:507-508 RoleClose, window.go Focus/Close/NewWindow blocking, loop.go:87-100 onMain waits, ui/keys.go:45 KeyComma, plugins/terminal/keys.go:21) — the mygo module's source is not excerpted here; showing those lines would settle each.

**[NV-3]** That Ctrl+, sends nothing in legacy terminal encoding on Windows/Linux (the basis of the README Shift-rule exception) — the brief itself concedes this is unverified ("legacy tables are not in the Go module"). The Ghostty native library's legacy encoding table, or the source the existing Ctrl+=/Ctrl− exception cites, would settle it.

**[NV-4]** That a v1 file's `fontSize` of 0 ("missing") is distinguishable from a stored 0 during migration, given the shown struct uses `omitempty` — implementation detail of the new parser; showing the planned parse logic or the v1 JSON semantics (Go's `json` unmarshals a missing float field as 0, so the migration must treat 0 as missing) would settle it. Note the shown `settings.go:31-33` already treats 0 (via `<6`) as invalid, so the "missing → 11.6" rule is consistent with today's behaviour, but the brief should confirm a hand-written `"fontSize": 0` in a v1 file is also treated as missing.

### WARN findings
**[W-1]** Criterion 10's parenthetical "⌘W (Ctrl+W on Windows and Linux)" mislabels the shortcut: per README.md:45 and the brief's own §4, Close Pane is ⌘W on macOS and Ctrl+Shift+W elsewhere; bare Ctrl+W elsewhere is the terminal's, and in Settings it closes via mygo's `RoleClose` default. Reword to "⌘W on macOS (Ctrl+Shift+W/Ctrl+W per §2 and §4 elsewhere)" so the criterion is testable as written on each platform.

**[W-2]** The write-gate is specified for unreadable/invalid and newer-version files, but §1's migration path (v1 → copy to `settings.v1.json` → overwrite) has no stated behaviour if the copy fails (disk full, permissions). State whether a failed copy aborts the migration (leaving the v1 file in place, read-only) — otherwise "must not destroy the user's file" (§1 Robustness) is untestable in that path.

**[W-3]** Criterion 7 says Default "cannot be removed or renamed by editing the file", but §1 only specifies restoration of missing `name`/`fontSize`; a hand-set `"name": "My Default"` is restored to "Default" per §1's last line, yet no acceptance test asserts the *rename* half, and the Settings sidebar label ("Default", "built-in") vs. the stored name needs one stated rule. Clarify that the UI label is always "Default" regardless of any stored name (or that the name is always rewritten to "Default" on load).

### NIT findings
**[NIT-1]** Title "General and Default's text size" retains the stray apostrophe from the earlier round (carried forward); "General and Default text size" reads better. Cosmetic only.

### Verdict
PASS — no BLOCK findings; implementation may proceed.