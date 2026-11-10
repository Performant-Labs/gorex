# Handoff-A: Phase 4 - Settings window, General and Default's text size, settings v2  (up-front plan review, round 3)

This is Phase 4 in the pipeline table. `architecture-reviewer.md` calls the same up-front A review "Phase 3". This round re-reviews the brief as amended in `d7e206b`, `48b00fc`, `b789aa7`, `8f49a29` and `3149f5d`. It is the arch gate's third review, so only a correctness finding can block.

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Brief reviewed:** `docs/handoffs/0001-settings-window/brief.md` (at `3149f5d`)   **Reuse map:** brief §5 "Reuse & extend-vs-new" (there is still no `survey.md`)   **Wireframe:** `docs/wireframes/settings-profiles/1-general.png`; for the Default page, only the text-size row and preview of `2-profile-overrides.png` (unchanged since round 2)
**Verdict:** BLOCK

## Summary

BLOCK, with one correctness finding. The round-2 block is resolved, and so are its four warns. §6 now puts `NewWindow`, `Focus`, `Close` and `Theme.SetSource` behind seams, and §3 no longer guards `SetChecked`. The new block comes from round 2's warn 3, which classified commands as app-wide. §4 says "App-wide commands run directly, not through `a.do`". In this codebase `a.do` is what redraws the main window after a menu command (`commands.go:204`). Neither mygo's menu click nor a direct state change asks for a frame. Built as written, ⌘T from the main window's menu (Ctrl+Shift+T on Linux) adds the tab, but the window does not draw it and focus does not move until the next 500 ms poll. Keys typed in that gap go to the old pane. The fix is one sentence in §4.

## Round-2 findings: status

| R2 # | Status | Where in the brief |
|---|---|---|
| 1 block: test seams miss the calls that block headless | resolved | §6 adds `newSettingsWindow`, `focusWindow`, `closeWindow` and the Theme seam in `applyAppearance`. §3 drops the `SetChecked` guard and asserts `IsChecked()` on the items `menu()` keeps. §8 adds the verified facts and the committed probe (`evidence/mygoprobe/output.txt`). |
| 2 warn: Show Tab items skip the window check | resolved | §4: one helper, used by `item()` and the Show Tab 1–9 items. |
| 3 warn: which commands run from Settings | adopted, with the defect below | §4 adds an app-wide field next to `Hidden`, read by the helper and by `App.run`. The rule for how app-wide commands run is block finding 1. |
| 4 warn: app menu bar off darwin | resolved | §4 "The menu bar off darwin". |
| 5 warn: `settingsView` vs the main view's queue | resolved | §4 "Settings' own queue". |

Also re-checked this round: closing a window from inside its own frame (the Windows Esc / Ctrl+W / Ctrl+Shift+W path) has a precedent in `windowControls` (`view.go:124`, `win.Close()` in the frame). `menuItemClicked` toggles a radio item, then calls `Click` on the main thread (mygo `menu.go:416-431`, `:85-88`). So keeping the Appearance items and calling `SetChecked` again from `setAppearance` agrees with mygo. The nested `evidence/mygoprobe/go.mod` keeps the probe out of `./...`, so criterion 12's `go vet` / `go test` do not pick it up.

## Findings

