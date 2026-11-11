package main

// Tests of issue #1, delivery 1: the Settings window, settings.json v2 and
// its migration, and the menus, the palette and Settings keeping in step.
//
// Beyond the names the brief gives (newSettingsWindow, focusWindow,
// closeWindow, App.openSettings, App.settingsWin, App.settingsView,
// cmdSettings), these tests need:
//
//   - setThemeSource func(mygo.ThemeSource): the theme seam of §6, whose
//     production value is mygo.Theme.SetSource. applyAppearance calls it.
//   - (*App).runsIn(cmd *command, win *mygo.Window) bool: the one helper
//     of §4 that decides whether cmd runs for a menu click from win.
//   - In the Settings window, the − and + buttons of the text size are
//     found by the labels "Smaller Text" and "Bigger Text" (their text or
//     their accessibility Label), the Appearance choices by "System",
//     "Light" and "Dark", and the sidebar's profile by "Default".
//   - settingsView works when a test drives it directly with ui.NewTester,
//     without openSettings having run (§6 "The page").
//
// The native calls never run here: under go test they wait on a main loop
// that never runs (brief §8). Every call that reaches mygo runs through
// within, which fails after 5 s rather than hanging.

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"gorex/internal/rex"
)

// within runs fn, and fails if it does not return in 5 seconds, as a
// call waiting on the main loop would not.
func within(t *testing.T, what string, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s did not return in 5 s: it waits on a main loop that never runs headless", what)
	}
}

// seamLog records the native calls the seams of §6 stand in for.
type seamLog struct {
	created, focused, closed []*mygo.Window
	themes                   []mygo.ThemeSource
}

// settingsApp writes file as settings.json in a directory of its own (none
// when file is ""), replaces the native calls with recorders, loads the
// settings and returns an app with its menu built, and the directory.
func settingsApp(t *testing.T, file string) (*App, *mygo.Menu, *seamLog, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("GOREX_DIR", dir)
	if file != "" {
		if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(file), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	s := installSeams(t)
	registerFonts()
	termFont.Size = 0 // loading must set it
	within(t, "loadSettings", loadSettings)
	a := &App{}
	var m *mygo.Menu
	within(t, "App.menu", func() { m = a.menu() })
	return a, m, s, dir
}

func installSeams(t *testing.T) *seamLog {
	t.Helper()
	s := &seamLog{}
	oldNew, oldFocus, oldClose, oldTheme := newSettingsWindow, focusWindow, closeWindow, setThemeSource
	oldSize := termFont.Size
	t.Cleanup(func() {
		newSettingsWindow, focusWindow, closeWindow, setThemeSource = oldNew, oldFocus, oldClose, oldTheme
		termFont.Size = oldSize
	})
	newSettingsWindow = func(a *App) *mygo.Window {
		w := &mygo.Window{}
		s.created = append(s.created, w)
		a.settingsWin = w
		return w
	}
	focusWindow = func(w *mygo.Window) { s.focused = append(s.focused, w) }
	closeWindow = func(w *mygo.Window) { s.closed = append(s.closed, w) }
	setThemeSource = func(src mygo.ThemeSource) { s.themes = append(s.themes, src) }
	return s
}

func (s *seamLog) lastTheme() mygo.ThemeSource {
	if len(s.themes) == 0 {
		return "(none applied)"
	}
	return s.themes[len(s.themes)-1]
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-3 }

func wantSize(t *testing.T, what string, want float64) {
	t.Helper()
	if !near(float64(termFont.Size), want) {
		t.Errorf("%s: text size %v, want %v", what, termFont.Size, want)
	}
}

// v2File is settings.json as version 2 writes it.
type v2File struct {
	Version    *int    `json:"version"`
	Appearance *string `json:"appearance"`
	Profiles   map[string]struct {
		Name     *string  `json:"name"`
		FontSize *float64 `json:"fontSize"`
	} `json:"profiles"`
}

// checkSaved checks that settings.json in dir is version 2 and holds
// appearance and Default's text size.
func checkSaved(t *testing.T, dir, appearance string, size float64) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatalf("settings.json: %v", err)
	}
	var f v2File
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatalf("settings.json is not JSON: %v\n%s", err, b)
	}
	if f.Version == nil || *f.Version != 2 {
		t.Fatalf("settings.json is not version 2:\n%s", b)
	}
	if f.Appearance == nil || *f.Appearance != appearance {
		t.Errorf("settings.json's appearance is not %q:\n%s", appearance, b)
	}
	d, ok := f.Profiles["default"]
	switch {
	case !ok:
		t.Errorf("settings.json has no profiles.default:\n%s", b)
	case d.Name == nil || *d.Name != "Default":
		t.Errorf("profiles.default is not named \"Default\":\n%s", b)
	case d.FontSize == nil || !near(*d.FontSize, size):
		t.Errorf("profiles.default.fontSize is not %v:\n%s", size, b)
	}
}

