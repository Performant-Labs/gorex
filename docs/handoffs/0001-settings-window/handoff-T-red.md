# Handoff-T-red: Phase 5 - Settings window, General and Default's text size, settings v2

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Issue:** #1 (delivery 1)
**Brief / wireframe reviewed:** `docs/handoffs/0001-settings-window/brief.md` (at `6238bf7`), `docs/wireframes/settings-profiles/1-general.png`, `2-profile-overrides.png` (text-size row and preview only), `handoff-A.md` (round 5), `decisions.md`

## A precondition

Confirmed: A returned PASS on the plan (round 5, `785327d`), with 3 advisory warns. How the tests treat them:
- Warn 1 (the no-main-window rule for `cmdEndAll` lives in two places): the tests pin only the behaviour (`TestEndAllNeedsMainWindow`) and do **not** assert what the helper returns for `cmdEndAll` with `a.win == nil`. F may keep the guard in `quitAndEnd` only, as A recommends, or in both places, as the brief says.
- Warn 2 (an unknown `appearance` in a v2 file): the brief does not decide it, so no test pins it.
- Warn 3 (`SetChecked` on nil kept items): every test builds `a.menu()` before it calls a writer, as brief §3 says. F does not have to add a nil guard (it may).

## The contract: names F must provide

The brief names `newSettingsWindow`, `focusWindow`, `closeWindow`, `App.openSettings`, `App.settingsWin`, `App.settingsView` and `cmdSettings`. The tests use those names and signatures as given in brief §6. The brief leaves the following unnamed, so the tests fix them:

| Name | Shape | Brief |
|---|---|---|
| `setThemeSource` | package `var setThemeSource func(mygo.ThemeSource)`, production value `mygo.Theme.SetSource`; `applyAppearance` calls it | §6 "Theme" |
| `(*App).runsIn` | `func (a *App) runsIn(cmd *command, win *mygo.Window) bool`: whether `cmd` runs for a menu click from `win` | §4 "One helper decides" |
| Settings' − and + | found by the labels `"Smaller Text"` and `"Bigger Text"` (their text, or `.Label(...)` as `view.go` labels its icon buttons) | §2 "Default page" |
| Appearance choices, sidebar entry | clicking the element showing or labelled `"System"` / `"Light"` / `"Dark"` chooses it; `"Default"` in the sidebar opens Default's page | §2 |
| `settingsView` driven directly | `ui.NewTester(a.settingsView, …)` works without `openSettings` having run: the preview terminal must be made lazily if it is missing | §6 "The page" |

Other notes for F:
- The test creator stands in for `newSettingsWindow`, sets `a.settingsWin` itself and returns the window, so `openSettings` may or may not assign the field.
- `loadSettings` must replace the in-memory settings with what the file holds. Today it unmarshals over the previous `prefs`, so leftovers leak from one load to the next. Several tests load a few files in one process and catch that.
- The value shows as `"%v pt"`-style text: `"11.6 pt"`, `"11 pt"`, `"13 pt"`.
- The read-only footer must contain `"Changes apply but are not saved"` and the reason text from brief §2 exactly: `"settings.json is newer than this version of GoRex"`, `"settings.json could not be read"`, `"the settings backup could not be written"`. The normal footer is exactly `"Changes apply and save as you make them."`.

## Tests authored

They are all in one new file, `settings_test.go` (package `main`, 15 test functions). They all sit at the **unit / headless-integration** tier: plain `go test` with `ui.NewTester`, temp dirs through `GOREX_DIR`, and recorders in place of the native calls (brief §6). Nothing cheaper reaches these behaviours, and no browser or live window is needed. Every call into mygo goes through `within`, which fails after 5 s instead of hanging (brief §6 "No test hangs").

