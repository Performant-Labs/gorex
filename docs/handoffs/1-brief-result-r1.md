## Brief Review (Round 1)

### BLOCK findings
None.

### Needs-verification findings
**[NV-1]** "In legacy terminal encoding Ctrl+, sends nothing" (§2, the justification for the Shift-rule exception) — the brief itself admits this is not verified against source; the evidence would be the legacy encoding tables in Ghostty's native library behind `plugins/terminal/internal/vt/encode.go` (not in the Go module), or a runtime probe on Windows/Linux.

**[NV-2]** "Native-UI windows run no menu shortcuts on Windows" (§2, §4) — this drives the entire Windows key-handling design; evidence would be the mygo Windows backend's key-equivalent/menu-accelerator dispatch code, or a probe on Windows. The shown `commands.go:331-338` excerpt only shows GoRex's own `shortcuts()` loop, not mygo's Windows behavior.

**[NV-3]** "`ui.KeyComma` exists at `ui/keys.go:45`" — cited but no excerpt of `ui/keys.go` is shown in this prompt; the excerpt of that line would settle it.

**[NV-4]** "`,` is a valid accelerator key" (cited to handoff-A round 2, `internal/accelerator`) — that source is not shown here; an excerpt of the accelerator parser accepting `,` would settle it.

**[NV-5]** The probe results for `SetChecked` / `Invalidate` returning headless and `Focus` / `Close` / `Theme.SetSource` blocking (§8) — the committed `docs/handoffs/0001-settings-window/evidence/mygoprobe/output.txt` and `probe_test.go` are described but their contents are not shown in this prompt.

**[NV-6]** "Linux and Windows share the same `update` / `postMain` path" for `SetChecked` headless (§6 "No test hangs") — the probe ran on darwin only; a run of the probe (or the non-darwin menu/loop code paths) on Linux/Windows would settle it.

**[NV-7]** "`Application.SetMenu` gives every window the app menu bar off darwin" (`app.go:319-323`) and "`RoleClose` defaults to CmdOrCtrl+W and closes the clicked window" (`menu.go:507-508`, `:575-576`) — cited with line numbers but the surrounding text is not shown in this prompt; those excerpts would settle it.

### WARN findings
**[W-1]** §1 does not specify what happens to a v1 (or v2) `appearance` value that is not `""`/`"light"`/`"dark"`/`"system"` (a hand edit like `"blue"`). "Default cannot be removed" restores Default's fields but nothing restores an invalid top-level `appearance`. Recommend: specify normalization to `"system"` on load/migration.

**[W-2]** The read-only "Write gate" (§6) is triggered only at load time (newer version, unparsable, backup failure). A save that fails later in the run (disk fills after launch) is not covered: the footer would still say "Changes apply and save as you make them" while `saveSettings` silently fails, contradicting §2's "The user is never left believing a change was saved when it was not." Recommend: either propagate a `saveSettings` failure into the read-only state and footer, or explicitly scope the guarantee to load-time detection.

**[W-3]** Criterion 9/10 Windows behavior (Settings' own handling of Esc, Ctrl+W, Ctrl+Shift+W, Ctrl+, and "exactly these four keys") is not covered by criterion 12's test requirement, and no test seam is specified for it — the key handling lives in `settingsView`, but the brief does not say how the four-key-only rule is pinned by a test (e.g. driving key events at `settingsView` via `ui.NewTester` on any OS). Recommend naming the test that pins the four-key set, or acknowledging it as manual-only.

**[W-4]** The Settings window's Esc/Ctrl+W close shortcuts on Windows are new user-facing shortcuts not recorded in the README Shortcuts table; criterion 13 only adds the Ctrl+, row. Recommend deciding whether the Settings-window close shortcuts belong in the table.

**[W-5]** Criterion 5's cross-reference ("except as provided by criterion 6") is loose: criterion 6 describes migration, including the 11.6 restoration, while criterion 5 carries the clamp rule. Also §1's rule "Only a value greater than 0 is clamped" applies to v1; the analogous v2 case (`fontSize: 0` in a v2 file) is only implied via "Default cannot be removed" treating 0 as missing. Recommend stating the v2 zero-value rule explicitly in criterion 5 or 7.

### NIT findings
**[NIT-1]** §2 describes the slider as "over 8–32 pt in 1 pt steps" while also requiring the thumb to sit "between steps" at 11.6; fine, but the acceptance criteria (4, 5) never pin the thumb position between steps — a small testable assertion ("thumb at the exact value") is left only in prose.

**[NIT-2]** The two "§8" sections ("Verified facts" and "Verified source excerpts") share a number; renumber for the handoff.

### Verdict
PASS — no BLOCK findings; implementation may proceed.