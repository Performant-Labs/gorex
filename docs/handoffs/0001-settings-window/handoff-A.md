# Handoff-A: Phase 4 - Settings window, General and Default's text size, settings v2  (up-front plan review, round 2)

This is Phase 4 in the pipeline table; `architecture-reviewer.md` calls the same up-front A review "Phase 3". This is a re-review of the brief as amended in `bf8424e` and `a923963`.

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Brief reviewed:** `docs/handoffs/0001-settings-window/brief.md` (at `a923963`)   **Reuse map:** brief §5 "Reuse & extend-vs-new" (there is still no `survey.md`)   **Wireframe:** `docs/wireframes/settings-profiles/1-general.png`; for the Default page, only the text-size row and preview of `2-profile-overrides.png`
**Verdict:** BLOCK

## Summary

BLOCK, with one block finding. The amended brief fixes both round-1 blocks. §4 now sends commands to the window the menu click reports, and §5 is a real reuse map. I checked the new mygo claims and they hold. The new block is the test seams in §3 and §6, which are wrong for this mygo version. `mygo.Theme.SetSource` (called by `setAppearance` and `loadSettings`) and `Window.Focus()` / `Window.Close()` wait on the main thread, and under `go test` they block forever (shown with a probe). So the single-window test in §6, and every appearance test that criteria 3, 8 and 12 need, would hang. The guard §3 adds is on `SetChecked`, which is safe headless. The round-1 seam suggestion (warn 3) came from me and had the same gap; this finding corrects it.

## Round-1 findings: status

| R1 # | Status | Where in the brief |
|---|---|---|
| 1 block: dispatch with a second window | resolved | §4 and criteria 10–11. On macOS, mygo passes `FocusedWindow()` as a main-menu click's `win` (`app.go:591`). On Linux and Windows it passes the window whose menu bar was used (`menu.go:85-88`). |
| 2 block: no Reuse map | resolved | §5 has EXTEND, FOLLOW, REUSE and NEW entries, and the NEW entry is justified. |
| 3 warn: test seam | adopted but incomplete | Replaced by block finding 1 below. |
| 4, 5, 6, 7, 8, 9, 11 warns | adopted | §3, §3, §5, §6, §5, §5, §7 |
| 10 warn: `OnActivate` | adopted; confirmed correct | mygo runs `OnActivate` from `applicationShouldHandleReopen:hasVisibleWindows:` (`internal/darwin/app.go:305`). That fires on a Dock click or reopen, not on every activation, so reopening on `a.win == nil` alone does not bring the main window back on ⌘-Tab. |

Also checked: `ui.KeyComma` exists (`ui/keys.go:45`), and `,` is a valid accelerator key (`internal/accelerator`). The `roleSubmenu` and `prepareItem` excerpts in §8 match `menu.go`. Only `testhooks.go:58` mentions `performKeyEquivalent`.

## Findings

| # | Severity | Plan element | Drift dimension | Finding | Suggested fix |
|---|---|---|---|---|---|
| 1 | block | §6 "Opening Settings", §3 "If `SetChecked` cannot run under `go test` … guard it", criterion 12 | abstraction / testability seam (platform side effects inside the single writers) | mygo's `init` runs `runtime.LockOSThread()` (`mygo.go:72-75`), so test goroutines never run on the main thread. No event loop runs under `go test`, so `onMain` (`loop.go:87-100`) posts and waits forever. A probe built against the pinned mygo found: `mygo.Theme.SetSource` **blocks** (`modules.go:259-274` calls `onMain`), and `(&mygo.Window{}).Focus()` **blocks** (`window.go:513-519`, `:607`); `Close()` uses the same `onMain` path (`window.go:534`). `NewMenu`, `MenuItem.SetChecked` and `(&mygo.Window{}).Invalidate()` **return**, and after `SetChecked(true)` on one radio item, `IsChecked()` reports the new state at once. What this means for the plan: (a) The single-window test in §6 calls `a.settingsWin.Focus()` on its second open and hangs. (b) `setAppearance` is the only writer (§5) and calls `applyAppearance` → `Theme.SetSource`, so any headless test of appearance hangs. That covers the Settings radio (criterion 3) and menu ↔ Settings sync (criterion 12). (c) The write-gate test in §6 goes through `loadSettings`, which calls `applyAppearance`, so it hangs too. (d) A test of §4's Close-Pane-closes-Settings path hangs on `Close()`. The guard in §3 targets the one call that is safe. | Put every blocking mygo call on the Settings and appearance paths behind the seam pattern §6 already uses (package-level func variables, production value = the mygo call). Tests replace them. Drop the `SetChecked` guard, and let tests assert `IsChecked()` on the items `menu()` keeps. The exact wording to add is in Notes for O. |
| 2 | warn | §4 "Menu items built by `item()` check the window" | cross-cutting: command dispatch | The Show Tab 1–9 items (`commands.go:140-145`) are not built by `item()`. They call `a.do` directly. With Settings in front, ⌘1–⌘9 would switch tabs in the hidden main window, against §4's rule that "No pane or tab command acts on the main window from it". This is harmless (it closes nothing), but it is a second path that skips the check. | Do the window check in one helper that both `item()` and the Show Tab items use. Don't copy it into each closure. |
| 3 | warn | §4 which commands run from Settings | pattern consistency / contract shape | §4 covers Close Pane and "pane or tab commands". It doesn't say what ⌘+ / ⌘− / ⌘0, ⌘⇧P / ⌘P (palette), ⌘⌥Q (Quit and End All Sessions) and Settings itself do when Settings is in front. It also doesn't say what they do once the main window is closed, when `run` drops everything except New Tab and Settings (`commands.go:214-220`). T and F could each pick a different answer. | Classify commands once, on the `command` table. The table already has a per-command flag, `Hidden`. Add a field for app-wide commands (Settings, the three text-size commands, the appearance commands, Quit and End All Sessions) that run whichever window is in front and with no main window. Both `item()`'s window check and `run`'s exception read that field, replacing the identity checks. |
| 4 | warn | Settings window off darwin | cross-cutting: menus | `Application.SetMenu` is also "the menu bar of windows without their own menu on Linux and Windows" (mygo `app.go:319-323`). So the Settings window gets Shell, Edit, View, Tabs and Window there, and the brief doesn't say whether that is intended. Criterion 10's Linux Ctrl+W depends on it: it works only through Window ▸ Close Window (`RoleClose`, whose `performRole` closes the clicked window, `menu.go:507-508`). | Say in §2 that the Settings window keeps the app menu bar on Linux and Windows. §4's check keeps those items safe, and Linux Ctrl+W comes from `RoleClose`. If it should have no menu bar instead, Linux Ctrl+W needs its own handler in `settingsView`, as Windows has. |
| 5 | warn | `settingsView` vs the App's posting queue | cross-cutting: frame lifecycle | `a.post` / `a.later` / `laterService` feed a queue that only the main window's `a.view` drains (`runPosted`, `view.go:22`). `a.view` also sets `a.services` (`view.go:21`). If the Settings view posted into that queue, its work would run only when the main window next draws, and never once the main window is closed. | State that `settingsView` calls `setAppearance` / `setFontSize` directly in its own frame. It doesn't use `a.later` / `laterService` and doesn't assign `a.services`. Its Windows Esc / Ctrl+W close the window through the seam from finding 1. |

