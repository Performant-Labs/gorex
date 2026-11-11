package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

// debugHook lets a script drive the app while GOREX_DEBUG names a
// directory: each line of its file "do" is run, and "shot" captures the
// window into shot.png. Besides the commands' titles, the verbs are: type,
// tab, focus, name, close, zoom, size, sleep, shot; drop <path>… (files
// dropped on the focused pane, as the window does it); clip-image <png> (an
// image on the clipboard); and paste (what the paste shortcut does).
func (a *App) debugHook() {
	dir := os.Getenv("GOREX_DEBUG")
	if dir == "" {
		return
	}
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	go func() {
		for range time.Tick(150 * time.Millisecond) {
			req := filepath.Join(dir, "do")
			b, err := os.ReadFile(req)
			if err != nil {
				continue
			}
			os.Remove(req)
			for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
				a.debug(dir, strings.TrimSpace(line))
			}
		}
	}()
}

// debugDrop drops the files of arg, separated by spaces that no backslash
// escapes, on the focused pane.
func (a *App) debugDrop(arg string) {
	t := a.tab()
	if t == nil || t.Focus == nil {
		return
	}
	var files []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			files = append(files, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(arg); i++ {
		switch {
		case arg[i] == '\\' && i+1 < len(arg) && arg[i+1] == ' ':
			cur.WriteByte(' ')
			i++
		case arg[i] == ' ':
			flush()
		default:
			cur.WriteByte(arg[i])
		}
	}
	flush()
	a.dropFiles(t.Focus, files)
}

func (a *App) debug(dir, line string) {
	verb, arg, _ := strings.Cut(line, " ")
	win := a.win
	if win == nil {
		return
	}
	switch verb {
	case "shot":
		time.Sleep(100 * time.Millisecond)
		if png, err := win.CapturePage(); err == nil {
			name := "shot.png"
			if arg != "" {
				name = arg
			}
			os.WriteFile(filepath.Join(dir, name), png, 0o644)
		}
	case "sleep":
		d, _ := time.ParseDuration(arg)
		time.Sleep(d)
	case "type":
		arg = strings.ReplaceAll(arg, `\r`, "\r")
		arg = strings.ReplaceAll(arg, `\e`, "\x1b")
		win.Update(func() {
			if t := a.tab(); t != nil && t.Focus != nil && t.Focus.term != nil {
				t.Focus.term.Send([]byte(arg))
			}
		})
	case "drop":
		// drop <path>…: files dropped on the focused pane, as the window
		// does it. Spaces in a path are written \ .
		win.Update(func() { a.debugDrop(arg) })
	case "clip-image":
		// clip-image <png>: puts an image on the clipboard, for paste.
		if png, err := os.ReadFile(arg); err == nil {
			if err := mygo.Clipboard.WriteImage(png); err != nil {
				log.Printf("clip-image: %v", err)
			}
		} else {
			log.Printf("clip-image: %v", err)
		}
	case "paste":
		// paste: what Command+V does in the focused pane: the clipboard's
		// text, or its image where there is no text.
		win.Update(func() {
			if t := a.tab(); t != nil && t.Focus != nil {
				a.pasteClipboard(t.Focus, a.getServices().ReadClipboard())
			}
		})
	case "size":
		var w, h int
		if _, err := fmtSscan(arg, &w, &h); err == nil {
			win.SetSize(w, h)
		}
	default:
		for _, cmd := range commands {
			if strings.EqualFold(strings.ReplaceAll(cmd.Title, " ", "-"), verb) {
				a.run(cmd)
			}
		}
		switch verb {
		case "tab":
			win.Update(func() { a.selectTab(atoi(arg) - 1) })
		case "focus":
			win.Update(func() {
				if t := a.tab(); t != nil {
					if ps := t.panes(); atoi(arg)-1 < len(ps) {
						t.setFocus(ps[atoi(arg)-1])
						a.focusReq = t.Focus
					}
				}
			})
		case "name":
			win.Update(func() {
				if t := a.tab(); t != nil {
					t.Name = arg
				}
			})
		case "close":
			win.Close()
		case "zoom":
			win.Update(func() {
				if win.IsMaximized() {
					win.Unmaximize()
				} else {
					win.Maximize()
				}
			})
		}
	}
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			n = n*10 + int(r-'0')
		}
	}
	return n
}

func fmtSscan(s string, w, h *int) (int, error) {
	a, b, _ := strings.Cut(s, " ")
	*w, *h = atoi(a), atoi(b)
	return 2, nil
}