// findItem finds the first item labeled label in items and their submenus.
func findItem(items []*mygo.MenuItem, label string) *mygo.MenuItem {
	for _, it := range items {
		if it.Label == label {
			return it
		}
		if found := findItem(it.Submenu, label); found != nil {
			return found
		}
	}
	return nil
}

func menuItem(t *testing.T, m *mygo.Menu, label string) *mygo.MenuItem {
	t.Helper()
	it := findItem(m.Items(), label)
	if it == nil {
		t.Fatalf("the menu has no item %q", label)
	}
	return it
}

// appearanceItem is View ▸ Appearance ▸ label.
func appearanceItem(t *testing.T, m *mygo.Menu, label string) *mygo.MenuItem {
	t.Helper()
	view := findItem(m.Items(), "View")
	if view == nil {
		t.Fatal("the menu has no View")
	}
	sub := findItem(view.Submenu, "Appearance")
	if sub == nil {
		t.Fatal("View has no Appearance")
	}
	it := findItem(sub.Submenu, label)
	if it == nil {
		t.Fatalf("View ▸ Appearance has no %q", label)
	}
	return it
}

// checked checks that View ▸ Appearance shows want, and only want.
func checked(t *testing.T, m *mygo.Menu, want string) {
	t.Helper()
	for _, label := range []string{"System", "Light", "Dark"} {
		if got := appearanceItem(t, m, label).IsChecked(); got != (label == want) {
			t.Errorf("View ▸ Appearance ▸ %s checked %v, want only %s checked", label, got, want)
		}
	}
}

// click chooses a menu item from win, as mygo's menuItemClicked does: a
// radio item is checked first, then its Click runs.
func click(t *testing.T, it *mygo.MenuItem, win *mygo.Window) {
	t.Helper()
	within(t, "the menu item "+it.Label, func() {
		if it.Type == mygo.MenuItemRadio {
			it.SetChecked(true)
		}
		it.Click(it, win)
	})
}

// clickText clicks the element of the Settings window showing or labeled s.
func clickText(t *testing.T, tt *ui.Tester, s string) {
	t.Helper()
	if err := tt.Click(s); err != nil {
		t.Fatalf("%v; texts %q", err, tt.Texts())
	}
	tt.Frame()
}

func settingsTester(a *App) *ui.Tester {
	tt := ui.NewTester(a.settingsView, 760, 560)
	tt.Frame()
	return tt
}

// TestSettingsFirstLaunch: with no settings file, the app runs on System
// and 11.6 pt, and Settings shows General with Appearance and one profile,
// Default, built in (criterion 1).
func TestSettingsFirstLaunch(t *testing.T) {
	a, m, s, _ := settingsApp(t, "")
	wantSize(t, "first launch", 11.6)
	if s.lastTheme() != mygo.ThemeSystem {
		t.Errorf("first launch applied theme %q, want system", s.lastTheme())
	}
	checked(t, m, "System")

	tt := settingsTester(a)
	for _, want := range []string{
		"General", "These apply to the whole app, whatever profile a pane uses.",
		"Appearance", "System", "Light", "Dark",
		"Default", "built-in",
		"Changes apply and save as you make them.",
	} {
		if !tt.HasText(want) {
			t.Errorf("Settings does not show %q; texts %q", want, tt.Texts())
		}
	}
	if !slices.ContainsFunc(tt.Texts(), func(s string) bool { return strings.EqualFold(s, "Profiles") }) {
		t.Errorf("Settings has no PROFILES heading; texts %q", tt.Texts())
	}
	if tt.HasText("New Profile") {
		t.Error("Settings shows New Profile…, which is delivery 2")
	}
}

