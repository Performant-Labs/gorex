# Handoff-A: Phase 4 - Settings window, General and Default's text size, settings v2  (up-front plan review, round 5)

This is Phase 4 in the pipeline table. `architecture-reviewer.md` calls the same up-front A review "Phase 3". This round re-reviews the brief as amended in `6238bf7`. It is the arch gate's fifth review, so only a correctness finding can block.

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Brief reviewed:** `docs/handoffs/0001-settings-window/brief.md` (at `6238bf7`)   **Reuse map:** brief §5 "Reuse & extend-vs-new" (there is still no `survey.md`)   **Wireframe:** `docs/wireframes/settings-profiles/1-general.png`; for the Default page, only the text-size row and preview of `2-profile-overrides.png` (unchanged since round 2)
**Verdict:** PASS

## Summary

PASS. The round-4 block is resolved. With no main window, Quit and End All Sessions now does nothing, as it does today: `quitAndEnd` returns at once when `a.win == nil`, and a test pins that `a.quitting` stays false. With the main window open, it is app-wide and runs through `a.do`. Both round-4 warns are adopted. This round again checked each app-wide command's path, both with no main window and from Settings, against the source, and found nothing that blocks. Three warns remain, all small and advisory. The gate converges: rounds 3 and 4 each raised a new block, and this round raises none.

## Round-4 findings: status

| R4 # | Status | Where in the brief |
|---|---|---|
| 1 block: Quit and End All Sessions with `a.win == nil` quits and leaves every session running | resolved | §4 app-wide list now reads "… and Quit and End All Sessions while the main window is open (below)". The new bullet "Quit and End All Sessions needs the session server's connection" says: through `a.do` when the main window is open, including from Settings; nothing when `a.win == nil`; `quitAndEnd` returns at once when `a.win == nil`; a test checks that `a.quitting` stays false. §4 "A click with no window" adds the exception. §8 adds `main.go:128` and `internal/rex/client.go:115-131`. Checked against `commands.go:302-314`, `main.go:117-130` and `app_test.go:45-72`. No existing test calls `quitAndEnd` or reads `a.quitting`, so the new guard breaks none of them. |
| 2 warn: two Ctrl+, paths in Settings on Windows | resolved | §2 "Windows": the app-wide loop is the **only** path for Ctrl+,, with no separate hand-written case. |
| 3 warn: name the nil-window test's command | resolved | §4 "A click with no window": the test uses Bigger Text, or an Appearance command through the Theme seam, and never New Tab or Quit and End All Sessions. |

## Findings

