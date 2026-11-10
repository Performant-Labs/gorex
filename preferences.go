package main

import (
	"fmt"
	"math"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
)

// The Settings window's native calls, which tests replace: under go test
// they would wait on a main loop that never runs.
var (
	// newSettingsWindow opens the Settings window: settingsWindow, set in
	// init, as the window's view runs the commands, one of which opens it.
	newSettingsWindow func(a *App) *mygo.Window
	focusWindow       = (*mygo.Window).Focus
	closeWindow       = (*mygo.Window).Close
)

func init() { newSettingsWindow = settingsWindow }

// settingsWindow opens the Settings window, which forgets itself when it
// closes.
func settingsWindow(a *App) *mygo.Window {
	win := mygo.NewWindow(mygo.WindowOptions{
		Title:             "Settings",
		Width:             760,
		Height:            560,
		MinWidth:          640,
		MinHeight:         460,
		StateKey:          "settings",
		DisableFullScreen: true,
		BackgroundColor:   "light-dark(#fcfcfd, #26262a)",
		Content:           ui.View(a.settingsView),
	})
	win.OnClosed(func() {
		if a.settingsWin == win {
			a.settingsWin = nil
		}
	})
	return win
}

// openSettings opens the Settings window, or brings it to the front:
// there is never a second.
func (a *App) openSettings() {
	if a.settingsWin != nil {
		focusWindow(a.settingsWin)
		return
	}
	a.settingsWin = newSettingsWindow(a)
}

// settingsView draws the Settings window: General, and the profiles,
// Default alone for now. Its controls change the settings through
// setAppearance and setFontSize, which save them and show them in the
// menus, and in this window again.
func (a *App) settingsView(c *ui.Context) {
	a.settingsKeys(c)
	k := colorsOf(c)
	c.Root().Background(k.panel)
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Width(220).Shrink(0).Padding(14, 10).Gap(2).Background(k.hover).Children(func() {
			a.settingsTab(c, k, "", "General", "")
			ui.Text(c, "PROFILES").FontSize(11).FontWeight(600).TextColor(k.textFaint).Padding(18, 12, 6, 12)
			a.settingsTab(c, k, defaultProfileID, defaultName, "built-in")
		})
		ui.Box(c).Width(1).Shrink(0).Background(k.panelBorder)
		ui.Column(c).Grow(1).MinWidth(0).Padding(26, 34, 20, 34).Children(func() {
			if a.settingsPage == defaultProfileID {
				a.defaultPage(c, k)
			} else {
				a.generalPage(c, k)
			}
			ui.Spacer(c)
			a.settingsFooter(c, k)
		})
	})
}

// settingsKeys handles the keys on Windows, where windows of native UI run
// no menu shortcuts: Esc, Ctrl+W and Ctrl+Shift+W (Close Pane's) close
// Settings, and the app-wide commands' shortcuts run as from the menus.
func (a *App) settingsKeys(c *ui.Context) {
	if runtime.GOOS != "windows" {
		return
	}
	if c.Shortcut(0, ui.KeyEscape) || c.Shortcut(ui.Ctrl, ui.KeyW) || c.Shortcut(ui.Ctrl|ui.Shift, ui.KeyW) {
		if a.settingsWin != nil {
			closeWindow(a.settingsWin)
		}
		return
	}
	for _, cmd := range commands {
		if cmd.AppWide && cmd.Other != (shortcut{}) && c.Shortcut(cmd.Other.Mods, cmd.Other.Key) {
			a.run(cmd)
		}
	}
}

// settingsTab is an entry of the sidebar, showing page when chosen.
func (a *App) settingsTab(c *ui.Context, k *colors, page, name, note string) {
	row := ui.Row(c.Key("page "+page)).Height(32).Padding(0, 12).Gap(10).Radius(8).AlignItems(ui.Center).Cursor(ui.CursorPointer)
	if a.settingsPage == page {
		row.Background(k.pressed)
	} else if row.Hovered() {
		row.Background(k.hover)
	}
	if row.Clicked() {
		a.settingsPage = page
	}
	row.Children(func() {
		ui.Text(c, name).FontSize(13.5).TextColor(k.text)
		if note != "" {
			ui.Text(c, note).FontSize(12).TextColor(k.textFaint)
		}
	})
}

// settingsTitle heads a page, over a rule.
func settingsTitle(c *ui.Context, k *colors, title, line string) {
	ui.Text(c, title).FontSize(22).FontWeight(700).TextColor(k.text)
	if line != "" {
		ui.Text(c, line).FontSize(13).TextColor(k.textMuted).Margin(6, 0, 0, 0)
	}
	ui.Box(c).Height(1).Background(k.panelBorder).Margin(16, 0, 18, 0)
}

// generalPage shows the settings of the whole app.
func (a *App) generalPage(c *ui.Context, k *colors) {
	settingsTitle(c, k, "General", "These apply to the whole app, whatever profile a pane uses.")
	ui.Text(c, "Appearance").FontSize(13.5).FontWeight(600).TextColor(k.text)
	ui.Row(c).Gap(14).Margin(12, 0, 0, 0).Children(func() {
		for _, choice := range []struct{ value, label string }{{"system", "System"}, {"light", "Light"}, {"dark", "Dark"}} {
			a.appearanceChoice(c, k, choice.value, choice.label)
		}
	})
}

