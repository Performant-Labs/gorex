# Handoff-A: Phase 4 - Settings window, General and Default's text size, settings v2  (up-front plan review, round 4)

This is Phase 4 in the pipeline table. `architecture-reviewer.md` calls the same up-front A review "Phase 3". This round re-reviews the brief as amended in `c35cf0f`. It is the arch gate's fourth review, so only a correctness finding can block.

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Brief reviewed:** `docs/handoffs/0001-settings-window/brief.md` (at `c35cf0f`)   **Reuse map:** brief §5 "Reuse & extend-vs-new" (there is still no `survey.md`)   **Wireframe:** `docs/wireframes/settings-profiles/1-general.png`; for the Default page, only the text-size row and preview of `2-profile-overrides.png` (unchanged since round 2)
**Verdict:** BLOCK

## Summary

BLOCK, with one correctness finding. The round-3 block is resolved, and all three round-3 warns are resolved or adopted. The new block is in the app-wide classification in §4. The brief makes Quit and End All Sessions (`cmdEndAll`) app-wide: it "runs whatever window is in front, and also when there is no main window". With no main window, its `Run` cannot end the sessions. `quitAndEnd` uses `a.client`, and the main window's `OnClosed` closes that connection. Built as written, choosing the command with the main window closed would quit GoRex and leave every session running, with no message. Today the command does nothing in that state.

I should have raised this in round 3. The classification was already in the brief at `3149f5d`, and I did not check each app-wide command's no-main-window path then. This round checks all of them. The text-size, Appearance, Settings… and New Tab paths are sound, so this is the only instance. The fix is one paragraph in §4.

## Round-3 findings: status

| R3 # | Status | Where in the brief |
|---|---|---|
| 1 block: app-wide commands skip `a.do`, so the main window is not redrawn | resolved | §4 "How app-wide commands run": through `a.do` when `a.win != nil`, directly only when `a.win == nil`. New Tab calls `focusWindow(a.win)` and then runs `newTab` through `a.do`. §3 and §4 now agree. §8 adds `content.go:145-181` and `menu.go:416-431`. Criterion 10 matches. |
| 2 warn: what the helper does with a nil click window | resolved | §4 "A click with no window": app-wide commands run, and commands that need the main window do nothing. A test pins both (see warn 3 below). |
| 3 warn: on Windows, Settings handles only four keys | adopted in this delivery | §2 "Windows": Settings' own key handling loops over the app-wide commands' `Other` shortcuts from the `command` table. Criterion 10 now covers Windows. Those commands are exactly `cmdSettings`, `cmdNewTab`, `cmdBigger`, `cmdSmaller` and `cmdActualSize` (`commands.go:29`, `:52-54`). The Appearance commands and `cmdEndAll` have no `Other`. One overlap is noted as warn 2. |
| 4 warn: one preview terminal per `App` | resolved | §5 REUSE: one preview terminal per `App`, made the first time Settings opens and kept outside `reset()`. Also checked this round: reusing it in a later Settings window is sound. `viewOf` rebinds the terminal's redraw to the services of whichever window draws it (mygo `plugins/terminal/view.go:36-48`), so `SetFont` redraws the current Settings window. |

## Findings