// TestSettingsMenuItem: Settings… is GoRex ▸ Settings… (⌘,) on macOS and
// Edit ▸ Settings… (Ctrl+,) elsewhere, and the palette shows its shortcut
// (criterion 2, brief §5).
func TestSettingsMenuItem(t *testing.T) {
	a, m, _, _ := settingsApp(t, "")
	if runtime.GOOS == "darwin" {
		var app *mygo.MenuItem
		for _, it := range m.Items() {
			if it.Role == mygo.RoleAppMenu {
				app = it
			}
		}
		if app == nil {
			t.Fatal("no app menu")
		}
		settings := findItem(app.Submenu, "Settings…")
		if settings == nil || settings.Accelerator != "CmdOrCtrl+," {
			t.Fatalf("the app menu has no Settings… with CmdOrCtrl+,: %+v", settings)
		}
		// The app menu written out in full keeps its standard items.
		at := map[mygo.MenuRole]int{}
		for i, it := range app.Submenu {
			if it.Role != "" {
				at[it.Role] = i
			}
		}
		for _, r := range []mygo.MenuRole{mygo.RoleServices, mygo.RoleHide, mygo.RoleHideOthers, mygo.RoleUnhide, mygo.RoleQuit} {
			if _, ok := at[r]; !ok {
				t.Errorf("the app menu lost its %q item", r)
			}
		}
		// About is an item of its own, to show the commit of the build.
		about := findItem(app.Submenu, "About GoRex")
		if about == nil || about.Click == nil {
			t.Fatalf("the app menu has no About GoRex that does something: %+v", about)
		}
		// Choosing it opens the panel with this build's commit as its version.
		var got []mygo.AboutPanelOptions
		old := showAboutPanel
		showAboutPanel = func(o mygo.AboutPanelOptions) { got = append(got, o) }
		about.Click(about, nil)
		showAboutPanel = old
		if len(got) != 1 || got[0].Version != buildCommit() {
			t.Errorf("About GoRex opened the panel with %+v, want Version %q", got, buildCommit())
		}
		if i := slices.Index(app.Submenu, settings); i < slices.Index(app.Submenu, about) {
			t.Error("Settings… comes before About")
		}
	} else {
		edit := findItem(m.Items(), "Edit")
		if edit == nil {
			t.Fatal("no Edit menu")
		}
		if s := findItem(edit.Submenu, "Settings…"); s == nil || s.Accelerator != "Ctrl+," {
			t.Fatalf("Edit has no Settings… with Ctrl+,: %+v", s)
		}
	}
	if got := (shortcut{ui.Ctrl, ui.KeyComma}).String(); got != "Ctrl+," {
		t.Errorf("Ctrl and comma written %q, want \"Ctrl+,\"", got)
	}
	if cmdSettings.Other != (shortcut{ui.Ctrl, ui.KeyComma}) {
		t.Errorf("Settings… elsewhere than on macOS is %v, want Ctrl+,", cmdSettings.Other)
	}
	if !slices.Contains(commands, &cmdSettings) {
		t.Error("Settings… is not in the command table, so Windows' shortcut loop never sees Ctrl+,")
	}
	keys := "⌘,"
	if runtime.GOOS != "darwin" {
		keys = "Ctrl+,"
	}
	i := slices.IndexFunc(a.paletteItems(), func(p paletteItem) bool { return p.title == "Settings…" })
	if i < 0 {
		t.Fatal("the palette has no Settings…")
	}
	if got := a.paletteItems()[i].keys; got != keys {
		t.Errorf("the palette shows Settings… with %q, want %q", got, keys)
	}
}