| # | Severity | Plan element | Drift dimension | Finding | Suggested fix |
|---|---|---|---|---|---|
| 1 | block | §4 "App-wide commands run directly, not through `a.do`, which drops work when `a.win == nil`" (with New Tab app-wide, criterion 10) | pattern consistency / cross-cutting: menu dispatch and redraw | `a.do` is how a menu command changes the main window and gets it redrawn. Its comment says it "changes the state on the main thread, from a menu, and draws" (`commands.go:204-210`), through `a.win.Update`. mygo requests a content frame only from `Update` and `Invalidate` (`content.go:145-181`, the only callers of `contentChanged`). `menuItemClicked` calls `Click` and nothing else (`menu.go:416-431`), and `newTab` only marks the layout for saving (`a.changed()` sets `saveDue`, `state.go:696`). So if New Tab runs directly when the main window is open, `newTab` changes `a.tabs` and `a.focusReq`, but no frame shows the new tab or moves focus to it. The next frame comes from `poll`'s 500 ms `win.Update` (`state.go:808-834`). Until then the old pane keeps the keyboard. ⌘T from the menu is the main way to open a tab on macOS, and Ctrl+Shift+T is the same on Linux. Today it goes through `a.run` → `a.do` and draws at once. §3 also says "`a.do` redraws only the main window, so it is not enough on its own", which assumes `a.do` is still in the path, so §3 and §4 disagree. Headless tests pump frames by hand, so neither T nor the Phase 7 gate would catch this. | App-wide commands go through `a.do` when the main window is open, and run directly only when `a.win == nil`. Exact wording in Notes for O. |
| 2 | warn | §4 "One helper decides whether a command runs for the window a click came from" | contract shape | mygo documents that a click's window "may be nil" for menus other than a window's own menu bar (`menu.go:85-88`). On macOS that is `FocusedWindow()` (`app.go:591`), which is nil when no GoRex window is key, for example with the main window minimized. The brief does not say what the helper does with a nil window, so T and F could each choose differently: run main-window commands on `a.win` as today, or drop them. | Say it in §4: with a nil window, app-wide commands run and main-window commands do nothing (no hidden pane is ever closed). Have a test pin it. |
| 3 | warn | §2 "Windows": Settings' own key handling handles exactly Esc, Ctrl+W, Ctrl+Shift+W and Ctrl+, | pattern consistency (the new app-wide contract) | §4 defines app-wide as "runs whatever window is in front", but on Windows Settings drops Ctrl+= / Ctrl+- / Ctrl+0 and Ctrl+Shift+T, which work from Settings on macOS and Linux. The brief chooses this on purpose (criterion 10), so it is not drift in this delivery. It does leave the classification true on only two of the three platforms, and it adds four hand-written cases where `shortcuts()` already loops over the `command` table (`commands.go:334-338`). | Follow-up: on Windows, have Settings' key handling loop over the app-wide commands' `Other` shortcuts from the same table, plus its close keys, so the field means the same thing on every platform. |
| 4 | warn | §5 REUSE "Hold one terminal across frames" (live preview) | cross-cutting: lifecycle | A `terminal.New` with a `Conn` starts `go t.read()` at once (mygo `plugins/terminal/terminal.go:214-217`, `:250-253`). `nopConn.Read` blocks forever (`select {}`, `state.go:255`), and `Close` only calls the Conn's `Close` (`terminal.go:568-571`, `closeSrc` at `:603-605`), which returns nil, so that goroutine never ends. If the preview terminal is made each time Settings opens, every open leaks one. | Make the preview terminal once per `App`, on the first open, and keep it with the other Settings state outside `reset()`. Do not make one per window. |

## Notes for O

Amend `brief.md`, then start a fresh run (not `resumeFromRunId`). To resolve finding 1, change §4 (the sentence at `brief.md:163-164`) from "App-wide commands run directly, not through `a.do`, which drops work when `a.win == nil`." to:

> App-wide commands run through `a.do` when the main window is open, so the main window redraws after them as it does today (`a.do` "changes the state on the main thread, from a menu, and draws", `commands.go:204`; mygo requests a frame only from `Update` and `Invalidate`, and a menu click calls neither). They run directly only when `a.win == nil`, where `a.do` would drop them. New Tab with `a.win == nil` opens the main window as today (`a.open` makes its first tab). Otherwise it calls `focusWindow(a.win)` and then runs `newTab` through `a.do`. The writers still `Invalidate()` an open Settings window (§3), because `a.do` redraws only the main window. The View ▸ Appearance radio items may keep calling `setAppearance` directly, as today (`commands.go:147-150`): a theme change redraws every window.

Change the New Tab clause in §4's app-wide list and criterion 10 to match ("brings the main window to the front, then opens the tab in it through `a.do`"). If the run is still in the brief gate, add to §8: `content.go:145-181` (`Invalidate` and `Update` are the only callers of `contentChanged`) and `menu.go:416-431` (`menuItemClicked` toggles a checkbox or radio item, runs its role, then calls `Click`, and requests no frame).

Recommended in the same edit, not required to pass: warn 2 (a nil click window) and warn 4 (one preview terminal per `App`). Warn 3 is follow-up work.

## Patterns referenced

- `commands.go`: `App.do` (`:204-210`), `App.run` (`:212-226`), `menu()` / `item()`, the Show Tab and Appearance items (`:126-202`), `shortcuts()` (`:321-344`)
- `state.go`: `newTab` / `changed` (`:266-278`, `:696`), `poll` (`:808-834`), `attach` / `nopConn` (`:194-257`), `post` / `later` / `runPosted` (`:113-136`)
- `view.go`: `a.view` (`:20-50`), and `windowControls` closing its window from inside a frame (`:88-125`)
- `main.go`: `App.open`, `OnActivate`, `reset()`
- mygo `v0.3.5-0.20261008155018-bad73d119dc8`: `content.go` (`Invalidate`, `Update`, `contentChanged`), `menu.go` (`MenuItem.Click` docs, `menuItemClicked`, `roleSubmenu`), `app.go` (`MenuItemClicked`, `SetMenu`), `window.go` (`NewWindow`, `Close`, `FocusedWindow`), `loop.go` (`RunOnMain`, `onMain`, `post`), `plugins/terminal/terminal.go` (`New`, `run`, `Close`, `SetFont`, `redraw`)
