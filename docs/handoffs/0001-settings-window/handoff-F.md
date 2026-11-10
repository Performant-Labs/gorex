# Handoff-F: Phase 6 - Settings window, General and Default's text size, settings v2

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Issue:** #1 (delivery 1)

## What was done

- `settings.go`: settings.json v2 (`settings` with `Version`, `Appearance` and `Profiles map[string]*profile`; `profile` with `Name` and `FontSize *float32`). Also: the pure `parseSettings(bytes) (settings, v1, error)`, `complete()` (restores Default, clamps to 8–32, normalises the appearance), and the v1 migration (backup to `settings.v1.json`, then the v2 rewrite in the same `loadSettings`). The write gate is `readOnly`, which holds the reason. The theme seam `setThemeSource`. The writers `setAppearance` and `setFontSize` now sync View ▸ Appearance and the Settings window through `settingsChanged()`.
- `preferences.go` (new): the seams `newSettingsWindow` (production value `settingsWindow`, set in `init`), `focusWindow` and `closeWindow`. `App.openSettings`. `App.settingsView`, which has the sidebar (General, PROFILES, Default built-in), the General page (Appearance System/Light/Dark as `ui.RadioBase` swatches), the Default page (− slider + and the "11.6 pt" value, then the live preview) and the footer (saving, or "not saved: <reason>"). It also handles the Windows-only keys (Esc, Ctrl+W, Ctrl+Shift+W, plus a loop over the app-wide commands' `Other` shortcuts). There is one preview terminal per `App`.
- `commands.go`: a new `AppWide` field on `command`, set on New Tab, Bigger/Smaller/Actual Size, the 3 Appearance commands, Quit and End All Sessions, and the new `cmdSettings` (Title "Settings…", `CmdOrCtrl+,`, "⌘,", `Other` Ctrl+,, in `paletteCommands`). `cmdSystem` and the menu's System item now use `"system"`. `ui.KeyComma: ","` is in `shortcut.String()`. The menu now does three things:
  - On darwin the app menu is written out in full, with Settings… after About. Off darwin Settings… goes under Edit.
  - The three Appearance radio items are kept on `App`.
  - Show Tab 1–9 are now built with `item()`.
  New `App.click(cmd, win)`: Close Pane from Settings closes Settings, and otherwise it checks `runsIn`. New `App.runsIn(cmd, win)` is the one window check. `App.run` changes: with no main window, an app-wide command runs directly and New Tab opens the window as before; with one, New Tab first calls `focusWindow(a.win)` and everything goes through `a.do`. `quitAndEnd` returns at once when `a.win == nil`.
- `main.go`: `OnActivate` reopens the main window whenever `a.win == nil`; the `hasVisibleWindows` condition is dropped.
- `state.go`: these `App` fields sit outside `reset()`: `settingsWin`, `settingsPage`, `preview` and `appearanceItems`.
- `README.md`: the feature line, a `⌘, | Ctrl+, | Settings` row, a sentence naming Ctrl+, together with Ctrl+= Ctrl+- Ctrl+0 as the exceptions to the Shift rule, and a Layout row for `preferences.go` (the `settings.go` row now mentions the migration).

## Design decisions

- **The file is named `preferences.go`.** It had to be a new one-word lowercase name other than `settings.go` (§5, `TestReadmeSettings`). `prefs.go` would be confused with the `prefs` variable in `settings.go`. Compounds like `settingsview.go` break the repository's single-word style.
- **`newSettingsWindow` is assigned in `init`.** A package-level initializer formed an initialization cycle: `cmdSettings` → `openSettings` → `newSettingsWindow` → `settingsView` → `commands` → `cmdSettings`. So the var is declared without a value and `init` sets it to the named func `settingsWindow`. Its production value is unchanged in effect, and tests replace it as before.
- **The `cmdEndAll` guard lives only in `quitAndEnd`** (A round-5 warn 1). `runsIn` treats `cmdEndAll` like any other app-wide command, and its `Run` decides. The field comment says so ("with no main window, its Run decides").
- **An unknown `appearance` maps to `"system"`** (A warn 2). `complete()` maps any value other than light or dark to `"system"`, in v1 and v2 alike. That matches "Default cannot be removed": the file is corrected on the next save, and the menu and Settings always show a choice. Today the same values already fell through to System at runtime.
- **The writers need no nil guard** (A warn 3). `settingsChanged` ranges over the `appearanceItems` map, which is nil until `menu()` runs, so there is no panic without a menu.
- **A newer-version file is read as well as possible** ("or what it could read", §6). A version-3 file is parsed as v2, completed, and its run is read-only, so its dark appearance and size still apply. An invalid file gives the defaults.
- **v2 `fontSize` is clamped when present.** A v2 `0` becomes 8. Only v1's `0` or less counts as missing (11.6), because only v1 cannot tell 0 from absent (§1).
- **Settings' − and + versus ⌘− and ⌘+.** In Settings, − is `ceil(size)-1` and + is `floor(size)+1`, so 11.6 goes to 11 or 12 (§2). ⌘+ and ⌘− keep today's ±1 (11.6 → 12.6), as criterion 4 says "as today". The slider is `ui.Slider(...).Step(1)`. mygo snaps it to `lo` plus whole steps, so it lands on whole points while the thumb sits at the exact value.
- **Show Tab items get a per-item `command{}`** passed to `item()`, so they go through the same `click` → `runsIn` → `run`. That is one window check and no second code path (§4).
- **The preview terminal** is made with `Transparent: true` inside a box painted with `lightTerm.Background` or `darkTerm.Background`, which reuses the terminal colours and adds no copy. It gets `Resize(60, 4)` before `Feed` and room for four lines, so the first line is not scrolled away. I found that from a headless render.
- **The selected-swatch ring** uses `c.Theme().Accent` (mygo's theme), so no new colour is added to `colors`.

## Reuse / extend-vs-new

- EXTENDED `App.setAppearance` and `App.setFontSize`. They are the only writers. Settings, the menu, the palette and ⌘+/⌘−/⌘0 all call them, and Settings never writes `prefs`, `termFont` or `saveSettings`.
- EXTENDED the `command` table with `cmdSettings` and the `AppWide` field. Windows' main-window Ctrl+, comes from the existing loop in `shortcuts()`. Settings' Windows loop iterates the same `commands` table.
- EXTENDED `App.run` (the New Tab–only exception became `AppWide`) and `item()` (the window check).
- EXTENDED the `OnActivate` handler (one condition dropped).
- FOLLOWED `App.open` for the window: `mygo.NewWindow(WindowOptions{StateKey, Content: ui.View(...)})`, then `win.OnClosed`.
- REUSED `attach`'s display-only terminal (`terminal.New` with `nopConn{}`, `termFont`, `lightTerm`/`darkTerm`, then `Feed`), `colors`/`colorsOf` and `iconButton`.
- NEW, as the brief justifies in §5: the Settings window and its view, in `preferences.go`.

## Architecture notes for A

- New package-level seams: `newSettingsWindow`, `focusWindow`, `closeWindow` and `setThemeSource`.
- New `command.AppWide` field. New `App.click` and `App.runsIn`.
- `App.run` changed semantics: with no main window, app-wide commands now run directly.
- The settings.json schema changed to v2, with a migration. The v1 backup is `settings.v1.json`.
- No new dependencies. `preferences.go` imports `plugins/terminal`, as `state.go` and `view.go` already do.
- `debug.go:70` (dev hook) calls `a.run(cmd)`, so it inherits the new `run`: New Tab now also focuses the main window there. That is harmless.

## Deviations from spec / wireframe

- The wireframe's "Later: notifications, confirm before quitting, …" line and the 8/32 tick labels under the slider are not shipped. The first is an annotation of future work. The labels would sit between − and + in the slider's column, and the brief does not ask for them. The "+ New Profile…" footer is left out, as the brief says.
- The Default page has the title "Default" and no subtitle. Wireframe 2's "Used by N panes" line is delivery 2.
- Otherwise none.

## Tier 1 self-check (incl. tests now GREEN)

    $ gofmt -l .                       # (no output)
    $ go vet ./... && GOOS=linux go vet ./... && GOOS=windows go vet ./...   # VET-OK
    $ GOOS=windows go test -c -o /dev/null . && GOOS=linux go test -c -o /dev/null .   # CROSS-OK
    $ go build ./...                   # ok
    $ go test -count=1 -race ./...
    ok  	gorex	5.340s
    ok  	gorex/internal/rex	1.387s
    ?   	gorex/tools/mkicon	[no test files]
    $ go test -count=1 -race -v .  (summary)
    --- PASS: TestPanesAndTabs / TestRestore / TestLabels / TestCloseButtons / TestPaletteScrolls
    --- PASS: TestSettingsFirstLaunch
    --- PASS: TestSettingsMenuItem
    --- PASS: TestSettingsOpensOnce
    --- PASS: TestCommandsFromSettings
    --- PASS: TestMenuClicksFromSettings
    --- PASS: TestEndAllNeedsMainWindow
    --- PASS: TestSettingsAppearanceInStep
    --- PASS: TestSettingsTextSize
    --- PASS: TestSettingsMigratesV1 (5 subtests)
    --- PASS: TestSettingsMigrationOverwritesOldBackup
    --- PASS: TestSettingsV2Load (4 subtests)
    --- PASS: TestSettingsSurviveRestart
    --- PASS: TestSettingsReadOnly (3 subtests)
    --- SKIP: TestSettingsKeysOnWindows   (Windows only; compiles for windows)
    --- PASS: TestReadmeSettings
    PASS
    ok  	gorex	5.378s

All 14 darwin-runnable new tests are GREEN, and the 5 existing tests still pass. I also rendered the General and Default pages headlessly in light and dark, with `ui.Tester.Image()` in a scratch copy of the tree. Nothing was committed and nothing opened on screen. That was to check the layout, and it caught the preview's first line scrolling away, which is now fixed.

## Evidence appendix

`docs/handoffs/0001-settings-window/evidence.md`

## Tests that look wrong (for T)

None.

## Known issues

- `TestSettingsKeysOnWindows` has not run: there is no Windows host. It compiles and vets for `GOOS=windows` only.
- These were not run live: closing Settings from inside its own frame on Windows (`closeWindow` in `settingsKeys`), and `NewWindow`/`a.open` called inside another window's frame (the palette's "Settings…", Ctrl+Shift+T in Settings on Windows with no main window). Each has a precedent in the code (evidence entries 4 and 5), but they are runtime behaviour for U.
- `saveSettings` still ignores a failed write of `settings.json` (disk full, say), as it does today, so the footer would still say "save as you make them". The brief names only the three read-only reasons, and this is unchanged behaviour, not something new.
- Not reachable headless, as T noted: the Dock reopen, ⌘1–9 doing nothing from Settings, and the Appearance highlight and preview look (U/S).

## Files changed

- `settings.go`
- `preferences.go` (new)
- `commands.go`
- `main.go`
- `state.go`
- `README.md`
