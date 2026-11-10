# Handoff-A-dup: Phase 7 - Settings window, General and Default's text size, settings v2  (anti-duplication gate)

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Diff base:** main (merge-base)   **Diff head:** cac5fce
**Reuse map:** docs/handoffs/0001-settings-window/brief.md (section 5)
**Verdict:** PASS

## Summary
F extended the objects the Reuse map named and built no parallel path. The one new object, the Settings window in `preferences.go`, is justified in writing in brief section 5 (no settings UI or second window existed to extend).

## Findings
| # | Severity | File:line | Finding | Suggested fix |
|---|---|---|---|---|
| 1 | warn | `preferences.go:216` | The preview terminal's `terminal.New(terminal.Options{Conn: nopConn{}, Transparent: true, Font: termFont, Theme: lightTerm, DarkTheme: darkTerm})` repeats the call at `state.go:205` (`attach`). The brief says to reuse that approach, and the shape is the same one-liner, so this is minor. | Optional follow-up: a small `displayTerminal()` helper used by both, if a third caller appears. |
| 2 | warn | `preferences.go:97-101` | Settings' Windows key loop mirrors the `shortcuts()` loop (`commands.go:391`). The brief asks for it explicitly, and it filters on `AppWide`. It is not a parallel command table. | None. |

Checked and clean:
- `setAppearance` and `setFontSize` (`settings.go`) are the only writers. Settings, menu, palette and shortcuts all call them. Settings never touches `prefs`, `termFont` or `saveSettings`.
- `cmdSettings` and the `AppWide` field extend the `command` table. The menu goes through `item()`, whose Show Tab items were folded into it, so there is one window check (`runsIn`), not two.
- `App.run`, `OnActivate` and `quitAndEnd` were extended in place.
- Window creation follows `App.open`. `colors`, `colorsOf` and `iconButton` are reused. No second colour set was added.
- The v2 format and the migration stay in `settings.go`.
- The new seams (`newSettingsWindow`, `focusWindow`, `closeWindow`, `setThemeSource`) are the ones brief section 6 specifies.
- The new file `preferences.go` is one lowercase word and is listed in the README's Layout table.

## Notes for F
None; no block findings.