// TestSettingsOpensOnce: the menu item opens Settings even with no main
// window and no focused window, and the menu, the palette and openSettings
// then bring that one window to the front: there is never a second
// (criteria 2 and 11).
func TestSettingsOpensOnce(t *testing.T) {
	a, m, s, _ := settingsApp(t, "")
	item := menuItem(t, m, "Settings…")
	click(t, item, nil)
	if len(s.created) != 1 {
		t.Fatalf("Settings… with no window made %d windows, want 1", len(s.created))
	}
	first := s.created[0]

	click(t, item, first)
	i := slices.IndexFunc(a.paletteItems(), func(p paletteItem) bool { return p.title == "Settings…" })
	if i < 0 {
		t.Fatal("the palette has no Settings…")
	}
	within(t, "the palette's Settings…", a.paletteItems()[i].run)
	within(t, "openSettings", a.openSettings)
	if len(s.created) != 1 {
		t.Errorf("opening Settings again made %d windows in all, want 1", len(s.created))
	}
	if want := []*mygo.Window{first, first, first}; !slices.Equal(s.focused, want) {
		t.Errorf("opening Settings again focused %v, want the first window 3 times %v", s.focused, want)
	}
}

// TestCommandsFromSettings: one helper decides which commands run for a
// menu click from a window. From the main window all of them run; from
// Settings or no window, only the app-wide ones, which also run with no
// main window (brief §4, criterion 10).
func TestCommandsFromSettings(t *testing.T) {
	a, _, _, _ := settingsApp(t, "")
	within(t, "openSettings", a.openSettings)
	settings, main := a.settingsWin, &mygo.Window{}
	t.Cleanup(func() { a.win = nil })
	appWide := []*command{&cmdSettings, &cmdNewTab, &cmdBigger, &cmdSmaller, &cmdActualSize, &cmdLight, &cmdDark, &cmdSystem, &cmdEndAll}
	for _, cmd := range commands {
		wide := slices.Contains(appWide, cmd)
		a.win = main
		if !a.runsIn(cmd, main) {
			t.Errorf("%s does not run from the main window", cmd.Title)
		}
		if got := a.runsIn(cmd, settings); got != wide {
			t.Errorf("%s from Settings: runs %v, want %v", cmd.Title, got, wide)
		}
		if got := a.runsIn(cmd, nil); got != wide {
			t.Errorf("%s from no window: runs %v, want %v", cmd.Title, got, wide)
		}
		// With no main window, Quit and End All Sessions does nothing in
		// quitAndEnd (TestEndAllNeedsMainWindow), whatever the helper says.
		a.win = nil
		if cmd == &cmdEndAll {
			continue
		}
		if got := a.runsIn(cmd, settings); got != wide {
			t.Errorf("%s from Settings, no main window: runs %v, want %v", cmd.Title, got, wide)
		}
		if got := a.runsIn(cmd, nil); got != wide {
			t.Errorf("%s from no window, no main window: runs %v, want %v", cmd.Title, got, wide)
		}
	}
}

// TestMenuClicksFromSettings: with Settings in front, the menu's text size
// runs, Close Pane closes Settings, and New Tab brings the main window to
// the front; from the main window, Close Pane leaves Settings open
// (criterion 10, brief §4).
func TestMenuClicksFromSettings(t *testing.T) {
	a, m, s, _ := settingsApp(t, "")
	within(t, "openSettings", a.openSettings)
	settings := a.settingsWin
	t.Cleanup(func() { a.win = nil })

	// No main window: app-wide commands run directly.
	click(t, menuItem(t, m, "Bigger Text"), settings)
	wantSize(t, "Bigger Text from Settings with no main window", 12.6)
	click(t, menuItem(t, m, "Bigger Text"), nil)
	wantSize(t, "Bigger Text from no window", 13.6)

	click(t, menuItem(t, m, "Close Pane"), settings)
	if !slices.Equal(s.closed, []*mygo.Window{settings}) {
		t.Errorf("Close Pane from Settings closed %v, want Settings %v", s.closed, settings)
	}

	main := &mygo.Window{}
	a.win = main
	s.closed, s.focused = nil, nil
	click(t, menuItem(t, m, "New Tab"), settings)
	if !slices.Contains(s.focused, main) {
		t.Errorf("New Tab from Settings focused %v, not the main window", s.focused)
	}
	click(t, menuItem(t, m, "Close Pane"), main)
	if len(s.closed) != 0 {
		t.Errorf("Close Pane from the main window closed %v", s.closed)
	}
}

