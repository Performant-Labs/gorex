# Brief: Settings window, General and Default's text size, settings v2 (issue #1, delivery 1)

Issue: https://github.com/Performant-Labs/gorex/issues/1 — this story is **delivery 1** of its
"Suggested delivery". Profiles beyond Default (delivery 2) are out of scope here.

## Problem

Appearance and text size can be changed only from the View menu, keyboard shortcuts and the command
palette. GoRex has no Settings window, and the standard ⌘, item does nothing. `settings.json`
(`settings.go`) holds two values, `appearance` and `fontSize`, in a flat v1 format that cannot hold the
profiles delivery 2 adds.

## The work

### 1. Settings file v2 and the migration (`settings.go`)

The new format, with only the built-in Default profile in this story:

    {
      "version": 2,
      "appearance": "system",
      "profiles": {
        "default": { "name": "Default", "fontSize": 11.6 }
      }
    }

- `appearance` is `"system"`, `"light"` or `"dark"`. Today `""` means System.
- `profiles.default` always stores every value. `"default"` is the only fixed profile id.
- **Migration.** A file without `version` is v1. Before overwriting it, copy it byte for byte to
  `settings.v1.json` next to it. Its `appearance` stays at the top level (`""` becomes `"system"`), and its
  `fontSize` moves to `profiles.default.fontSize`. A user's appearance and text size look exactly the same
  after the migration.
- **Clamping.** Text size is 8–32 pt everywhere. Today loading accepts 6–40 while `setFontSize` clamps to
  8–32, so a v1 value outside 8–32 is clamped on load.
- **Robustness.** A missing file gives the defaults (System, 11.6 pt). An unreadable or invalid file must
  not crash the app, and must not destroy the user's file. A file with a `version` newer than 2 is left
  untouched on disk.
- **Default cannot be removed.** If `profiles.default` is missing or lacks `name` or `fontSize` (for
  example after a hand edit), it is restored with the defaults for whatever is missing. Its name is
  always "Default".

### 2. The Settings window

Wireframes, committed in this branch:
- `docs/wireframes/settings-profiles/1-general.png` (General, first launch)
- `docs/wireframes/settings-profiles/2-profile-overrides.png` (a profile's page; for Default, use only its
  text size row, without the override and inheritance parts, which are delivery 2)
- The SVG sources and `gen.py` are next to them.

- **Opening it.** On macOS: GoRex ▸ **Settings…** (⌘,). On Windows and Linux: **Edit ▸ Settings…**
  (Ctrl+,), because those platforms have no app menu. Pressing the shortcut, or choosing the item, while the
  window is open brings that one window to the front. There is never a second Settings window.
