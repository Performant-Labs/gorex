## Brief Review (Round 1)

### BLOCK findings
None.

### Needs-verification findings
**[NV-1]** "Today loading accepts 6–40 while `setFontSize` clamps to 8–32" — this claim about current `settings.go` behavior is asserted, not evidenced. A shown excerpt of the load/validation code and `setFontSize` in `settings.go` would settle it; if load actually accepts 8–32 only, acceptance 5's "including when loading" is already satisfied and the migration clamping note is redundant (harmless, but should be accurate).

**[NV-2]** "Native-UI windows run no menu shortcuts on Windows (`shortcuts()` in `commands.go`)" — a runtime claim about the existing window/menu plumbing. An excerpt of `shortcuts()` in `commands.go` showing that menu accelerators are not dispatched in native windows would settle it.

**[NV-3]** "View ▸ Appearance's radio items … their `Checked` state is fixed at that point" (built once in `App.menu()`) — a claim about current menu construction. An excerpt of `App.menu()`/`commands.go` showing where the radio items are built and whether `Checked` is ever updated would settle it.

**[NV-4]** "In legacy terminal encoding Ctrl+, sends nothing" — a factual claim about terminal input encoding used to justify the README exception. A reference to the existing Ctrl+=/Ctrl+- exception text in README.md (or the encoding table it cites) would settle it.

**[NV-5]** "⌘, … does nothing" today — a claim about current behavior. An excerpt of the menu construction / shortcut handling showing no ⌘, binding would settle it.

### WARN findings
**[W-1]** Slider step semantics vs. the 11.6 pt default are underspecified. The slider is "8–32 pt in 1 pt steps" while the default is 11.6 and non-whole values "stay as they are until changed". Unspecified: where the slider thumb sits for a non-whole value; what −/+ do from 11.6 (10.6/12.6, or 11/12?); what the slider shows after the user drags (round to whole?). Recommend specifying: thumb rounds to nearest whole, −/+ step 1 pt from the current value, and the acceptance test for criterion 4 pins the 11.6 → −/→ + behavior.

**[W-2]** Acceptance 9 covers only Windows for Esc/Ctrl+W, but the brief's Windows section is the only place window-close keyboard handling is mentioned. macOS/Linux Esc and ⌘W/Ctrl+W behavior in the Settings window is unspecified — state explicitly whether they apply (and if not, why) so the diff gate can check it.

**[W-3]** Migration write timing is ambiguous for the "invalid file" path. Criterion 8 says an invalid file "never … loses the user's file", but the brief doesn't say whether an invalid file triggers the v1 backup + rewrite (risking overwriting a hand-corruptable-but-recoverable file with defaults) or is merely ignored in memory. Recommend: invalid/unparseable file → defaults in memory, file untouched on disk; only a *recognized v1* file is backed up and rewritten.

**[W-4]** "Single-window bring-to-front" testability (criterion 10): the brief asks for headless tests of this logic, but doesn't require the bring-to-front decision (open vs. focus) to be extracted as testable logic separate from the windowing API. Recommend stating that open-or-focus is a pure decision (e.g., a function returning focus vs. create) so the test can exist without a window; otherwise criterion 10 is partially untestable as written.

### NIT findings
**[NIT-1]** The v2 example shows `profiles.default.name` as "Default" and the rule says name is always "Default" — consider stating that a hand-edited non-"Default" name is normalized back, so criterion 7's "renamed by editing the file" is unambiguous.

**[NIT-2]** Criterion 3 mentions "a change from … the palette" changing appearance, but the palette is only specified to gain a "Settings…" command; if appearance palette commands already exist that's fine, otherwise drop "palette" from criterion 3 or specify those commands.

### Verdict
PASS — no BLOCK findings; implementation may proceed, with WARNs resolved or consciously accepted first.