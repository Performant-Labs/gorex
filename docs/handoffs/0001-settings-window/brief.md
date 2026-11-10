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
  Go's `json` reads a missing number as 0, so the migration cannot tell a missing `fontSize` from a
  hand-written `"fontSize": 0`. Both are treated as missing and become 11.6. Only a value greater than 0
  is clamped to 8–32.
- **A failed backup stops the migration.** If `settings.v1.json` cannot be written (for example a
  read-only directory or a full disk), the v1 file is not overwritten. The app runs on the migrated values
  in memory, and the file is loaded read-only for the rest of the run, as in §6 "Write gate". So the user's
  v1 file is never replaced unless its byte-for-byte copy exists. The next launch tries the migration
  again.
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
  close it. Because `shortcuts()` runs only in the main window's view, the Settings window's own key
  handling on Windows also handles **Ctrl+,** (`cmdSettings.Other`): it calls `openSettings()`, which
  focuses the one Settings window and never makes a second (criteria 2 and 9).
- Ctrl+, on Windows and Linux breaks the README's rule that GoRex shortcuts take Shift, under the same
  exception as Ctrl+= and Ctrl+- for text size: in legacy terminal encoding Ctrl+, sends nothing. Record
  this exception in README.md next to the existing one. "The existing one" is the Ctrl+= Ctrl+- Ctrl+0 row
  of the Shortcuts table (`README.md:54`), which goes without Shift; the rule itself is the paragraph at
  `README.md:57-60` (excerpt in §8), and today no sentence there names the text-size exception. Add the
  Ctrl+, row to the table and a sentence to that paragraph naming Ctrl+, together with Ctrl+= Ctrl+-
  Ctrl+0 as the shortcuts that go without Shift, and why (criterion 13).
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
  tests build `a.menu()`, keep the three Appearance items, and assert `IsChecked()`. No guard is needed:
  the queued update is a `postMain`, which does not wait, and a probe on the pinned mygo under `go test`
  showed `SetChecked` returning on items built by `NewMenu` with `IsChecked()` already updated (§8,
  "`SetChecked` headless").
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
  (`main.go:46-50`). With Settings open and the main window closed, `hasVisibleWindows` is true, so today's
  handler would **not** reopen the main window. With Settings open and the main window closed, clicking
  the Dock icon must still bring the main window back, so **change `main.go:46-50`** to drop the
  `hasVisibleWindows` condition: the handler becomes `if a.win == nil { a.open() }` (the parameter stays
  in the signature, unused). The reopen then depends on `a.win == nil` alone. This regresses no other
  path:
  - **No window at all** (Dock click after the main window closed, Settings not open): `hasVisibleWindows`
    is false and `a.win == nil`, so it reopens, as today.
  - **Main window open** (visible, hidden or minimized): `a.win != nil`, so it opens nothing, as today.
  - **App quits when all windows close**: that is `OnWindowAllClosed` (`main.go:51-55`), which this
    change does not touch. It quits only off darwin, and `OnActivate` fires only on darwin: its only
    caller is `applicationShouldHandleReopen:hasVisibleWindows:` (mygo `internal/darwin/app.go:305-306`).
    Off darwin, closing the main window while Settings is open leaves the app running until Settings
    also closes, and then `OnWindowAllClosed` quits it.
  - **⌘-Tab / app activation**: `applicationShouldHandleReopen` fires on a Dock click or reopen, not on
    every activation, so ⌘-Tab does not reopen the main window (handoff-A round 2, R1 warn 10).

  Settings-window state lives on `App` outside `App.reset()`, which clears only per-main-window state.

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
- **EXTEND** the `OnActivate` handler in `main()` (`main.go:46-50`): drop `!hasVisibleWindows &&` so it
  reopens the main window whenever `a.win == nil`, including when Settings is the only open window (§4
  "App lifecycle", criterion 11). No other lifecycle handler changes.
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
  `302-308`). `update` changes the item under the lock, then hands the native update to `postMain`
  without waiting (`menu.go:243-264`, abridged):

      243	func (it *MenuItem) update(fn func()) {
      244		menuMu.Lock()
      245		fn()
      ...
      258		if len(snaps) > 0 {
      259			postMain(func() {
      260				for _, s := range snaps {
      261					backend().UpdateMenuItem(s)

  `postMain` is `loop.post` (`loop.go:81-82`), which appends to the queue and, on the first post, calls
  `backend().Signal()` and returns (`loop.go:19-31`); unlike `onMain` (`loop.go:87-100`) it never waits.
  **Verified safe under `go test` after `NewMenu` ran:** the probe
  (`scratchpad/mygoprobe/probe_test.go`, re-run for this amendment) builds radio items with
  `mygo.NewMenu`, calls `SetChecked(true)` on one and reports `SetChecked(dark): returned` and
  `after SetChecked: light=false dark=true`. The probe and its full output are committed in
  `docs/handoffs/0001-settings-window/evidence/mygoprobe/` (`probe_test.go`, `go.mod`, `output.txt`), built
  against the same pinned mygo (`v0.3.5-0.20261008155018-bad73d119dc8`) as this repository's `go.mod`, and
  re-run for this amendment with the same result. So §3 needs no guard: tests call `a.menu()` and
  `SetChecked`, and assert `IsChecked()`. The probe ran on darwin only; Linux and Windows share the
  same `update` / `postMain` path in `menu.go` and `loop.go`.
- **`onMain` waits** (§6, §8 list): `loop.go:87-100`:

      87	func onMain(fn func()) {
      88		if isMainThread() {
      89			fn()
      90			return
      91		}
      92		done := make(chan struct{})
      93		if !loop.post(func() {
      ...
      99		<-done
      100	}

  `init` locks the main goroutine to the main thread (`mygo.go:72-75`, `runtime.LockOSThread()`), so a
  test goroutine is never on it and waits on `<-done` forever. The same probe reports
  `Theme.SetSource(dark): BLOCKED (>3s)`, `(&mygo.Window{}).Focus(): BLOCKED (>3s)` and
  `(&mygo.Window{}).Invalidate(): returned`.
- **`Focus`, `Close`, `Invalidate`** (§6): `window.go:513-519` is `func (w *Window) do(fn func(n
  platform.Window)) { onMain(func() { … }) }`; `window.go:607` is
  `func (w *Window) Focus() { w.do(platform.Window.Focus) }`; `window.go:534` is
  `func (w *Window) Close() { onMain(func() { w.close() }) }`. `content.go:145-148`: `Invalidate` returns
  when `w.content == nil`, before its `postMain`.
- **macOS menu click's window** (§4): `app.go:591`:

      591	func (appHandler) MenuItemClicked(id int)        { menuItemClicked(id, FocusedWindow()) }

- **`RoleClose`** (§4): `performRole` (`menu.go:507-508`):

      507		case RoleClose:
      508			win.close()

- **`OnActivate` is darwin-only** (§4 "App lifecycle"): `Application.OnActivate` (`app.go:443-444`) adds
  to `onActivate`, which `appHandler.Activated` fires (`app.go:576-577`); the only caller of `Activated`
  in the module is `internal/darwin/app.go:305-306`:

      305	method("applicationShouldHandleReopen:hasVisibleWindows:", func(self id, _ objc.SEL, app id, visible bool) bool {
      306		theBackend.h.Activated(visible)

  GoRex's handler today (`main.go:46-50`):

      46	mygo.App.OnActivate(func(hasVisibleWindows bool) {
      47		if !hasVisibleWindows && a.win == nil {
      48			a.open()
      49		}
      50	})

- **`ui.KeyComma`** (§5): mygo's `ui/keys.go:45` declares `KeyComma` in the `Key` constants, next to
  `KeyMinus` and `KeyEqual` (`:43-44`), which `shortcut.String()` already maps (`commands.go:120`). The
  terminal plugin maps it to the character `,` (`plugins/terminal/keys.go:21`: `ui.KeyComma: …, ','}`).
  So `ui.KeyComma: ","` in `shortcut.String()`'s key names gives the accelerator `Ctrl+,`, and `,` is a
  valid accelerator key (handoff-A round 2, `internal/accelerator`).
- **README's Shift rule** (§2): `README.md:54` and `:57-60`:

      54	| ⌘+ ⌘− ⌘0 | Ctrl+= Ctrl+- Ctrl+0 | Text size |
      ...
      57	Elsewhere than on macOS, Control and a letter are the terminal's (Ctrl+C,
      58	Ctrl+D, Ctrl+W…), so GoRex's shortcuts take Shift too, as copying and
      59	pasting do (Ctrl+Shift+C, Ctrl+Shift+V), and none take Control and Alt,
      60	which type characters (AltGr) on many keyboards.

  The claim that Ctrl+, sends nothing in legacy terminal encoding is issue #1's ("Ctrl+, on Windows and
  Linux"). It is not verified against source here: mygo's terminal encodes keys through Ghostty's native
  library (`plugins/terminal/internal/vt/encode.go`, `KeyEncoder`), whose legacy tables are not in the
  Go module.
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
5. The value is clamped to 8–32 everywhere, including when loading, except as provided by criterion 6: a
   v1 file with no `fontSize` has no value to clamp, and migrates to 11.6 (§1 "v1 without `fontSize`").
6. A v1 file migrates to v2 on first load: appearance and text size are unchanged (`""` becomes
   `"system"`), except that a text size outside 8–32 is clamped per criterion 5 and a missing text size
   becomes 11.6, and the original file is kept byte for byte as `settings.v1.json`.
7. Default cannot be removed or renamed by editing the file: a missing or incomplete `profiles.default`
   is restored.
8. Settings survive a restart. A missing, invalid or newer-version file never crashes the app or loses the
   user's file, including after a setting is changed later in the same run (§6 "Write gate").
   If the `settings.v1.json` backup cannot be written, the v1 file is left as it was (§1 "A failed backup
   stops the migration").
9. On Windows, Esc and Ctrl+W close the Settings window, and Ctrl+, with Settings focused keeps that one
   window in front without opening a second (§2 "Windows").
10. With Settings frontmost, its close shortcuts close Settings and close no pane: ⌘W on macOS; on
    Windows, Ctrl+W and Esc (Settings' own handling, §2); on Linux, Ctrl+W (Window ▸ Close Window,
    `RoleClose`); and on Windows and Linux, Ctrl+Shift+W (Close Pane's shortcut there, which from Settings
    closes Settings, §4). With Settings frontmost, ⌘T, ⌘D, ⌘K and ⌘1–⌘9 (their Ctrl+Shift equivalents
    elsewhere) do not act on the hidden main window.
11. GoRex ▸ Settings… works after the main window is closed, and clicking the Dock icon with only
    Settings open brings the main window back (the `OnActivate` change, §4 "App lifecycle"). Clicking the
    Dock icon with no window open still reopens the main window, clicking it with the main window open
    opens no second one, and off darwin the app still quits when its last window closes.
12. `go vet ./...` and `go test -count=1 -race ./...` pass. New behaviour that can be tested without a
    window (the v2 format, migration, clamping, restoring Default, single-window bring-to-front logic,
    menu and Settings state staying in step) has tests that fail if the behaviour is removed.
13. README.md records Ctrl+, in the Shortcuts table and names it, with Ctrl+= Ctrl+- Ctrl+0, as an
    exception to the Shift rule (§2), and its Layout table lists the new Settings source file (§5 "NEW").

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
- Round 2 amendment (brief review round 1 findings): `docs/handoffs/0001-settings-window/decisions.md` and
  `handoff-A.md` (re-read); `main.go`, `settings.go`, `commands.go`, `README.md`, `style.go` (re-read in
  full); `state.go:100-140` and `:190-260`, `view.go:15-30`, `app_test.go:40-75` and `:295-337`; issue #1
  (re-read); mygo `menu.go` (`SetChecked`, `update`, `prepareItem`, `performRole`, `roleSubmenu`,
  `MenuItem.Click` docs), `loop.go` (all), `app.go:315-325`, `:438-445`, `:576-577`, `:585-595`,
  `mygo.go:70-76`, `window.go:510-540` and `:607-615`, `content.go:145-153`, `ui/keys.go:40-50`,
  `plugins/terminal/keys.go` (all), `plugins/terminal/internal/vt` (search of `encode.go` and `vt_test.go`
  for the legacy Ctrl encoding: not found in Go source), `internal/darwin/app.go:159-164` and `:300-310`, and
  a search of the module for callers of `Activated` and `OnActivate`
- The headless probe `scratchpad/mygoprobe/probe_test.go` (session scratchpad of the architecture review),
  read and re-run with `go test -count=1 -v`
- Not re-read for this amendment: the wireframes and `gen.py` (no finding concerns them)