## Notes for O

Amend `brief.md`, then start a fresh run (not `resumeFromRunId`). To resolve finding 1:

1. **§6 "Opening Settings":** replace the single creation variable with a seam that covers every native call the Settings window makes, for example three package-level func variables:
   - `newSettingsWindow func(a *App) *mygo.Window`. In production it calls `mygo.NewWindow(mygo.WindowOptions{StateKey: "settings", Content: ui.View(a.settingsView), …})` and then registers `win.OnClosed(…)` to clear `a.settingsWin`. `OnClosed` is a method on `*mygo.Window`, not a `WindowOptions` field (`window.go:1338`); `App.open` registers it the same way.
   - `focusWindow = (*mygo.Window).Focus`
   - `closeWindow = (*mygo.Window).Close`

   Tests return `&mygo.Window{}` from the creator (its `Invalidate()` is safe headless, which is what §3 needs) and replace focus and close with recorders.
2. **§6, new bullet "Theme":** `applyAppearance` (`settings.go`) calls a package-level func variable whose production value is `mygo.Theme.SetSource`, not `mygo.Theme.SetSource` directly. Tests replace it, so `setAppearance`, `loadSettings` and the write-gate test run headless and can assert which theme was applied.
3. **§3:** delete "If `SetChecked` cannot run under `go test` … guard it". Replace it with: `menu()` and `SetChecked` are safe headless. `SetChecked` changes `Checked` immediately and only queues the native update. So tests build `a.menu()`, keep the three Appearance items, and assert `IsChecked()`.
4. **§8:** add these verified facts as excerpts: `mygo.go:72-75` (`LockOSThread`), `loop.go:87-100` (`onMain` waits), `modules.go:259-274` (`SetSource` calls `onMain`), `window.go:513-519` / `:534` / `:607` (`Focus` / `Close` call `onMain`), `content.go:145-148` (`Invalidate` returns early without content), and `menu.go:243-264` (`update` mutates, then `postMain`).

Recommended in the same edit, but not required to pass: warns 2–5 (one window-check helper, an app-wide field on `command`, the menu bar off darwin, and `settingsView` not using the main view's queue).

## Patterns referenced

- `settings.go`: `loadSettings`, `applyAppearance`, `setAppearance`, `setFontSize`
- `commands.go`: `command` table and `Hidden`, `init()`, `App.menu()` / `item()` and the Show Tab items, `App.run` / `App.do`, `shortcuts()`
- `main.go`: `App.open` (`win.OnClosed`), `OnActivate`, `reset()`; `view.go:20-24` and `state.go:104-136` (`services`, `post` / `later` / `runPosted`)
- `app_test.go`: `newTestApp` (`t.Setenv("GOREX_DIR")`, `ui.NewTester`, `a.win == nil`)
- mygo `v0.3.5-0.20261008155018-bad73d119dc8`: `mygo.go` (`init`, `backend`), `loop.go`, `modules.go` (`ThemeModule.SetSource`), `window.go` (`NewWindow`, `do`, `Close`, `Focus`, `OnClosed`, `FocusedWindow`), `content.go` (`Invalidate`, `Update`), `menu.go` (`SetChecked` / `update`, `prepareItem`, `roleSubmenu`, `performRole`, `menuItemClicked`), `app.go` (`SetMenu`, `MenuItemClicked`), `internal/darwin/app.go` (`applicationShouldHandleReopen`, `Signal`)
