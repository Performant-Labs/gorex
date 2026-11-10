package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"

	"gorex/internal/rex"
)

// settings are the user's choices, kept in settings.json beside the
// server's state. Version 2 of the file:
//
//	{
//	  "version": 2,
//	  "appearance": "system",
//	  "profiles": {
//	    "default": { "name": "Default", "fontSize": 11.6 }
//	  }
//	}
//
// Version 1, a file without "version", held "appearance" ("" for the
// system's) and "fontSize" alone.
type settings struct {
	Version int `json:"version"`
	// Appearance is the whole app's: "system", "light" or "dark".
	Appearance string `json:"appearance"`
	// Profiles are how the panes' terminals look, by id. Default, the one
	// built in, is always there and stores every value.
	Profiles map[string]*profile `json:"profiles"`
}

// profile is a set of the terminals' settings. A value left nil is unset:
// a profile other than Default stores only what it changes from Default,
// and inherits the rest.
type profile struct {
	Name     string   `json:"name"`
	FontSize *float32 `json:"fontSize,omitempty"`
}

const (
	settingsVersion  = 2
	defaultProfileID = "default"
	defaultName      = "Default"
	defaultFontSize  = 11.6
	minFontSize      = 8
	maxFontSize      = 32
)

// prefs are the settings the app runs on.
var prefs = defaultSettings()

// readOnly tells why settings.json is not written in this run, as the
// Settings window says, or is "" when changes are saved.
var readOnly string

// The reasons a run is read-only.
const (
	readOnlyNewer      = "settings.json is newer than this version of GoRex"
	readOnlyUnreadable = "settings.json could not be read"
	readOnlyBackup     = "the settings backup could not be written"
)

// errNewer is parseSettings' error for a file of a version after this one.
var errNewer = errors.New("settings.json is of a newer version")

func settingsPath() string   { return filepath.Join(rex.Dir(), "settings.json") }
func settingsV1Path() string { return filepath.Join(rex.Dir(), "settings.v1.json") }

func defaultSettings() settings {
	s := settings{}
	s.complete()
	return s
}

// complete makes s a whole version 2: Default is restored with what it
// lacks and named Default again, the text size is clamped, and the
// appearance is one of the three.
func (s *settings) complete() {
	s.Version = settingsVersion
	switch s.Appearance {
	case "light", "dark":
	default:
		s.Appearance = "system"
	}
	if s.Profiles == nil {
		s.Profiles = map[string]*profile{}
	}
	d := s.Profiles[defaultProfileID]
	if d == nil {
		d = &profile{}
		s.Profiles[defaultProfileID] = d
	}
	d.Name = defaultName
	size := float32(defaultFontSize)
	if d.FontSize != nil {
		size = clampFontSize(*d.FontSize)
	}
	d.FontSize = &size
}

// fontSize is Default's text size.
func (s *settings) fontSize() float32 { return *s.Profiles[defaultProfileID].FontSize }

func clampFontSize(size float32) float32 { return min(max(size, minFontSize), maxFontSize) }

// parseSettings reads settings.json, of version 1 or 2, as version 2. It
// reports whether the file was of version 1, and errNewer for a version
// after 2, read as well as it can be. A file that is not JSON gives the
// defaults and its error.
func parseSettings(b []byte) (s settings, v1 bool, err error) {
	var f struct {
		settings
		// FontSize is version 1's text size. JSON's missing number reads
		// as 0, so 0 and less are no text size.
		FontSize float32 `json:"fontSize"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return defaultSettings(), false, err
	}
	s = f.settings
	if s.Version < settingsVersion {
		v1 = true
		s.Profiles = map[string]*profile{defaultProfileID: {}}
		if f.FontSize > 0 {
			size := f.FontSize
			s.Profiles[defaultProfileID].FontSize = &size
		}
	} else if s.Version > settingsVersion {
		err = errNewer
	}
	s.complete()
	return s, v1, err
}

// loadSettings reads settings.json and applies it. A file of version 1 is
// copied byte for byte to settings.v1.json, then written anew as version
// 2; if the copy cannot be written, the file is left as it is. A file that
// cannot be read, or is newer than this version of GoRex, is not written
// in this run.
func loadSettings() {
	prefs, readOnly = defaultSettings(), ""
	b, err := os.ReadFile(settingsPath())
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		readOnly = readOnlyUnreadable
	default:
		s, v1, err := parseSettings(b)
		prefs = s
		switch {
		case errors.Is(err, errNewer):
			readOnly = readOnlyNewer
		case err != nil:
			readOnly = readOnlyUnreadable
		case v1:
			if os.WriteFile(settingsV1Path(), b, 0o600) != nil {
				readOnly = readOnlyBackup
			} else {
				saveSettings()
			}
		}
	}
	termFont.Size = prefs.fontSize()
	applyAppearance()
}

// saveSettings writes the settings, unless the run is read-only.
func saveSettings() {
	if readOnly != "" {
		return
	}
	b, _ := json.MarshalIndent(prefs, "", "  ")
	os.MkdirAll(rex.Dir(), 0o700)
	os.WriteFile(settingsPath(), b, 0o600)
}

// setThemeSource applies the app's appearance: mygo.Theme.SetSource, which
// waits on the main loop, and which tests replace.
var setThemeSource = mygo.Theme.SetSource

func applyAppearance() {
	switch prefs.Appearance {
	case "light":
		setThemeSource(mygo.ThemeLight)
	case "dark":
		setThemeSource(mygo.ThemeDark)
	default:
		setThemeSource(mygo.ThemeSystem)
	}
}

// setAppearance changes the whole app's appearance: "system", "light" or
// "dark". It and setFontSize are the only writers of the settings.
func (a *App) setAppearance(v string) {
	prefs.Appearance = v
	prefs.complete()
	applyAppearance()
	saveSettings()
	a.settingsChanged()
}

// setFontSize changes Default's text size, in every pane.
func (a *App) setFontSize(size float32) {
	size = clampFontSize(size)
	prefs.Profiles[defaultProfileID].FontSize = &size
	termFont.Size = size
	for _, t := range a.tabs {
		for _, p := range t.panes() {
			if p.term != nil {
				p.term.SetFont(termFont)
			}
		}
	}
	if a.preview != nil {
		a.preview.SetFont(termFont)
	}
	saveSettings()
	a.settingsChanged()
}

// settingsChanged shows a change of the settings where they show apart
// from the main window, which a.do draws: View ▸ Appearance's choice and
// the Settings window.
func (a *App) settingsChanged() {
	for v, it := range a.appearanceItems {
		it.SetChecked(prefs.Appearance == v)
	}
	if a.settingsWin != nil {
		a.settingsWin.Invalidate()
	}
}