// TestEndAllNeedsMainWindow: with no main window, Quit and End All
// Sessions does nothing, and the sessions go on (brief §4).
func TestEndAllNeedsMainWindow(t *testing.T) {
	a, _ := newTestApp(t)
	if a.win != nil {
		t.Fatal("the test app has a main window")
	}
	within(t, "Quit and End All Sessions", func() { cmdEndAll.Run(a) })
	time.Sleep(500 * time.Millisecond) // the server ends 50 ms after it is asked
	if a.quitting {
		t.Error("Quit and End All Sessions with no main window set quitting")
	}
	// A server that shuts down ends every session and removes its socket.
	if _, err := os.Stat(rex.SocketPath()); err != nil {
		t.Errorf("the session server shut down, ending the sessions: %v", err)
	}
}

// TestSettingsAppearanceInStep: Settings, the palette and View ▸
// Appearance change the theme, save it, and leave the menu showing it
// (criterion 3, brief §3).
func TestSettingsAppearanceInStep(t *testing.T) {
	a, m, s, dir := settingsApp(t, "")
	tt := settingsTester(a)

	clickText(t, tt, "Dark")
	if s.lastTheme() != mygo.ThemeDark {
		t.Errorf("choosing Dark in Settings applied %q", s.lastTheme())
	}
	checked(t, m, "Dark")
	checkSaved(t, dir, "dark", 11.6)

	within(t, "the palette's Appearance: Light", func() { cmdLight.Run(a) })
	if s.lastTheme() != mygo.ThemeLight {
		t.Errorf("Appearance: Light applied %q", s.lastTheme())
	}
	checked(t, m, "Light")

	click(t, appearanceItem(t, m, "System"), nil)
	if s.lastTheme() != mygo.ThemeSystem {
		t.Errorf("View ▸ Appearance ▸ System applied %q", s.lastTheme())
	}
	checked(t, m, "System")
	checkSaved(t, dir, "system", 11.6)
}

// TestSettingsTextSize: Settings' − and + and slider go to whole points
// and save, the value shows as it is, and Settings follows ⌘+ and ⌘0
// (criterion 4, brief §2 "Default page").
func TestSettingsTextSize(t *testing.T) {
	a, _, _, dir := settingsApp(t, "")
	tt := settingsTester(a)
	clickText(t, tt, "Default")
	for _, want := range []string{"Text size", "11.6 pt"} {
		if !tt.HasText(want) {
			t.Fatalf("Default's page does not show %q; texts %q", want, tt.Texts())
		}
	}

	clickText(t, tt, "Smaller Text")
	wantSize(t, "− from 11.6", 11)
	checkSaved(t, dir, "system", 11)
	if !tt.HasText("11 pt") {
		t.Errorf("after −, Settings does not show 11 pt; texts %q", tt.Texts())
	}

	within(t, "Actual Size", func() { cmdActualSize.Run(a) })
	tt.Frame()
	if !tt.HasText("11.6 pt") {
		t.Errorf("after ⌘0, Settings does not show 11.6 pt; texts %q", tt.Texts())
	}

	clickText(t, tt, "Bigger Text")
	wantSize(t, "+ from 11.6", 12)
	checkSaved(t, dir, "system", 12)

	within(t, "Bigger Text", func() { cmdBigger.Run(a) })
	tt.Frame()
	if !tt.HasText("13 pt") {
		t.Errorf("after ⌘+, Settings does not show 13 pt; texts %q", tt.Texts())
	}

	// The slider lies between − and +: a click halfway along it, near
	// 20 pt of 8–32, lands on a whole point.
	lo, _ := tt.Find("Smaller Text")
	hi, _ := tt.Find("Bigger Text")
	tt.ClickAt((lo.X+lo.W+hi.X)/2, lo.Y+lo.H/2)
	tt.Frame()
	size := float64(termFont.Size)
	if size != math.Round(size) || size < 14 || size > 26 {
		t.Errorf("a click halfway along the slider gave %v pt, want a whole point near 20", size)
	}
	checkSaved(t, dir, "system", size)
}