| Test | Behaviour pinned | Criterion |
|---|---|---|
| `TestSettingsFirstLaunch` | With no file, the app loads 11.6 pt and the System theme, View ▸ Appearance ▸ System is checked (one spelling, `"system"`), and Settings shows General, its line, Appearance System/Light/Dark, PROFILES, Default, built-in and the saving footer. No "New Profile" | 1, §3 |
| `TestSettingsMenuItem` | On darwin: GoRex ▸ Settings… with `CmdOrCtrl+,`, after About, and the full app menu keeps About/Services/Hide/Hide Others/Show All/Quit. Elsewhere: Edit ▸ Settings… with `Ctrl+,`. `shortcut{Ctrl, KeyComma}.String()` is `"Ctrl+,"`, `cmdSettings.Other` is Ctrl+,, `cmdSettings` is in `commands` (Windows' `shortcuts()` loop), and the palette shows `⌘,` / `Ctrl+,` | 2, §5 |
| `TestSettingsOpensOnce` | The menu item opens Settings with **no main window and no focused window**. After that, the menu, the palette's "Settings…" and `openSettings` focus that one window and never make a second | 2, 11 |
| `TestCommandsFromSettings` | `runsIn` over every command in `commands`: from the main window everything runs. From Settings or a nil window, only the app-wide set runs (Settings…, New Tab, Bigger/Smaller/Actual Size, the 3 Appearance commands, Quit and End All Sessions), with or without a main window | 10, §4 |
| `TestMenuClicksFromSettings` | Real menu `Click`s. With no main window, Bigger Text from Settings and from a nil window changes the size. Close Pane from Settings calls `closeWindow(settingsWin)`. With the main window open, New Tab from Settings calls `focusWindow(a.win)`, and Close Pane from the main window does not close Settings | 10, §4 |
| `TestEndAllNeedsMainWindow` | With `a.win == nil`, `cmdEndAll.Run` leaves `a.quitting` false and the session server still running (its socket is still there) | §4 |
| `TestSettingsAppearanceInStep` | Clicking Dark in Settings applies Dark, checks View ▸ Appearance ▸ Dark only, and saves `"dark"`. The palette's Light checks Light. The menu's System item applies System, checks it and saves `"system"` | 3, §3 |
| `TestSettingsTextSize` | Default's page shows "Text size" and "11.6 pt". − goes 11.6 → 11 (saved, shown), ⌘0 → 11.6 is shown, + goes 11.6 → 12 (saved), ⌘+ → 13 is shown. A click halfway along the slider (between − and +) lands on a whole point near 20, and it is saved | 4 |
| `TestSettingsMigratesV1` | 5 v1 files, one per branch: dark+14; no appearance → `"system"`; no fontSize → 11.6; 50 → 32; 7 → 8. Each one: the size and theme are applied, `settings.json` is **already v2 when `loadSettings` returns**, and `settings.v1.json` holds the original bytes | 5, 6, 8 |
| `TestSettingsMigrationOverwritesOldBackup` | An old `settings.v1.json` is replaced by the bytes being migrated | §1 |
| `TestSettingsV2Load` | v2 fontSize 100 → 32; no `profiles` → Default restored at 11.6; Default renamed "Mine" → saved as "Default" and its 15 kept; Default without fontSize → 11.6 | 5, 7 |
| `TestSettingsSurviveRestart` | After setFontSize(15) and setAppearance("dark"), the file is v2, and the next `loadSettings` restores 15 and Dark | 8 |
| `TestSettingsReadOnly` | 3 cases: version 3, invalid JSON, and a v1 file whose backup cannot be written (`settings.v1.json` is a non-empty directory). Each one: no crash; the loaded size is right (11.6 for invalid, 14 for the failed backup); later `setFontSize` and `setAppearance` **still apply** but `settings.json` is byte-for-byte unchanged; no backup is made of a non-v1 file; the Settings footer says not saved, and gives the reason | 8, §2, §6 |
| `TestSettingsKeysOnWindows` | **Windows only (skipped elsewhere).** In Settings, Esc, Ctrl+W and Ctrl+Shift+W call `closeWindow(settingsWin)`, Ctrl+, focuses the one window, and Ctrl+= changes the size | 9, 10 |
| `TestReadmeSettings` | README's Shortcuts table has a ⌘, / Ctrl+, row, the Shift-rule prose names Ctrl+, together with Ctrl+=, and the file declaring `settingsView` is a new one-word lowercase `.go` file (not `settings.go`) listed in the Layout table | 13 |

Criterion 12 (`go vet ./...` and `go test -count=1 -race ./...` pass) is the suite itself: T-green runs both.

### Not covered headless (for U and S)

These have no seam in the brief, so no test that would fail for the right reason can reach them. U should walk them live; S should read the code:
- **Dock reopen (`OnActivate`, criterion 11).** The handler is an inline closure in `main()`, and `a.open()` connects to a server and calls `NewWindow`.
- **Show Tab 1–9 from Settings do nothing (criterion 10).** These items are not `command`s, and with a main window their action is queued on mygo's main loop through `a.do`, which never runs headless. Commands that need the main window are pinned through `runsIn`, but the Show Tab items' use of the check is not.
- **Settings' selected Appearance highlight, and the live preview at the chosen size and appearance.** These are visual only. The `ui.Tester` API exposes no selection state.
- **Linux Ctrl+W in Settings** comes from mygo's `RoleClose`, not GoRex code.
- **Windows keys.** `TestSettingsKeysOnWindows` skips on darwin and Linux, and this repo has no Windows CI. It is cross-compiled only: `GOOS=windows go vet .` passes against the scratch stubs. Running it needs a Windows host.

## RED confirmation

**1. In the worktree (the real tree, no production code changed):** the test package fails to build, and every error is an identifier from the contract above. There is no typo, no import error and no setup error. Command: `go test -count=1 -gcflags=-e .` (full output in `evidence/t-red/compile-red.txt`):

    # gorex [gorex.test]
    ./settings_test.go:91:42: undefined: newSettingsWindow
    ./settings_test.go:91:61: undefined: focusWindow
    ./settings_test.go:91:74: undefined: closeWindow
    ./settings_test.go:91:87: undefined: setThemeSource
    ./settings_test.go:100:5: a.settingsWin undefined (type *App has no field or method settingsWin)
    ./settings_test.go:236:23: a.settingsView undefined (type *App has no field or method settingsView)
    ./settings_test.go:317:5: undefined: cmdSettings
    ./settings_test.go:355:30: a.openSettings undefined (type *App has no field or method openSettings)
    ./settings_test.go:377:9: a.runsIn undefined (type *App has no field or method runsIn)
    ... (30 errors, all of these 9 names)
    FAIL	gorex [build failed]

**2. Assertion-level RED (scratch copy, not committed).** This shows that each test fails on its own assertion and not only on the build. It ran in a copy of the tree in the session scratchpad, with one stub file (`evidence/t-red/stubs.go.txt`) that declares the contract names with **today's behaviour**:
- `openSettings` and `settingsView` do nothing;
- `runsIn` follows today's `run`: everything runs while there is a main window, and New Tab runs without one;
- `cmdSettings` is a zero command that is in no menu.

The copy has two more one-line changes, and neither is feature logic. A `settingsWin` field is added to `App`. `applyAppearance` calls `setThemeSource`, the seam alone, without which loadSettings hangs headless. Command: `go test -count=1 -race -v .` (full output in `evidence/t-red/red-output.txt`). Results:

    --- PASS: TestPanesAndTabs / TestRestore / TestLabels / TestCloseButtons / TestPaletteScrolls   (existing suite unaffected)
    --- FAIL: TestSettingsFirstLaunch          Settings does not show "General"; texts []
    --- FAIL: TestSettingsMenuItem             the app menu has no Settings… with CmdOrCtrl+,: <nil>
    --- FAIL: TestSettingsOpensOnce            the menu has no item "Settings…"
    --- FAIL: TestCommandsFromSettings         Split Right from Settings: runs true, want false … Appearance: System from Settings, no main window: runs false, want true
    --- FAIL: TestMenuClicksFromSettings       Bigger Text from Settings with no main window: text size 11.6, want 12.6; Close Pane from Settings closed []; New Tab from Settings focused [], not the main window
    --- FAIL: TestEndAllNeedsMainWindow        Quit and End All Sessions with no main window set quitting; the session server shut down, ending the sessions: stat …/server.sock: no such file or directory
    --- FAIL: TestSettingsAppearanceInStep     ui: no element shows "Dark"; texts []
    --- FAIL: TestSettingsTextSize             ui: no element shows "Default"; texts []
    --- FAIL: TestSettingsMigratesV1/{dark,system,no_size,too_big,too_small}   settings.json is not version 2: {"appearance":"dark","fontSize":14}; too_big: text size 11.6, want 32; too_small: text size 7, want 8
    --- FAIL: TestSettingsMigrationOverwritesOldBackup   settings.v1.json holds "{\"old\":true}", want the original
    --- FAIL: TestSettingsV2Load/{too_big,no_profiles,renamed,no_size}          text size 14, want 32 / 11.6 / 15 / 11.6 (today's loader ignores v2's profiles); settings.json is not version 2
    --- FAIL: TestSettingsSurviveRestart       settings.json is not version 2: {"appearance":"dark","fontSize":15}
    --- FAIL: TestSettingsReadOnly/{newer,invalid,backup_fails}                 settings.json was overwritten: {"appearance":"light","fontSize":17}; Settings does not say changes are not saved because "…"
    --- SKIP: TestSettingsKeysOnWindows        Settings handles these keys itself only on Windows
    --- FAIL: TestReadmeSettings               the Shortcuts table has no row with ⌘, and Ctrl+,; the Shift rule does not name Ctrl+, with Ctrl+=; README's Layout table does not list stubs.go
    FAIL	gorex	3.738s

Every one of the 14 tests that runs on darwin fails on a feature assertion. None fails on a panic, a hang or a setup error. `go vet .` passes on the scratch copy for darwin, `GOOS=linux` and `GOOS=windows`. The test file is `gofmt`-clean.

## Ready for F

The RED is valid. F may implement against `settings_test.go`. GREEN is: `go vet ./...` and `go test -count=1 -race ./...` pass, with all 14 darwin-runnable tests green, and `TestSettingsKeysOnWindows` green on a Windows host if one is available.
