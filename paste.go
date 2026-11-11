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

// pasteKey is the shortcut of paste: Command+V on macOS, and elsewhere
// Control+Shift+V, as the terminal's own.
func pasteKey() (ui.Modifiers, ui.Key) {
	if runtime.GOOS == "darwin" {
		return ui.Super, ui.KeyV
	}
	return ui.Ctrl | ui.Shift, ui.KeyV
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
	a.pasteImage(p)
}

// pasteImage pastes the image on the clipboard into the pane, as a dropped
// file; nothing happens when there is none.
func (a *App) pasteImage(p *Pane) {
	if p == nil || p.term == nil || p.closed {
		return
	}
	png := clipboardImage()
	if len(png) == 0 {
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

// menuPaste runs after Edit ▸ Paste has done what its role does, which is
// the text: it pastes the image where there is no text. win is the window
// the menu was chosen in.
func (a *App) menuPaste(win *mygo.Window) {
	if win == nil || win != a.win {
		return
	}
	a.pasteAfterRole(a.getServices().ReadClipboard())
}

// pasteAfterRole is menuPaste once the window is known to be the main one.
func (a *App) pasteAfterRole(text string) {
	if text != "" || a.paletteOpen || a.renaming != nil {
		return // the role pasted the text, or a field has the focus
	}
	if t := a.tab(); t != nil {
		a.pasteImage(t.Focus)
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