// TestSettingsMigratesV1: a v1 file becomes v2 as it loads, keeping the
// appearance ("" is System) and the text size, clamped to 8–32 or 11.6
// when missing, with the original kept as settings.v1.json (criteria 5
// and 6).
func TestSettingsMigratesV1(t *testing.T) {
	for _, tc := range []struct {
		name, file, appearance string
		theme                  mygo.ThemeSource
		size                   float64
	}{
		{"dark", `{"appearance":"dark","fontSize":14}`, "dark", mygo.ThemeDark, 14},
		{"system", `{"fontSize":13}`, "system", mygo.ThemeSystem, 13},
		{"no size", `{"appearance":"light"}`, "light", mygo.ThemeLight, 11.6},
		{"too big", `{"appearance":"light","fontSize":50}`, "light", mygo.ThemeLight, 32},
		{"too small", `{"fontSize":7}`, "system", mygo.ThemeSystem, 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, s, dir := settingsApp(t, tc.file)
			wantSize(t, "after loading "+tc.file, tc.size)
			if s.lastTheme() != tc.theme {
				t.Errorf("loading %s applied %q, want %q", tc.file, s.lastTheme(), tc.theme)
			}
			checkSaved(t, dir, tc.appearance, tc.size)
			b, err := os.ReadFile(filepath.Join(dir, "settings.v1.json"))
			if err != nil || string(b) != tc.file {
				t.Errorf("settings.v1.json holds %q (%v), want the original %q", b, err, tc.file)
			}
		})
	}
}

// TestSettingsMigrationOverwritesOldBackup: a backup left by an earlier
// attempt is replaced by the file migrated now (brief §1).
func TestSettingsMigrationOverwritesOldBackup(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.v1.json"), []byte(`{"old":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOREX_DIR", dir)
	file := `{"appearance":"dark","fontSize":14}`
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte(file), 0o600)
	installSeams(t)
	within(t, "loadSettings", loadSettings)
	if b, _ := os.ReadFile(filepath.Join(dir, "settings.v1.json")); string(b) != file {
		t.Errorf("settings.v1.json holds %q, want %q", b, file)
	}
	checkSaved(t, dir, "dark", 14)
}

// TestSettingsV2Load: a v2 file's text size is clamped, and a missing or
// renamed Default is restored (criteria 5 and 7).
func TestSettingsV2Load(t *testing.T) {
	for _, tc := range []struct {
		name, file, appearance string
		size                   float64
	}{
		{"too big", `{"version":2,"appearance":"dark","profiles":{"default":{"name":"Default","fontSize":100}}}`, "dark", 32},
		{"no profiles", `{"version":2,"appearance":"light"}`, "light", 11.6},
		{"renamed", `{"version":2,"appearance":"light","profiles":{"default":{"name":"Mine","fontSize":15}}}`, "light", 15},
		{"no size", `{"version":2,"appearance":"system","profiles":{"default":{"name":"Default"}}}`, "system", 11.6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _, _, dir := settingsApp(t, tc.file)
			wantSize(t, "after loading "+tc.file, tc.size)
			within(t, "setAppearance", func() { a.setAppearance(tc.appearance) })
			checkSaved(t, dir, tc.appearance, tc.size)
		})
	}
}

// TestSettingsSurviveRestart: what is changed is what the next launch
// loads (criterion 8).
func TestSettingsSurviveRestart(t *testing.T) {
	a, _, s, dir := settingsApp(t, "")
	within(t, "setFontSize", func() { a.setFontSize(15) })
	within(t, "setAppearance", func() { a.setAppearance("dark") })
	checkSaved(t, dir, "dark", 15)

	termFont.Size, s.themes = 0, nil
	within(t, "loadSettings", loadSettings)
	wantSize(t, "the next launch", 15)
	if s.lastTheme() != mygo.ThemeDark {
		t.Errorf("the next launch applied %q, want dark", s.lastTheme())
	}
}

