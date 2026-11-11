package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Pasting an image: a screenshot on the clipboard (⌃⌘⇧4, or Copy Image in a
// browser) is written to a file and goes the way a dropped file does, so an
// agent over ssh gets it too: neither Claude Code nor opencode can read the
// clipboard of this computer from the other end of an ssh.
//
// Text wins: when the clipboard has any, it is pasted as text, as before. Files
// copied in the Finder keep that too, as their names are text.
//
// Which key does it depends on the platform. On macOS the Edit menu's Paste
// owns Command+V: AppKit gives the key to the menu, which sends paste: to the
// terminal (text only), so the key never reaches the pane's shortcut and a
// role item's Click is not called. Edit ▸ Paste Image (Shift+Command+V) is an
// item of its own, so it runs. Elsewhere Control+Shift+V is the terminal's own
// paste key and reaches the pane's shortcut.

var (
	// clipboardImage returns the image on the clipboard as PNG, nil when it
	// holds none. It is a variable for tests.
	clipboardImage = func() []byte { return mygo.Clipboard.ReadImage() }
	// pasteDir is where pasted images are written, a variable for tests.
	pasteDir = func() (string, error) {
		base, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(base, "GoRex", "pastes"), nil
	}
)

// pasteKeepTime is how long a pasted image stays on this computer. A local
// pane's agent reads it by its path, so it must outlive the paste.
const pasteKeepTime = remoteDropHours * time.Hour

// pasteKey is the pane's paste shortcut, Control+Shift+V as the terminal's own,
// and none on macOS, where Command+V is the Edit menu's and never reaches the
// pane: a shortcut there would be dead, and would paste twice if it ever ran
// alongside the role.
func pasteKey() (mods ui.Modifiers, key ui.Key, ok bool) {
	if runtime.GOOS == "darwin" {
		return 0, 0, false
	}
	return ui.Ctrl | ui.Shift, ui.KeyV, true
}

// pasteClipboard pastes into the pane what the clipboard holds: its text, or,
// when it has none, its image.
func (a *App) pasteClipboard(p *Pane, text string) {
	if p == nil || p.term == nil || p.closed {
		return
	}
	if text != "" {
		p.term.Paste(text)
		return
	}
	a.pasteImage(p, false)
}

// pasteImage pastes the image on the clipboard into the pane, as a dropped
// file. When there is none nothing happens, unless asked for (loud), which
// says so in the pane's header.
func (a *App) pasteImage(p *Pane, loud bool) {
	if p == nil || p.term == nil || p.closed {
		return
	}
	png := clipboardImage()
	if len(png) == 0 {
		if loud {
			p.setDropNote(a, "No image on the clipboard", true, dropMessageTime)
		}
		return
	}
	if len(png) > maxDropSize {
		p.setDropNote(a, fmt.Sprintf("Paste failed: the image is larger than %d MB", maxDropSize>>20), true, dropMessageTime)
		return
	}
	path, err := savePaste(png, time.Now())
	if err != nil {
		p.setDropNote(a, "Paste failed: "+err.Error(), true, dropMessageTime)
		return
	}
	a.dropFilesFrom(p, []string{path}, true)
}

// pasteImageFocused is Edit ▸ Paste Image: the clipboard's image into the
// focused pane. A field of the app that has the focus, as the palette's, is
// not a pane to paste into.
func (a *App) pasteImageFocused() {
	if a.paletteOpen || a.renaming != nil {
		return
	}
	if t := a.tab(); t != nil {
		a.pasteImage(t.Focus, true)
	}
}

// savePaste writes the image to a file of its own, readable by its owner
// only, and removes the pasted images older than pasteKeepTime.
func savePaste(png []byte, now time.Time) (string, error) {
	dir, err := pasteDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if old, _ := filepath.Glob(filepath.Join(dir, "paste-*.png")); old != nil {
		for _, f := range old {
			if st, err := os.Stat(f); err == nil && now.Sub(st.ModTime()) > pasteKeepTime {
				os.Remove(f)
			}
		}
	}
	f, err := os.CreateTemp(dir, "paste-*.png")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(png); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}
