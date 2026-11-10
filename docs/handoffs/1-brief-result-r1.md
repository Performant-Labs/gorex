## Brief Review (Round 1)

### BLOCK findings
None.

### Needs-verification findings
- **[NV-1]** The claim that Ctrl+, sends nothing in legacy terminal encoding on Windows/Linux (the justification for the Shift-rule exception) — the brief itself admits the Ghostty legacy encode tables are not in the Go module. Evidence that would settle it: a Ghostty/legacy-encoding table excerpt (or `plugins/terminal/internal/vt` encode output) showing Ctrl+Comma maps to no byte.
- **[NV-2]** `SetChecked` / `Invalidate` return headless on Linux and Windows — the probe ran on darwin only. Evidence: the same probe (`docs/handoffs/0001-settings-window/evidence/mygoprobe/probe_test.go`) run on a Linux (and ideally Windows) backend.
- **[NV-3]** On Linux, Close Pane's menu accelerator (Ctrl+Shift+W) is delivered to the Settings window's menu bar with `win` = Settings, so the §4 window check closes Settings — rests on the cited `app.go:319-323` / `menu.go:85-88` behavior, excerpts not shown verbatim here. Evidence: those mygo lines plus an observed accelerator click's `win` argument.
- **[NV-4]** On macOS, menu accelerators reach the main menu even when a native-UI Settings window is key (no `performKeyEquivalent` override) — supported by the search claim over `internal/darwin`, which is an assertion about absence, not a shown excerpt. Evidence: the mygo darwin app/menu source showing key-equivalent routing to the main menu.
- **[NV-5]** ⌘-Tab / app activation does not fire `applicationShouldHandleReopen` (so the new `OnActivate` handler does not reopen the main window on every activation) — rests on AppKit behavior; evidence: Apple documentation for `applicationShouldHandleReopen:hasVisibleWindows:` or the mygo darwin binding's invocation sites.

### WARN findings
- **[W-1]** §2 ("Windows") says "on Linux and macOS it [Ctrl+Shift+W] reaches Settings as Close Pane's menu accelerator" — but on macOS `cmdClosePane.Accel` is `CmdOrCtrl+W` (⌘W, commands.go:32) and darwin's `init()` returns before promoting `Other` (commands.go:74-77), so Ctrl+Shift+W matches no menu accelerator on macOS and does nothing there. Criterion 10 correctly describes macOS via ⌘W only. Fix the §2 sentence to exclude macOS; otherwise it invites implementing a nonexistent macOS path.
- **[W-2]** Unspecified v2 edge case: a v2 file with an **unknown appearance string** (e.g. `"appearance": "purple"`). §1 defines the valid set and v1 `""` normalisation, but not v2 garbage. The excerpt of `applyAppearance` is truncated, so today's default branch is unshown. Specify the behaviour (fall back to System read-only, or System and write).
- **[W-3]** Unspecified: a file with an explicit `"version": 1`. §1 says "A file without `version` is v1" — a hand-edited `"version": 1` is neither "without version" nor "newer than 2". Say whether it is treated as v1 (migrated) or invalid (read-only defaults).
- **[W-4]** The "unreadable file" footer string ("… settings.json could not be read") conflates an unreadable file with an unparseable one, and §6's write gate lists "newer version" and "could not be parsed" but not "could not be read" (it appears only in §2). One list, three reasons, in both places, so the tests in criterion 8 cover all three.
- **[W-5]** Criterion 5 and criterion 6 cross-reference each other for the missing-`fontSize` exception; a reader must hop between them to reconstruct the rule. Consider stating the full clamp/missing matrix once (value 0 or absent → 11.6; 0 < v < 8 → 8; v > 32 → 32) in one place.

### NIT findings
- **[NIT-1]** §5's `cmdSettings` spec gives `Keys: "⌘,"` but off darwin `init()` overwrites `Keys` with the `Other` string ("Ctrl+,"); fine, but the README row (criterion 13) should be written from that, not from "⌘,".
- **[NIT-2]** §6's seam `newSettingsWindow` production body shows `OnClosed(…)` "clears `a.settingsWin`" — also state whether it drops the preview terminal (§5 says one terminal per App for the app's life; it survives, but the closure should be explicit that it clears only the window field).

### Verdict
PASS — no BLOCK findings; implementation may proceed.