// TestSettingsReadOnly: a file newer than v2, a file that cannot be read,
// and a v1 file whose backup cannot be written are never overwritten, even
// by a change later in the run; the change still applies, and Settings
// says it is not saved, and why (criterion 8, brief §6 "Write gate").
func TestSettingsReadOnly(t *testing.T) {
	for _, tc := range []struct {
		name, file string
		// setup prepares the directory before loading.
		setup  func(t *testing.T, dir string)
		size   float64 // the text size loaded, or 0 for any
		reason string
	}{
		{
			name:   "newer",
			file:   `{"version":3,"appearance":"dark","profiles":{"default":{"name":"Default","fontSize":20}}}`,
			reason: "settings.json is newer than this version of GoRex",
		},
		{
			name:   "invalid",
			file:   `{"appearance": "dark", `,
			size:   11.6,
			reason: "settings.json could not be read",
		},
		{
			name: "backup fails",
			file: `{"appearance":"dark","fontSize":14}`,
			// A directory, not empty, where the backup goes.
			setup: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "settings.v1.json", "x"), 0o700); err != nil {
					t.Fatal(err)
				}
			},
			size:   14,
			reason: "the settings backup could not be written",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("GOREX_DIR", dir)
			os.WriteFile(filepath.Join(dir, "settings.json"), []byte(tc.file), 0o600)
			if tc.setup != nil {
				tc.setup(t, dir)
			}
			s := installSeams(t)
			registerFonts()
			termFont.Size = 0
			within(t, "loadSettings", loadSettings)
			if tc.size != 0 {
				wantSize(t, "after loading", tc.size)
			}
			a := &App{}
			within(t, "App.menu", func() { a.menu() })
			within(t, "setFontSize", func() { a.setFontSize(17) })
			within(t, "setAppearance", func() { a.setAppearance("light") })
			wantSize(t, "a change in a read-only run", 17)
			if s.lastTheme() != mygo.ThemeLight {
				t.Errorf("a change in a read-only run applied %q, want light", s.lastTheme())
			}
			if b, _ := os.ReadFile(filepath.Join(dir, "settings.json")); !bytes.Equal(b, []byte(tc.file)) {
				t.Errorf("settings.json was overwritten:\n%s", b)
			}
			if tc.setup == nil {
				if _, err := os.Stat(filepath.Join(dir, "settings.v1.json")); err == nil {
					t.Error("a backup was written of a file that is not v1")
				}
			}

			tt := settingsTester(a)
			if !tt.HasText("Changes apply but are not saved") || !tt.HasText(tc.reason) {
				t.Errorf("Settings does not say changes are not saved because %q; texts %q", tc.reason, tt.Texts())
			}
			if tt.HasText("Changes apply and save as you make them.") {
				t.Error("Settings says changes are saved in a read-only run")
			}
		})
	}
}