// appearanceChoice is a choice of appearance: a swatch of it, and its
// name.
func (a *App) appearanceChoice(c *ui.Context, k *colors, value, label string) {
	selected := prefs.Appearance
	r := ui.RadioBase(c.Key(value), &selected, value).Cursor(ui.CursorPointer)
	if r.Changed() && selected == value && prefs.Appearance != value {
		a.setAppearance(value)
	}
	on := prefs.Appearance == value
	r.Children(func() {
		ui.Column(c).Gap(8).AlignItems(ui.Center).Children(func() {
			sw := ui.Box(c).Size(120, 76).Radius(10).Background(k.hover)
			if on {
				sw.Border(2.5, c.Theme().Accent)
			} else {
				sw.Border(1, k.panelBorder)
			}
			sw.Draw(func(p *ui.Painter, r ui.Rect) {
				in := ui.Rect{X: r.X + 10, Y: r.Y + 10, W: r.W - 20, H: r.H - 20}
				switch value {
				case "light":
					p.Fill(in, lightTerm.Background, 6)
				case "dark":
					p.Fill(in, darkTerm.Background, 6)
				default: // half of each
					p.FillGradient(in, ui.LinearGradient{From: lightTerm.Background, To: darkTerm.Background, Angle: 90, Start: 0.495, End: 0.505}, 6)
				}
			})
			ui.Text(c, label).FontSize(13).TextColor(k.text)
		})
	})
}

// defaultPage shows Default's settings, and a preview of them.
func (a *App) defaultPage(c *ui.Context, k *colors) {
	settingsTitle(c, k, defaultName, "")
	size := prefs.fontSize()
	ui.Text(c, "Text size").FontSize(13.5).FontWeight(600).TextColor(k.text)
	// − and + and the slider go to whole points; a size that is not one
	// shows as it is, between them.
	ui.Row(c).Gap(12).AlignItems(ui.Center).Margin(10, 0, 0, 0).Children(func() {
		if iconButton(c, k, "minus", "Smaller Text", 30, 16).Border(1, k.panelBorder).Tooltip("Smaller Text").Clicked() {
			a.setFontSize(float32(math.Ceil(float64(size))) - 1)
		}
		v := float64(size)
		if ui.Slider(c, &v, minFontSize, maxFontSize).Step(1).Grow(1).MinWidth(0).Label("Text size").Changed() {
			a.setFontSize(float32(v))
		}
		if iconButton(c, k, "plus", "Bigger Text", 30, 16).Border(1, k.panelBorder).Tooltip("Bigger Text").Clicked() {
			a.setFontSize(float32(math.Floor(float64(size))) + 1)
		}
		ui.Text(c, fmt.Sprintf("%v pt", prefs.fontSize())).FontSize(13.5).TextColor(k.text).MinWidth(64).Shrink(0)
	})
	ui.Text(c, "PREVIEW · in the app's current appearance").FontSize(11).FontWeight(600).TextColor(k.textFaint).Margin(22, 0, 8, 0)
	if t := a.previewTerminal(); t != nil {
		paper := lightTerm.Background
		if c.Theme().Dark {
			paper = darkTerm.Background
		}
		ui.Box(c).Padding(10, 14).Radius(10).Clip().Background(paper).Border(1, k.panelBorder).Children(func() {
			// Room for a line more than it shows, so that none scrolls away.
			terminal.View(c, t).Height(prefs.fontSize() * 1.4 * 4)
		})
	}
}

// previewText is what the preview shows: a few lines of a shell.
const previewText = "\x1b[?25l~/Projects/gorex \x1b[32mmain\x1b[0m\r\n$ go test ./...\r\nok   gorex  0.42s"

// previewTerminal is the preview's terminal, made the first time it shows
// and kept: one that only shows what it is fed, whose reader never ends.
func (a *App) previewTerminal() *terminal.Terminal {
	if a.preview == nil {
		t, err := terminal.New(terminal.Options{Conn: nopConn{}, Transparent: true, Font: termFont, Theme: lightTerm, DarkTheme: darkTerm})
		if err != nil {
			return nil
		}
		t.Resize(60, 4)
		t.Feed([]byte(previewText))
		a.preview = t
	}
	return a.preview
}

// settingsFooter says that changes save as they are made, or why they do
// not.
func (a *App) settingsFooter(c *ui.Context, k *colors) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Margin(16, 0, 0, 0).Children(func() {
		if readOnly == "" {
			ui.Text(c, "Changes apply and save as you make them.").FontSize(12).TextColor(k.textFaint)
			return
		}
		ui.Box(c).Size(8, 8).Radius(4).Shrink(0).Background(k.attention)
		ui.Text(c, "Changes apply but are not saved: "+readOnly).FontSize(12).TextColor(k.textMuted)
	})
}
