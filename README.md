# GoRex

A replica of [Superlogical's Rex](https://www.superlogical.com/updates/public-testing-beginning)
terminal, written in Go with [MyGo](https://mygo.egoist.dev)'s native UI and its
terminal plugin (Ghostty's libghostty-vt). No webview, no cgo: one ~15 MB app.

![GoRex](docs/screenshot.png)

## What it does

- **Tabs and split panes.** Every pane is a card over the window's gradient,
  with its program's icon, name and directory in its header, and buttons to
  split right, split down, zoom and close. Drag the gaps between panes to
  resize them (double-click resets a split to half).
- **Persistent sessions.** Shells run in a session server, a background
  process of the same binary (`GoRex -server`). Quit GoRex and everything
  keeps running; open it again and every tab, split and pane comes back as
  it was, programs still running. The server keeps each session's screen in
  a headless libghostty-vt emulator and sends a snapshot on attach, so
  full-screen programs (lazygit, vim, htop…) come back exactly.
- **Program activity.** The server watches each terminal's foreground
  process (`tcgetpgrp`, `sysctl`, `proc_pidinfo`; on Windows, the shell's
  newest child, its command line and directory): headers and tabs name the
  program (`Node`, `Git Changes` for lazygit, `Codex`, `SSH host`…) and its
  working directory, a green dot shows a program printing, and an orange dot
  marks a pane whose program finished or rang the bell out of sight. When
  the window is in the background, a long command finishing posts a
  notification.
- **Tab icons** stack the tiles of a tab's programs, the focused pane's in
  front, as Rex does. Tabs reorder by dragging, rename by double-click, and
  have a context menu.
- **Images for Claude Code and opencode.** Drop a file on a pane, or paste a
  screenshot (⇧⌘V, Edit ▸ Paste Image; Ctrl+Shift+V elsewhere), and GoRex pastes
  the file's path as one bracketed paste, which both agents turn into an
  attachment (`[Image #1]` in Claude Code, `[Image 1]` in opencode). Text on
  the clipboard still pastes as text, and so do files copied in the Finder.
  When the pane runs `ssh`, the path of this computer means nothing over
  there, and neither agent can read this computer's clipboard from the other
  end, so GoRex first copies the file to `~/.gorex/drops` on that host and
  pastes the path of the copy. Progress and failures show in the pane's
  header. See [Images over ssh](#images-over-ssh) for how, and for the limits.
- **Host chip** with the machine's name and model; its popover shows the
  chip, memory, OS and the session server.
- **Command palette** (⇧⌘P, Ctrl+Shift+P, or the ⌘ button): every command, and every pane
  to jump to, fuzzy-matched.
- Light and dark appearances (View ▸ Appearance), text size (⌘+ ⌘− ⌘0),
  JetBrains Mono embedded, and a Settings window (⌘,) for both.

## Shortcuts

| macOS | Windows, Linux | |
|---|---|---|
| ⌘T | Ctrl+Shift+T | New tab |
| ⌘D / ⇧⌘D | Ctrl+Shift+D / Ctrl+Shift+E | Split right / down |
| ⌘W / ⇧⌘W | Ctrl+Shift+W | Close pane / tab |
| ⇧⌘↩ | Ctrl+Shift+Enter | Zoom the pane |
| ⌥⌘ arrows | Alt+arrows | Focus the pane in that direction |
| ⌃⌘ arrows | Alt+Shift+arrows | Move the nearest divider |
| ⌃⌘= | | Equalize panes |
| ⌘1…⌘9, ⇧⌘[ ⇧⌘], ⌃Tab | Ctrl+Shift+1…9, Ctrl+PageUp/PageDown, Ctrl+Tab | Switch tabs |
| ⇧⌘R | Ctrl+Shift+R | Rename tab |
| ⇧⌘P, ⌘P | Ctrl+Shift+P | Command palette |
| ⌘K | Ctrl+Shift+K | Clear |
| ⌘V | Ctrl+Shift+V | Paste text (elsewhere than on macOS, an image too where the clipboard has no text) |
| ⇧⌘V | Edit menu, palette | Paste Image: the clipboard's image, as a dropped file |
| ⌘+ ⌘− ⌘0 | Ctrl+= Ctrl+- Ctrl+0 | Text size |
| ⌘, | Ctrl+, | Settings |
| ⌥⌘Q | | Quit and end all sessions |

Elsewhere than on macOS, Control and a letter are the terminal's (Ctrl+C,
Ctrl+D, Ctrl+W…), so GoRex's shortcuts take Shift too, as copying and
pasting do (Ctrl+Shift+C, Ctrl+Shift+V), and none take Control and Alt,
which type characters (AltGr) on many keyboards. Ctrl+= Ctrl+- Ctrl+0 for
the text size and Ctrl+, for Settings go without Shift, as in other apps:
in the terminal's legacy encoding Ctrl+, sends nothing, so the terminal
loses nothing to it.

## Images over ssh

How a drop or a pasted screenshot reaches an agent:

- **Local pane:** the file's absolute path, with shell characters escaped, in
  one bracketed paste. A pasted screenshot is first written to
  `GoRex/pastes` in the user cache directory (`~/Library/Caches` on macOS,
  `~/.cache` on Linux, `%LocalAppData%` on Windows; mode 600; removed after
  a day).
- **Pane running `mosh`:** the same as `ssh`. mosh replaces itself with
  `mosh-client`, whose first argument still holds the typed command, so GoRex
  reads the host and a `--ssh=…` command from it and copies over ssh. The
  quotes of a `--ssh="…"` command are gone by then, so mosh's own `-p` right
  after it is read as ssh's; the copy then fails with its reason in the header.
  A `mosh-client` started by hand, with just an address, is not recognised.
- **Pane running `ssh`:** the file is copied over a second connection that
  repeats the pane's `-p -i -J -F -o` options (not its forwards, remote
  command or control-master options) to `~/.gorex/drops` on the host, mode 600,
  and the absolute path of the copy is pasted. It needs key or agent
  authentication, or a control master the pane's config already opens; a
  password prompt is not possible. Copies older than a day are removed from
  that directory on the next copy: the agents keep their own copy of a pasted
  image, so the file is needed only until the paste lands.
- **Several files** are pasted in one paste, separated by spaces.
- **Nothing is pasted on a failure** (no key access, host down, ssh ended
  during the copy): the pane's header says why.

Limits of the agents themselves (from their docs and source, and tested on
Claude Code 2.1.322 and opencode 2.2.5 and 2.0.29):

- Both need a bracketed paste: a path typed as keystrokes stays text. A pane
  whose program has not turned bracketed paste on, such as a plain shell,
  gets the text.
- Claude Code takes JPEG, PNG, GIF and WebP; opencode adds PDF. Both resize
  large images. opencode refuses an attachment over 20 MB at the paste, and
  GoRex refuses larger files with a message, so as not to copy one that
  would not attach.
- opencode attaches the image whatever its model, but a text-only model then
  says it cannot see it.

Not handled: a second hop (ssh run inside the ssh), `docker exec`
and password-only hosts (the file goes to the first ssh host, or the copy
fails); folders; images dragged out of a browser (a drop gives GoRex file
paths only); and pasting a screenshot from the terminal's right-click menu
or with ⌘V on macOS, where the Edit menu keeps ⌘V for text (use ⇧⌘V or
Edit ▸ Paste Image).

Dropping the same file on other terminals, read from their source and not
tried with a real drag: iTerm2, Ghostty, WezTerm and Alacritty paste the local
path, which attaches an image in both agents when they run on the same
computer; kitty does so for one file only; Terminal.app and the paths of
Windows Terminal are not known. None copies the file, so over ssh the path
points at the Mac. The notes are in `docs/research/`.

## Develop and build

GoRex needs [Go](https://go.dev/dl/) 1.27+ and MyGo 0.3.0, whose CLI
`go.mod` pins as a tool:

```sh
go tool mygo dev     # GoRex Dev, rebuilt and restarted as the code changes
go test ./...        # the server, and the view without a window
go tool mygo build   # build/darwin-arm64/GoRex.app and a .dmg
go run ./tools/mkicon  # redraw resources/icon.png
```

`go get -tool github.com/egoist/mygo/cmd/mygo@latest` updates MyGo and its
CLI together.

The app keeps its state in its data directory: the server's socket and
log, `layout.json` and `settings.json`, in `~/Library/Application
Support/GoRex` for the built app and `GoRex Dev` for `mygo dev`'s, so
that developing never touches the sessions of the app you use
(`GOREX_DIR` names another directory). The sessions outlive the app, but
not a rebuild: a development build replaces a server that an older build
started, ending its sessions, which then start again in the same
directories.

### Windows

GoRex runs on Windows 10 1809 or later, whose ConPTY runs the shells; its
data directory is `%AppData%\GoRex`. Panes run `$SHELL` when it names a
Windows program (`nu`, `C:\Program Files\Git\bin\bash.exe`), else
PowerShell (`pwsh`, then `powershell`), else `cmd`. PowerShell's `cd`
leaves its process's directory where it was, so GoRex starts it with a
prompt that sets it, wrapping the profile's own prompt: the header then
follows it, and new panes open there. The window is frameless and its
title bar draws minimize, maximize and close with the other buttons; it
resizes from its sides and bottom, not its top.

## Layout

| | |
|---|---|
| `main.go` | the app, its window, and `-server` |
| `state.go` | tabs, the tree of splits, panes, saving and restoring the layout |
| `view.go`, `tabs.go`, `host.go`, `commands.go` | the interface: title bar, tabs, panes, host popover, menus and palette |
| `programs.go` | how programs show: names, glyphs, tile colors |
| `style.go`, `settings.go` | colors, fonts, terminal themes; appearance and text size, and `settings.json` and its migration from version 1 |
| `preferences.go` | the Settings window: General, and the Default profile's text size |
| `internal/rex` | the session server and its client: PTYs, foreground processes, headless screens, the socket protocol |

Not replicated: Rex's connections to servers on other machines; GoRex's
server listens on a local socket only.

Icons: [Lucide](https://lucide.dev) (ISC) and [Simple Icons](https://simpleicons.org)
(CC0). Font: [JetBrains Mono](https://www.jetbrains.com/lp/mono/) (OFL).