| # | Severity | Plan element | Drift dimension | Finding | Suggested fix |
|---|---|---|---|---|---|
| 1 | warn | §4 "One window check, one classification" (the field "runs whatever window is in front, and also when there is no main window") together with "A click with no window" ("except Quit and End All Sessions with `a.win == nil`") and "`quitAndEnd` returns at once when `a.win == nil`" | abstraction level: one rule in two places | The "no main window" rule for `cmdEndAll` lives in two places: an exception in the window-check helper and a guard in `quitAndEnd`. Either one alone makes the command do nothing, so the behaviour is correct. But the field's definition ("also when there is no main window") is then wrong for one of its own members. And the helper, which §4 means to be the single classification, would carry per-command special cases (New Tab's `RunOnMain(a.open)` from today's `run`, `commands.go:215-218`, plus this one). | Keep the guard in `quitAndEnd`, which the test pins, as the one owner. Have the helper run `cmdEndAll` like any other app-wide command. Word the field as "runs whatever window is in front; with no main window, its `Run` decides" (New Tab opens the window, Quit and End All Sessions does nothing). This is a recommendation, not required. |
| 2 | warn | §3 "One spelling of System": "Normalise `""` only in the v1 migration, so the menu check, the Settings selection and the file always agree" | pattern consistency: robustness of the pure parse function (§6 "Parsing") | A v2 file whose `appearance` is not `"system"`, `"light"` or `"dark"` (a hand edit to `""` or `"auto"`) parses without error. Its value then reaches `prefs`. `applyAppearance` falls through to System (`settings.go:45-52`), so no View ▸ Appearance radio item is checked and Settings selects nothing. The brief restores a broken `profiles.default` (§1 "Default cannot be removed") but says nothing about a broken `appearance`, so the stated goal of agreement does not hold for this case. Nothing crashes and no data is lost, so this is advisory. | Say in §1 what the parse function does with an unknown `appearance` in v2. Either map it to `"system"` (the way Default is restored, and the file is written on the next change), or treat the file as invalid and read-only (§6 "Write gate"). The first matches "Default cannot be removed". |
| 3 | warn | §3 "Where the syncing happens": the writers call `MenuItem.SetChecked` on the three radio items "which `App` keeps when `menu()` builds them"; §6 "Write gate" and "Parsing" tests change a setting | test seams (contract for T and F) | A test that calls `setAppearance` without first calling `a.menu()` would panic, not hang. `SetChecked` dereferences its receiver (`it.update`, then `it.Type`, mygo `menu.go:219-226`), and the kept items are nil until `menu()` runs. The criterion-8 write-gate test ("load such a file, then change a setting") and any `settingsView` test that clicks an Appearance choice can hit this. §6 enumerates the headless hazards carefully but leaves this one out. | Add one line to §3 or §6: the writers skip the menu sync when `menu()` has not built the items, so tests may call them without a menu. Otherwise, every test that calls a writer builds `a.menu()` first. |

## Notes for O

PASS: nothing needs amending before T. Two optional notes, not required:

- Warns 1–3 can be folded into the brief before T starts. If they are not, F and T should treat them as guidance. Warn 3 is the one most likely to cost a red-green cycle.
- Editorial: the brief has two `### 8.` headings ("Verified facts about the pinned mygo" at line 305 and "Verified source excerpts" at line 334). Renumbering the second to 9 would make the "§8" cross-references unambiguous.

Carried from rounds 2–4, still unchecked and left to F and T because it is runtime behaviour: whether calling `mygo.NewWindow` (or `a.open`) inside another window's frame is safe. This happens with the palette's "Settings…", Windows' `shortcuts()` loop through `runPosted`, and Ctrl+Shift+T in Settings on Windows with `a.win == nil`. The palette already runs `cmd.Run` inside the main frame, so there is a precedent.

## Patterns referenced

- `commands.go`: the `command` table (`:28-59`), `menu()` / `item()` / Appearance radio items (`:126-202`), `App.do` / `App.run` (`:204-226`), `quitAndEnd` (`:302-314`), `shortcuts()` (`:321-344`), `paletteItems` (`:354-384`)
- `main.go`: `main()` handlers (`:41-58`), `App.open` and the main window's `OnClosed` (`:65-131`), `saveNow` / `reset()` (`:133-159`)
- `settings.go` (all, `:1-74`); `app_test.go`: `newTestApp` (`:45-72`), `TestPaletteScrolls` (`:301-337`)
- mygo `v0.3.5-0.20261008155018-bad73d119dc8`: `content.go:145-171` (`Invalidate`; `Update` runs `fn` from the main loop's queue, not inside a frame, so it also runs for a minimised window), `loop.go:19-100` (`post`, `RunOnMain`, `onMain`), `menu.go:219-226` (`SetChecked`), `:300-335` (`prepareItem`), `:507-508` (`RoleClose`), `:522-525` (`macOnlyRoles`: `RoleAppMenu` is hidden off darwin; `RoleWindowMenu` is not, so Linux's Window ▸ Close Window exists), `:615-685` (`roleSubmenu`), `plugins/terminal/terminal.go:471-477` (`SetFont`), `:555-601` (`Close` calls `redraw` too, so calling `SetFont` on a closed pane's terminal with `a.win == nil` repeats a call `OnClosed` already makes)