- The palette gets a **"Settings…"** command.
- **Windows.** Native-UI windows run no menu shortcuts on Windows (`shortcuts()` in `commands.go`). Ctrl+,
  reaches the main window through the existing `command` table loop in `shortcuts()` (see "Reuse &
  extend-vs-new": no hand-written case). The Settings window needs its own Esc and Ctrl+W handling to
  close it.
- Ctrl+, on Windows and Linux breaks the README's rule that GoRex shortcuts take Shift, under the same
  exception as Ctrl+= and Ctrl+- for text size: in legacy terminal encoding Ctrl+, sends nothing. Record
  this exception in README.md next to the existing one.
- **Sidebar.** **General**, then the **PROFILES** heading with one entry, **Default** with the label
  "built-in". The "+ New Profile…" footer is delivery 2: leave it out.
- **General page.** Title "General" and the line "These apply to the whole app, whatever profile a pane
  uses." **Appearance** has three choices, System, Light and Dark, each with a small swatch, and the
  current one is selected. Choosing one changes the whole app at once.
- **Default page.** Title "Default". **Text size** is a slider over 8–32 pt in 1 pt steps, with − and +
  buttons and the value shown (for example "11.6 pt"; a value that is not a whole number stays as it is
  until it is changed). Under it is a **live preview**, a few lines of terminal text drawn in the app's
  current appearance at the chosen size.
- **No OK or Apply.** Changes apply immediately to every pane and save automatically. The footer says
  "Changes apply and save as you make them."

### 3. Keeping everything in step

- ⌘+, ⌘− and ⌘0 (Ctrl+= / Ctrl+- / Ctrl+0 elsewhere) keep their behaviour: they change Default's text
  size, so every pane changes (⌘0 resets it to 11.6 pt). An open Settings window shows the new value at once.
- View ▸ Appearance's radio items are built once in `App.menu()` (`commands.go`), and their `Checked`
  state is fixed at that point. They must show the right choice after a change from Settings or the
  palette, and Settings must show a change made from the menu or the palette.
- **Where the syncing happens.** `App.setAppearance` and `App.setFontSize` are the only writers (see
  "Reuse & extend-vs-new"). After writing, they call `MenuItem.SetChecked` on the three View ▸ Appearance
  radio items, which `App` keeps when `menu()` builds them, and `Invalidate()` the Settings window when it
  is open. `a.do` redraws only the main window, so it is not enough on its own. If `SetChecked` cannot run
  under `go test` (no backend), guard it so a headless `App` with no menu skips it.
- **One spelling of System.** In memory and in the file, System is `"system"`. Change `cmdSystem`
  (`commands.go`) and the menu's `appearance("System", "")` to pass `"system"`. Normalise `""` only in the
  v1 migration, so the menu check, the Settings selection and the file always agree.

### 4. Commands while Settings is the key window

On macOS the main menu handles key equivalents whichever window is in front: mygo's darwin surface does
not override `performKeyEquivalent`, and the macOS Window menu role has no Close item. Today `item()` in
`App.menu()` ignores the `*mygo.Window` its `Click` receives, and `App.run` acts only on `a.win`. Without a
change, ⌘W in Settings would run Close Pane and kill the main window's focused session, and ⌘T, ⌘D and ⌘K
would act on the hidden main window.

- Menu items built by `item()` check the window their `Click` receives.
- When that window is the Settings window, Close Pane's item (⌘W on macOS, Ctrl+Shift+W elsewhere)
  closes Settings. No pane or tab command acts on the main window from it.
- `App.run`'s "no main window" exception, which today covers only New Tab, also covers Settings, so
  GoRex ▸ Settings… works after the main window is closed.
- **App lifecycle.** `OnActivate` reopens the main window only when `!hasVisibleWindows && a.win == nil`
  (`main.go`). With Settings open and the main window closed, clicking the Dock icon must still bring the
  main window back: the reopen depends on `a.win == nil` alone. Settings-window state lives on `App`
  outside `App.reset()`, which clears only per-main-window state.

### 5. Reuse & extend-vs-new

- **EXTEND** `App.setAppearance` and `App.setFontSize` (`settings.go`) as the only writers of appearance
  and Default's text size. The Settings controls, the menu, the palette and ⌘+ / ⌘− / ⌘0 all call them.
  Settings never writes `prefs`, `termFont` or `saveSettings()` itself.
- **EXTEND** the `command` table in `commands.go` with `cmdSettings`: Title "Settings…", Accel
  "CmdOrCtrl+,", Keys "⌘,", `Other: shortcut{ui.Ctrl, ui.KeyComma}`. Add it to `paletteCommands`. Its menu
  item comes from `item(&cmdSettings)`. Windows' main-window Ctrl+, comes from the existing loop in
  `shortcuts()`, not a hand-written case. Add `ui.KeyComma: ","` to `shortcut.String()`'s key names.
- **Menu placement.** `{Role: mygo.RoleAppMenu}` uses mygo's default submenu, and mygo replaces it only
  when `Submenu` is set. So on darwin, write the app menu out in full: About, Settings…, Services, Hide,
  Hide Others, Show All, Quit. Add Edit ▸ Settings… only off darwin, with the same `runtime.GOOS`
  branching `menu()` already uses.
- **FOLLOW** `App.open` (`main.go`) to create the Settings window: `mygo.WindowOptions` with
  `Content: ui.View(...)`, and `OnClosed` clearing the field.
- **REUSE** `attach`'s display-only terminal for the live preview: `terminal.New` with `nopConn{}`,
  `termFont` and `lightTerm` / `darkTerm`, then `Feed` (`state.go`). Hold one terminal across frames, feed
  it sample text, and call `SetFont` when the size changes. Do not paint monospace text with `ui.Painter`,
  and do not keep a second copy of the terminal colours.
- **REUSE** the `colors` struct and `colorsOf(c)` (`style.go`) for the window's colours. Any new colour is
  added to both `lightColors` and `darkColors`.
- **NEW**, justified because no settings UI or second window exists to extend: the Settings window and its
  view, `func (a *App) settingsView(c *ui.Context)`, in one new file with a one-word lowercase name, listed
  in README's Layout table. The v2 format, the migration and the mutators stay in `settings.go`.

### 6. Test seams

- **Opening Settings.** `func (a *App) openSettings()` focuses `a.settingsWin` when it is set. Otherwise
  it creates the window through one package-level func variable: in production `mygo.NewWindow` with
  `StateKey "settings"`, `Content: ui.View(a.settingsView)`, and `OnClosed` clearing the field. `go test`
  cannot create a real window (`mygo.NewWindow` off the main thread waits for `App.waitReady`), so tests
  replace the variable to check that a second open focuses the first window instead of making another.
- **The page.** `settingsView` is driven headless with `ui.NewTester`, the way `TestPaletteScrolls` drives
  the palette, with `a.win == nil` as in `newTestApp`.
- **Parsing.** Parse, migrate, clamp and restore-Default form one pure function from bytes to settings.
  `loadSettings` is a thin wrapper that does the IO and applies the result (`termFont.Size`, `mygo.Theme`).
  Tests isolate the file with `t.Setenv("GOREX_DIR", …)`, as `newTestApp` does.
- **Write gate.** `setAppearance` and `setFontSize` call `saveSettings()` unconditionally today, and it
  overwrites `settings.json`. A file whose `version` is newer than 2, or that could not be parsed, is
  loaded **read-only**: `saveSettings` does not write it, and the app runs on defaults (or what it could
  read). So "newer version left untouched" and "invalid file not destroyed" hold.

### 7. Forward-compat

Delivery 2 builds on this v2 format. Non-Default profiles store only their overrides, and a missing key
means "inherit". So the Go profile type must be able to represent a missing field (for example a pointer,
or omitempty with zero meaning unset), and `"default"` is a named constant. Delivery 2 then adds profiles
without changing the type. In this delivery `termFont` is derived from Default.

## Out of scope (delivery 2)

Profiles other than Default; the New Profile sheet; rename, duplicate and delete; per-pane profiles and
the `profile` field in `layout.json`; the pane header chip; View ▸ Profile; New Tab with Profile; the
profile-related palette entries; overrides, Reset and inheritance UI.

## Acceptance

1. On first launch (no settings file), Settings shows General with Appearance: System, and one profile,
   Default, labelled built-in.
2. ⌘, (Ctrl+, on Windows and Linux, including on Windows where native windows run no menu shortcuts)
   opens Settings, or brings the open window to the front. The menu item and the palette's "Settings…" do
   the same.
3. Changing Appearance in Settings changes the whole app immediately, and View ▸ Appearance shows the
   same choice. A change from the menu or the palette shows in an open Settings window.
4. Changing Default's text size (slider, −, +) changes every pane immediately and is saved. ⌘+, ⌘− and ⌘0
   still work as today, and an open Settings window follows them.
5. The value is clamped to 8–32 everywhere, including when loading.
6. A v1 file migrates to v2 on first load: appearance and text size are unchanged (`""` becomes
   `"system"`), and the original file is kept byte for byte as `settings.v1.json`.
7. Default cannot be removed or renamed by editing the file: a missing or incomplete `profiles.default`
   is restored.
8. Settings survive a restart. A missing, invalid or newer-version file never crashes the app or loses the
   user's file.
9. On Windows, Esc and Ctrl+W close the Settings window.
10. ⌘W (Ctrl+W on Windows and Linux) with Settings frontmost closes Settings and closes no pane. ⌘T, ⌘D
    and ⌘K with Settings frontmost do not act on the hidden main window.
11. GoRex ▸ Settings… works after the main window is closed, and clicking the Dock icon with only
    Settings open brings the main window back.
12. `go vet ./...` and `go test -count=1 -race ./...` pass. New behaviour that can be tested without a
    window (the v2 format, migration, clamping, restoring Default, single-window bring-to-front logic,
    menu and Settings state staying in step) has tests that fail if the behaviour is removed.

## Rigor

second-opinion
