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

  `11.6` is the shipped default. It is not a whole number, and it stays as it is in the file until the user
  changes the text size (see "Default page" in §2).

- `appearance` is `"system"`, `"light"` or `"dark"`. Today `""` means System.
- `profiles.default` always stores every value. `"default"` is the only fixed profile id.
- **Migration.** A file without `version` is v1. Before overwriting it, copy it byte for byte to
  `settings.v1.json` next to it. Its `appearance` stays at the top level (`""` becomes `"system"`), and its
  `fontSize` moves to `profiles.default.fontSize`. A user's appearance and text size look exactly the same
  after the migration, except that a text size outside 8–32 is clamped (see "Clamping" and acceptance
  criterion 5).
- **v1 without `fontSize`.** A v1 file may have no `fontSize` (the v1 struct tags it `omitempty`,
  `settings.go:18`), which today reads as 0 and is replaced by 11.6 (`settings.go:31-33`). Such a file
  migrates to `profiles.default.fontSize` 11.6, as "Default cannot be removed" below restores a missing
  value, and is not clamped up to 8.
- **Clamping.** Text size is 8–32 pt everywhere. Today loading accepts 6–40 while `setFontSize` clamps to
  8–32, so a v1 value outside 8–32 is clamped on load. Today's code (`settings.go`):

      31	if prefs.FontSize < 6 || prefs.FontSize > 40 {
      32		prefs.FontSize = defaultFontSize
      33	}
      ...
      63	size = min(max(size, 8), 32)

  Note that today's load replaces a value outside 6–40 with 11.6 rather than clamping it; after this change
  such a value is clamped to 8 or 32 like any other value outside 8–32.
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
  "built-in". The "+ New Profile…" footer is delivery 2: leave it out. Wireframe 1 draws that footer
  (bottom of the sidebar, note 4); read the wireframe with it ignored, and do not ship it from the image.
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
  is open. `a.do` redraws only the main window, so it is not enough on its own. `menu()` and `SetChecked`
  are safe headless: `SetChecked` changes `Checked` at once and only queues the native update (§8). So
  tests build `a.menu()`, keep the three Appearance items, and assert `IsChecked()`. No guard is needed.
- **One spelling of System.** In memory and in the file, System is `"system"`. Change `cmdSystem`
  (`commands.go`) and the menu's `appearance("System", "")` to pass `"system"`. Normalise `""` only in the
  v1 migration, so the menu check, the Settings selection and the file always agree.

### 4. Commands while Settings is the key window

On macOS the main menu handles key equivalents whichever window is in front: mygo's darwin surface does
not override `performKeyEquivalent`, and the macOS Window menu role has no Close item. Today `item()` in
`App.menu()` ignores the `*mygo.Window` its `Click` receives, and `App.run` acts only on `a.win`. Without a
change, ⌘W in Settings would run Close Pane and kill the main window's focused session, and ⌘T, ⌘D and ⌘K
would act on the hidden main window.

- On macOS a main-menu click's window is `FocusedWindow()`. On Linux and Windows it is the window whose
  menu bar was used.
- **One window check, one classification.** Add a field to the `command` table, next to `Hidden`, that
  marks a command **app-wide**: it runs whatever window is in front, and also when there is no main window.
  The app-wide commands are Settings…, New Tab (as today, it opens or brings the main window to the front
  first), Bigger Text, Smaller Text, Actual Size, the three Appearance commands, Quit and End All Sessions.
  Every other command, including the palette, needs the main window. One helper decides whether a command
  runs for the window a click came from. `item()` uses it, the Show Tab 1–9 items (built today without
  `item()`) use it too, so ⌘1–⌘9 in Settings do nothing, and `App.run` uses the same field in place of its
  New Tab–only exception. App-wide commands run directly, not through `a.do`, which drops work when
  `a.win == nil`.
- When the click came from the Settings window, Close Pane's item (⌘W on macOS, Ctrl+Shift+W elsewhere)
  closes Settings. No command that needs the main window acts on it from Settings.
- **The menu bar off darwin.** `Application.SetMenu` gives every window, Settings included, the full app
  menu bar on Linux and Windows. That is intended: its items go through the same window check. On Linux,
  Ctrl+W in Settings reaches Window ▸ Close Window (`RoleClose` closes the window it was clicked in). On
  Windows, Settings handles Esc and Ctrl+W itself (§2).
- **Settings' own queue.** `a.post`, `a.later` and `laterService` feed a queue that only the main view
  drains (`view.go`). `settingsView` calls `setAppearance` and `setFontSize` directly, and never posts into
  that queue or assigns `a.services`.
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