// TestSettingsWriteFails: when settings.json cannot be written during a
// run, the change still applies, and Settings says it was not saved rather
// than that it was.
func TestSettingsWriteFails(t *testing.T) {
	a, _, _, dir := settingsApp(t, "")
	if readOnly != "" {
		t.Fatalf("a first launch is read-only: %q", readOnly)
	}
	// A directory, not empty, where settings.json goes.
	if err := os.MkdirAll(filepath.Join(dir, "settings.json", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	within(t, "setFontSize", func() { a.setFontSize(17) })
	wantSize(t, "a change that cannot be written", 17)
	if readOnly != readOnlyWrite {
		t.Errorf("readOnly = %q after a failed write, want %q", readOnly, readOnlyWrite)
	}
	tt := settingsTester(a)
	if !tt.HasText("Changes apply but are not saved") || !tt.HasText(readOnlyWrite) {
		t.Errorf("Settings does not say the change was not saved; texts %q", tt.Texts())
	}
	if tt.HasText("Changes apply and save as you make them.") {
		t.Error("Settings says changes are saved after a failed write")
	}
}

// TestSettingsKeysOnWindows: on Windows, where native-UI windows run no
// menu shortcuts, Settings closes on Esc, Ctrl+W and Ctrl+Shift+W, keeps
// one window on Ctrl+, and runs Ctrl+= (criteria 9 and 10).
func TestSettingsKeysOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Settings handles these keys itself only on Windows")
	}
	a, _, s, _ := settingsApp(t, "")
	within(t, "openSettings", a.openSettings)
	settings := a.settingsWin
	tt := settingsTester(a)
	for _, k := range []struct {
		name string
		mods ui.Modifiers
		key  ui.Key
	}{{"Esc", 0, ui.KeyEscape}, {"Ctrl+W", ui.Ctrl, ui.KeyW}, {"Ctrl+Shift+W", ui.Ctrl | ui.Shift, ui.KeyW}} {
		s.closed = nil
		tt.Key(k.mods, k.key)
		tt.Frame()
		tt.Frame()
		if !slices.Equal(s.closed, []*mygo.Window{settings}) {
			t.Errorf("%s in Settings closed %v, want Settings", k.name, s.closed)
		}
	}
	tt.Key(ui.Ctrl, ui.KeyComma)
	tt.Frame()
	tt.Frame()
	if len(s.created) != 1 || !slices.Contains(s.focused, settings) {
		t.Errorf("Ctrl+, in Settings: %d windows made, focused %v; want 1 made and Settings focused", len(s.created), s.focused)
	}
	tt.Key(ui.Ctrl, ui.KeyEqual)
	tt.Frame()
	tt.Frame()
	wantSize(t, "Ctrl+= in Settings", 12.6)
}

// TestReadmeSettings: README lists Ctrl+, among the shortcuts, names it
// with the text-size keys as an exception to the Shift rule, and lists the
// Settings window's source file in Layout (criterion 13).
func TestReadmeSettings(t *testing.T) {
	b, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(b)
	lines := strings.Split(readme, "\n")
	if !slices.ContainsFunc(lines, func(l string) bool {
		return strings.HasPrefix(l, "|") && strings.Contains(l, "⌘,") && strings.Contains(l, "Ctrl+,")
	}) {
		t.Error("the Shortcuts table has no row with ⌘, and Ctrl+,")
	}
	// The prose between the Shortcuts table and the next section.
	end := strings.Index(readme, "## Develop")
	if end < 0 {
		t.Fatal("README has no Develop section after its shortcuts")
	}
	table := strings.Index(readme, "| ⌘T")
	if table < 0 || table > end {
		t.Fatal("README's Shortcuts table is not where it was")
	}
	var rule []string
	for _, l := range strings.Split(readme[table:end], "\n") {
		if !strings.HasPrefix(l, "|") {
			rule = append(rule, l)
		}
	}
	if r := strings.Join(rule, "\n"); !strings.Contains(r, "Ctrl+,") || !strings.Contains(r, "Ctrl+=") {
		t.Errorf("the Shift rule does not name Ctrl+, with Ctrl+=:\n%s", r)
	}

	// The file that declares settingsView.
	decl := regexp.MustCompile(`(?m)^func \([a-z]+ \*App\) settingsView\(`)
	names, _ := filepath.Glob("*.go")
	var file string
	for _, name := range names {
		if b, err := os.ReadFile(name); err == nil && !strings.HasSuffix(name, "_test.go") && decl.Match(b) {
			file = name
		}
	}
	if file == "" {
		t.Fatal("no file declares settingsView")
	}
	if !regexp.MustCompile(`^[a-z]+\.go$`).MatchString(file) || file == "settings.go" {
		t.Errorf("settingsView is in %s, not a new file with a one-word lowercase name", file)
	}
	i := strings.Index(readme, "## Layout")
	if i < 0 || !strings.Contains(readme[i:], "`"+file+"`") {
		t.Errorf("README's Layout table does not list %s", file)
	}
}