| # | Severity | Plan element | Drift dimension | Finding | Suggested fix |
|---|---|---|---|---|---|
| 1 | block | §4 "One window check, one classification": "Quit and End All Sessions" is app-wide, so it "runs … also when there is no main window", and §4 "How app-wide commands run" runs it "directly only when `a.win == nil`" | cross-cutting: lifecycle (the command depends on state the main window owns) | `cmdEndAll.Run` is `a.quitAndEnd()` (`commands.go:58`). It ends the sessions by sending `Shutdown()` on `a.client`, then calls `mygo.App.Quit()`, and it ignores `Shutdown`'s error (`commands.go:302-313`). `a.client` belongs to the main window's lifecycle. `App.open` connects it (`main.go:66-73`), and the main window's `OnClosed` closes it (`main.go:128`), before `a.win = nil`. On a closed connection `call` returns an error at once: `c.err` is set by the reader, or the write fails (`internal/rex/client.go:115-131`). So with `a.win == nil` the session server gets no shutdown request, and the app quits anyway. GoRex quits and every session keeps running, the opposite of the command's name, and nothing tells the user. The README lists ⌥⌘Q as "Quit and end all sessions" (`README.md:55`), and the error banner tells users to use it to restart a stale server (`main.go:89`). This is new reachable behaviour. Today `run` drops every command except New Tab when `a.win == nil` (`commands.go:214-220`), so the command does nothing. The plan makes this path reachable on macOS (⌥⌘Q with the main window closed, with or without Settings open) and off darwin (Shell ▸ Quit and End All Sessions in Settings' menu bar, §4 "The menu bar off darwin"). S checks against the brief and the tests run with a connected client, so no later phase would catch this. | Keep the command app-wide while the main window is open, through `a.do` like the others, including from Settings. When `a.win == nil` it does nothing, as today. Exact wording is in Notes for O. |
| 2 | warn | §2 "Windows": a hand-written Ctrl+, case, then a loop over the app-wide commands' `Other` shortcuts | pattern consistency | `cmdSettings` is app-wide and its `Other` is Ctrl+, (§5), so the new loop already handles Ctrl+, in Settings. If F also keeps the separate Ctrl+, case that §2 describes first, one key press runs `openSettings()` twice. That is harmless, because it only focuses the window twice, but §5 forbids the same thing for the main window ("not a hand-written case"). | Say in §2 that the loop is Settings' only Ctrl+, path, with no separate case. |
| 3 | warn | §4 "A click with no window … A test pins both" | test seams (contract for T) | Headless tests cannot run two app-wide commands with `a.win == nil`. New Tab runs `mygo.RunOnMain(a.open)` (`commands.go:216-217`). `RunOnMain` is `onMain`, which waits forever off the main thread (§8), and `a.open` is not behind a §6 seam. Quit and End All Sessions, without finding 1's guard, would shut down the test's real server (`newTestApp` connects one, `app_test.go:52-69`) and call `mygo.App.Quit()`. | Name the command the nil-window test uses: Bigger Text, or an Appearance command through the §6 Theme seam. Add the finding 1 check, that `a.quitting` stays false with `a.win == nil`. |

## Notes for O

Amend `brief.md`, then start a fresh run, not `resumeFromRunId`. This is the second round in a row with a new blocking finding (round 3's was new too). `workflow-coding-pipeline.md` treats that as an escalation signal. The operator sees this note in the stop payload either way.

To resolve finding 1, change §4's app-wide list (`brief.md:163-164`, "… the three Appearance commands, Quit and End All Sessions.") to end "… the three Appearance commands, and Quit and End All Sessions while the main window is open (below)." Then add this bullet after §4 "How app-wide commands run":

> **Quit and End All Sessions needs the session server's connection.** `quitAndEnd` ends the sessions through `a.client` (`commands.go:302-313`). `App.open` connects it (`main.go:66-73`), and the main window's `OnClosed` closes it (`main.go:128`). With `a.win == nil` the connection is closed, `Shutdown()` fails at once (`internal/rex/client.go:115-131`), `quitAndEnd` ignores the error and quits, and every session would keep running. So with the main window open, Quit and End All Sessions is app-wide like the others: from Settings too, it runs through `a.do`. With `a.win == nil` it does nothing, as today (`commands.go:214-220`). `quitAndEnd` returns at once when `a.win == nil`. A test pins that with `a.win == nil`, running `cmdEndAll` leaves `a.quitting` false.

Make the same change in §4 "A click with no window" ("app-wide commands run (as above), except Quit and End All Sessions with `a.win == nil`"). Add to §8: `main.go:128` (the main window's `OnClosed` closes `a.client`) and `internal/rex/client.go:115-131` (`call` fails at once on a lost or closed connection).

I chose "does nothing, as today" over connecting a new client to shut the server down. That alternative is new behaviour outside this delivery's scope, and it would open a server connection outside the main window's lifecycle, which nothing else in GoRex does.

Recommended in the same edit, not required to pass: warn 2 (one Ctrl+, path in Settings on Windows) and warn 3 (name the nil-window test's command).

## Patterns referenced

- `commands.go`: the `command` table (`:28-59`), `App.do` / `App.run` (`:204-226`), `quitAndEnd` (`:302-314`), `shortcuts()` (`:321-344`), `menu()` / `item()` (`:126-202`)
- `main.go`: `App.open` (`:65-131`), the main window's `OnClosed` (`:117-130`), `saveNow` / `reset()` (`:133-159`), `OnActivate` (`:46-50`)
- `internal/rex/client.go`: `call` (`:115-149`), `Shutdown` / `Close` (`:182-184`)
- `state.go`: `post` / `later` / `runPosted` (`:113-136`); `settings.go`: `setFontSize` / `setAppearance` (`:55-71`)
- mygo `v0.3.5-0.20261008155018-bad73d119dc8`: `content.go:145-171` (`Invalidate`, `Update` posts and does not wait), `loop.go` (`post`, `RunOnMain`, `onMain`), `plugins/terminal/terminal.go:471-477` (`SetFont`), `:553-600` (`Close`), `plugins/terminal/view.go:36-48` (`viewOf` rebinds `draw` per window)