- **Every native call the Settings window makes goes through a seam.** Under `go test`, `mygo.NewWindow`,
  `(*mygo.Window).Focus`, `(*mygo.Window).Close` and `mygo.Theme.SetSource` all wait on a main loop that
  never runs, so a test calling any of them hangs (§8). Three package-level func variables:
  - `newSettingsWindow func(a *App) *mygo.Window`. In production it calls
    `mygo.NewWindow(mygo.WindowOptions{StateKey: "settings", Content: ui.View(a.settingsView), …})` and
    then registers `win.OnClosed(…)` to clear `a.settingsWin`. `OnClosed` is a method on `*mygo.Window`,
    not a `WindowOptions` field; `App.open` registers it the same way.
  - `focusWindow = (*mygo.Window).Focus`
  - `closeWindow = (*mygo.Window).Close`

  `func (a *App) openSettings()` calls `focusWindow(a.settingsWin)` when it is set, and otherwise
  `newSettingsWindow(a)`. Tests return `&mygo.Window{}` from the creator (its `Invalidate()` is safe
  headless, which §3 needs) and replace focus and close with recorders. They check that a second open
  focuses the first window instead of making another, and that Close Pane from Settings closes it.
- **Theme.** `applyAppearance` (`settings.go`) calls a package-level func variable whose production value
  is `mygo.Theme.SetSource`, never `mygo.Theme.SetSource` directly. Tests replace it, so `setAppearance`,
  `loadSettings` and the write-gate test run headless and can assert which theme was applied.
- **The page.** `settingsView` is driven headless with `ui.NewTester`, the way `TestPaletteScrolls` drives
  the palette, with `a.win == nil` as in `newTestApp`.
- **Parsing.** Parse, migrate, clamp and restore-Default form one pure function from bytes to settings.
  `loadSettings` is a thin wrapper that does the IO and applies the result (`termFont.Size`, `mygo.Theme`).
  Tests isolate the file with `t.Setenv("GOREX_DIR", …)`, as `newTestApp` does.
- **Write gate.** `setAppearance` and `setFontSize` call `saveSettings()` unconditionally today, and it
  overwrites `settings.json`. A file whose `version` is newer than 2, or that could not be parsed, is
  loaded **read-only**: `saveSettings` does not write it, and the app runs on defaults (or what it could
  read). So "newer version left untouched" and "invalid file not destroyed" hold. The read-only state lasts
  for the whole run of the app, not only the load: a later `setAppearance` or `setFontSize` (from Settings,
  the menu, the palette or ⌘+ / ⌘− / ⌘0) still changes the app but does not write `settings.json`. The
  tests for criterion 8 load such a file, then change a setting, and check the file on disk is unchanged.

### 7. Forward-compat

Delivery 2 builds on this v2 format. Non-Default profiles store only their overrides, and a missing key
means "inherit". So the Go profile type must be able to represent a missing field (for example a pointer,
or omitempty with zero meaning unset), and `"default"` is a named constant. Delivery 2 then adds profiles
without changing the type. In this delivery `termFont` is derived from Default.

### 8. Verified facts about the pinned mygo (`go.mod`)

Checked by the architecture reviewer against the pinned mygo source and a probe test built on it:

- `mygo.go:72-75`: `init` calls `runtime.LockOSThread`, so the main goroutine owns the main thread.
- `loop.go:87-100`: `onMain` hands work to the main loop and waits. Under `go test` no main loop runs, so
  it waits forever.
- `modules.go:259-274`: `Theme.SetSource` calls `onMain`, so it blocks headless.
- `window.go:513-519` (`Focus`), `:534` (`Close`) and `:607`: they call `onMain`, so they block on a
  zero `mygo.Window`.
- `window.go:1338`: `OnClosed` is a method on `*mygo.Window`.
- `content.go:145-148`: `Invalidate` returns early on a window with no content, so it is safe headless.
- `menu.go:243-264`: `update` (behind `SetChecked`) changes the item, then `postMain`s the native update
  without waiting. `NewMenu` and `SetChecked` return headless, and `IsChecked()` reports the new state
  right after `SetChecked`.
- `app.go:591`: a macOS main-menu click's window is `FocusedWindow()`. `menu.go:85-88`: on Linux and
  Windows it is the window whose menu bar was used. `menu.go:507-508`: `RoleClose` closes that window.
- `app.go:319-323`: off darwin, `Application.SetMenu` gives every window the app menu bar.
- `internal/darwin/app.go:305`: `OnActivate` fires from `applicationShouldHandleReopen`.

### 8. Verified source excerpts

The claims above, checked against this branch and against mygo
`v0.3.5-0.20261008155018-bad73d119dc8` (the version `go.mod` pins; paths below are in its module directory).

- **Load range and clamp** (§1): `settings.go:31-33` and `:63`, quoted under "Clamping".
- **Windows shortcuts** (§2, §5): `shortcuts()` handles commands only on Windows, by looping over the
  `commands` table, so a `command` with `Other` set needs no hand-written case (`commands.go:331-338`):

      331	if runtime.GOOS != "windows" {
      332		return
      333	}
      334	for _, cmd := range commands {
      335		if cmd.Other != (shortcut{}) && c.Shortcut(cmd.Other.Mods, cmd.Other.Key) {
      336			a.later(c, func() { cmd.Run(a) })

  `commands` is `paletteCommands` plus `cmdPalette`, `cmdPalette2` and `cmdClear` (`commands.go:72`), so
  `cmdSettings` reaches the loop by being in `paletteCommands`. Off darwin, `init()` turns `Other` into the
  menu accelerator (`commands.go:74-85`).
- **No `performKeyEquivalent` override** (§4): searching mygo's `internal/darwin` finds the selector only in
  `testhooks.go:58`, which sends `performKeyEquivalent:` to the main menu; no class adds a method for it.
- **Window menu role** (§4): `roleSubmenu` in `menu.go:672-681` gives the macOS Window menu Minimize, Zoom
  and Bring All to Front, with no Close item. Off darwin it is `{Role: RoleMinimize}, {Role: RoleClose}`,
  and `RoleClose` defaults to "Close Window", `CmdOrCtrl+W` (`menu.go:575-576`); `performRole` closes the
  `win` the click reports (`menu.go:507-508`).
- **App menu role** (§5): `prepareItem` fills a role's submenu only when none is set (`menu.go:325-327`):

      325	if it.Submenu == nil {
      326		it.Submenu = roleSubmenu(it.Role)
      327	}

  and `roleSubmenu(RoleAppMenu)` is About, Services, Hide, Hide Others, Show All, Quit (`menu.go:621-632`).
- **`SetChecked` headless** (§3): `SetChecked` calls `update` (`menu.go:219-226`), which touches the
  backend only when the item has a `uid`, set by `prepareItem` when `NewMenu` builds it (`menu.go:247-264`,
  `302-308`); it then posts to the main loop, whose first post calls `backend().Signal()`
  (`loop.go:18-31`). So an `App` whose `menu()` never ran has no items to call; whether the post is safe
  under `go test` after `menu()` ran is not verified here, hence the guard in §3.
- **`NewWindow` and `StateKey`** (§6): off the main thread `NewWindow` calls `App.waitReady()`
  (`window.go:294-295`). `WindowOptions.StateKey string` is a field (`window.go:153-160`), already used by
  `App.open` (`main.go:98`).
- **`Click`'s window** (§4): `item()` ignores it (`commands.go:128-131`):

      128	item := func(cmd *command) *mygo.MenuItem {
      129		return &mygo.MenuItem{Label: cmd.Title, Accelerator: cmd.Accel, Click: func(*mygo.MenuItem, *mygo.Window) {
      130			a.run(cmd)

  mygo documents `win` as "the window whose menu bar or context menu was chosen" on Linux and Windows, and
  for other menus "the focused window, which may be nil" (`menu.go:85-88`).
- **`run` and `OnActivate`** (§4): `run` returns when `a.win == nil` unless the command is New Tab
  (`commands.go:214-220`). `OnActivate` opens the window only when `!hasVisibleWindows && a.win == nil`
  (`main.go:46-50`). `reset()` is `main.go:155-159`.
- **Display-only terminal and headless tests** (§5, §6): `attach` builds `terminal.New(terminal.Options{Conn:
  nopConn{}, Transparent: true, Font: termFont, Theme: lightTerm, DarkTheme: darkTerm})` and calls `Feed`
  (`state.go:196-198`); `nopConn` is `state.go:252-257`; `SetFont` is used in `setFontSize`
  (`settings.go:69`). `newTestApp` uses `t.Setenv("GOREX_DIR", dir)` and `ui.NewTester(a.view, 1000, 620)`
  (`app_test.go:49`, `:71`), and `TestPaletteScrolls` drives the palette with `a.win == nil`
  (`app_test.go:303-337`).

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
   `"system"`), except that a text size outside 8–32 is clamped per criterion 5 and a missing text size
   becomes 11.6, and the original file is kept byte for byte as `settings.v1.json`.
7. Default cannot be removed or renamed by editing the file: a missing or incomplete `profiles.default`
   is restored.
8. Settings survive a restart. A missing, invalid or newer-version file never crashes the app or loses the
   user's file, including after a setting is changed later in the same run (§6 "Write gate").
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

## Source material read

- `docs/handoffs/0001-settings-window/brief.md`, `decisions.md`, `handoff-A.md`
- `docs/handoffs/1-brief-result-r1.md` (the reviewer's round 1 output)
- Issue #1 (https://github.com/Performant-Labs/gorex/issues/1)
- `settings.go`, `commands.go`, `main.go`, `state.go`, `style.go`, `app_test.go`, `README.md`, `go.mod`
- `docs/wireframes/settings-profiles/1-general.png` and `.svg`, `2-profile-overrides.png` and `.svg`, `gen.py`
- mygo `v0.3.5-0.20261008155018-bad73d119dc8`: `menu.go`, `loop.go`, `window.go` (`WindowOptions.StateKey`,
  `NewWindow`), `plugins/terminal/keys.go`, and a search of `internal/darwin` for `performKeyEquivalent`
