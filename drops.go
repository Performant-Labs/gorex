package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"
)

// Files dropped on a pane become text in it: the path of each file, as a
// terminal pastes it when Finder drops files on it, which is how a
// program such as Claude Code takes an image. When the pane runs ssh, the
// path of the client computer is no use to the program on the other end, so
// the files are copied there first, over a second connection with the
// pane's own ssh options, and the paths pasted are those of the copies.

const (
	// remoteDropDir is where copies go, from the remote user's home.
	remoteDropDir = ".gorex/drops"
	// remoteDropHours is how long a copy stays. The agents keep their own
	// copy of a pasted image (Claude Code copies it at the paste, opencode
	// holds its bytes), so a day only covers a paste that did not happen.
	remoteDropHours = 24
	// maxDropSize is the largest file copied to a remote: opencode refuses
	// an attachment over 20 MiB when the path is pasted, and would show
	// nothing of why, so a larger file is refused here, with a reason.
	maxDropSize = 20 << 20
)

// sshBin is the ssh program, a variable for tests.
var sshBin = "ssh"

// sshConn is how a pane's ssh reaches its host.
type sshConn struct {
	// dest is the destination as typed: host, user@host, an alias of
	// ~/.ssh/config or an ssh:// URI.
	dest string
	// opts are the options that tell how to reach it, to give to another
	// ssh to the same host.
	opts []string
}

// parseSSH reads the arguments of an ssh command line, the program's name
// first. It keeps the options that select the host, the port, the user, the
// keys, the config and the jump hosts, and leaves what would act on the
// other connection too: forwards, a remote command, a terminal, a control
// master. Like OpenSSH, it reads options after the host too (ssh host -p 22)
// and stops at the first other word, which starts the remote command.
func parseSSH(args []string) (sshConn, bool) {
	const withValue = "BbcDEeFIiJLlmOoPpQRSWw"
	var c sshConn
	for i := 1; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			if c.dest == "" && i+1 < len(args) {
				c.dest = args[i+1]
			}
			break
		}
		if len(a) < 2 || a[0] != '-' {
			if c.dest == "" {
				c.dest = a
				continue
			}
			break // the remote command
		}
		for j := 1; j < len(a); j++ {
			f := a[j]
			if !strings.ContainsRune(withValue, rune(f)) {
				if f == '4' || f == '6' {
					c.opts = append(c.opts, "-"+string(f))
				}
				continue
			}
			v := a[j+1:]
			if v == "" {
				i++
				if i >= len(args) {
					return sshConn{}, false
				}
				v = args[i]
			}
			if keepOption(f, v) {
				c.opts = append(c.opts, "-"+string(f), v)
			}
			break
		}
	}
	return c, c.dest != ""
}

// keepOption tells whether ssh's option -f v describes the connection, not
// what to do on it.
func keepOption(f byte, v string) bool {
	switch f {
	case 'B', 'b', 'c', 'F', 'I', 'i', 'J', 'l', 'm', 'P', 'p', 'S':
		return true
	case 'o':
		k, _, _ := strings.Cut(strings.ReplaceAll(v, "=", " "), " ")
		k = strings.ToLower(strings.TrimSpace(k))
		if strings.Contains(k, "forward") {
			return false
		}
		switch k {
		case "remotecommand", "requesttty", "sessiontype", "localcommand", "permitlocalcommand",
			"controlmaster", "controlpersist", "exitonforwardfailure", "tunnel", "tunneldevice",
			"escapechar", "stdinnull", "forkafterauthentication", "batchmode", "connecttimeout":
			return false
		}
		return true
	}
	return false
}

// isSSH tells that a session's foreground program is ssh, and gives how it
// reaches its host.
func isSSH(program string, args []string) (sshConn, bool) {
	name := strings.TrimSuffix(strings.ToLower(filepath.Base(program)), ".exe")
	if len(args) == 0 {
		return sshConn{}, false
	}
	switch name {
	case "ssh":
		return parseSSH(args)
	case "mosh-client":
		return parseMosh(args)
	}
	return sshConn{}, false
}

// parseMosh reads the arguments of a mosh-client. mosh logs in over ssh, then
// replaces itself with mosh-client <"-# the arguments it was given |"> <ip> <port>,
// so the host and the ssh options are in that first argument, joined by
// spaces (its quotes are gone, and the program's own name is not there). The ip is not used: the user is not in it.
func parseMosh(args []string) (sshConn, bool) {
	if len(args) < 2 || !strings.HasPrefix(args[1], "-# ") || !strings.HasSuffix(args[1], " |") {
		return sshConn{}, false
	}
	t := strings.Fields(strings.TrimSuffix(strings.TrimPrefix(args[1], "-# "), " |"))
	if len(t) == 0 {
		return sshConn{}, false
	}
	// What the ssh of the login gets, from --ssh="ssh -p 2222 -i key": its
	// tokens follow the option's first word.
	ssh := []string{"ssh"}
	// Mosh's own options that take a value, in a separate word.
	moshValue := "p"
	var dest string
	for i := 0; i < len(t) && dest == ""; i++ {
		a := t[i]
		switch {
		case a == "--":
			i = len(t)
		case strings.HasPrefix(a, "--ssh="):
			if p := strings.TrimPrefix(a, "--ssh="); filepath.Base(p) != "ssh" {
				return sshConn{}, false // another program than ssh
			}
			for i+1 < len(t) && strings.HasPrefix(t[i+1], "-") && !strings.HasPrefix(t[i+1], "--") {
				i++
				ssh = append(ssh, t[i])
				if f := t[i]; len(f) == 2 && strings.ContainsRune("BbcDEeFIiJLlmOoPpQRSWw", rune(f[1])) && i+1 < len(t) {
					i++
					ssh = append(ssh, t[i])
				}
			}
		case strings.HasPrefix(a, "--"):
			// A long option of mosh, with its value attached.
		case len(a) == 2 && a[0] == '-' && strings.Contains(moshValue, a[1:]):
			i++ // its value
		case len(a) > 1 && a[0] == '-':
			// -4, -6, -a, -n and the like
		default:
			dest = a
		}
	}
	if dest == "" {
		return sshConn{}, false
	}
	return parseSSH(append(ssh, dest))
}

// dropName makes the name of the copy of a file: the time, for the days it
// stays and for names that repeat, and the file's name with only characters
// a shell leaves alone, and its extension, which tells what the file is.
func dropName(now time.Time, n int, base string) string {
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	clean := func(s string) string {
		return strings.Map(func(r rune) rune {
			if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-') {
				return r
			}
			return '_'
		}, s)
	}
	stem, ext = clean(stem), clean(ext)
	if len(stem) > 60 {
		stem = stem[:60]
	}
	return fmt.Sprintf("%s-%d-%s%s", now.Format("20060102-150405"), n, strings.TrimLeft(stem, "."), ext)
}

// remoteScript runs on the remote, with the content of the file on its
// standard input: it makes the directory, removes the old copies, writes
// the file, and ends with its absolute path on a line of its own. name has
// only safe characters, and the script none of single quotes, as it goes
// to the remote's shell inside a pair of them.
func remoteScript(name string) string {
	return fmt.Sprintf(`umask 077; d="$HOME/%s"; mkdir -p "$d" && find "$d" -type f -mmin +%d -exec rm -f {} + 2>/dev/null; cat > "$d/%s" && printf "\n%%s\n" "$d/%s"`,
		remoteDropDir, remoteDropHours*60, name, name)
}

// upload copies the file to the remote and returns its path there. dir is
// where the pane's ssh runs, from which its relative option paths resolve.
func (c sshConn) upload(ctx context.Context, dir, path, name string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !st.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a file", filepath.Base(path))
	}
	if st.Size() > maxDropSize {
		return "", fmt.Errorf("%s is larger than %d MB", filepath.Base(path), maxDropSize>>20)
	}
	args := []string{"-T", "-o", "BatchMode=yes", "-o", "ConnectTimeout=10", "-o", "ControlMaster=no", "-o", "RemoteCommand=none",
		// What the config gives every connection to the host must not act
		// twice: forwards the pane already holds, and a local command.
		"-o", "ClearAllForwardings=yes", "-o", "PermitLocalCommand=no"}
	args = append(args, c.opts...)
	args = append(args, "--", c.dest, "sh -c '"+remoteScript(name)+"'")
	cmd := exec.CommandContext(ctx, sshBin, args...)
	if dir != "" {
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			cmd.Dir = dir
		}
	}
	var out, errb bytes.Buffer
	cmd.Stdin, cmd.Stdout, cmd.Stderr = f, &out, &limitBuffer{b: &errb, n: 4096}
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("timed out copying %s to %s", filepath.Base(path), c.dest)
		}
		return "", fmt.Errorf("%s: %s", c.dest, lastLine(errb.String(), err.Error()))
	}
	got := lastLine(out.String(), "")
	if !strings.HasPrefix(got, "/") {
		return "", fmt.Errorf("%s: no path came back for %s", c.dest, filepath.Base(path))
	}
	return got, nil
}

// uploadAll copies the files, one connection each, and returns their
// paths on the remote.
func (c sshConn) uploadAll(dir string, files []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	now := time.Now()
	var paths []string
	for i, f := range files {
		p, err := c.upload(ctx, dir, f, dropName(now, i, filepath.Base(f)))
		if err != nil {
			return paths, err
		}
		paths = append(paths, p)
	}
	return paths, nil
}

// limitBuffer keeps the last n bytes written to it: ssh says why it failed
// at the end of its output, after any debug lines.
type limitBuffer struct {
	b *bytes.Buffer
	n int
}

func (l *limitBuffer) Write(p []byte) (int, error) {
	l.b.Write(p)
	if extra := l.b.Len() - l.n; extra > 0 {
		l.b.Next(extra)
	}
	return len(p), nil
}

var _ io.Writer = (*limitBuffer)(nil)

// lastLine returns the last line of s that is not empty, or def.
func lastLine(s, def string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if l := strings.TrimSpace(lines[len(lines)-1]); l != "" {
		return l
	}
	return def
}

// escapeUnix escapes a Unix path's shell characters with a backslash, as
// a terminal pastes a dropped file.
func escapeUnix(path string) string {
	var b strings.Builder
	for _, r := range path {
		if !(r > unicode.MaxASCII || unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("/._-+:@%,=~", r)) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// pasteQuote writes a path of this machine as its terminals take it: with
// escapes on Unix, and on Windows in quotes when it holds a space.
func pasteQuote(path string) string {
	if runtime.GOOS != "windows" {
		return escapeUnix(path)
	}
	if strings.ContainsAny(path, " \t") {
		return `"` + path + `"`
	}
	return path
}

// pasteLine joins paths, quoted by quote, with a space after each, as a
// terminal does with several dropped files.
func pasteLine(paths []string, quote func(string) string) string {
	var b strings.Builder
	for _, p := range paths {
		b.WriteString(quote(p))
		b.WriteByte(' ')
	}
	return b.String()
}

// dropMessageTime is how long a pane's header keeps what a drop did.
const dropMessageTime = 7 * time.Second

// dropFiles handles files dropped on a pane. On the pane of an ssh, they
// are copied to its host in the background, and the paste, or the error,
// comes when that is done.
func (a *App) dropFiles(p *Pane, files []string) { a.dropFilesFrom(p, files, false) }

// dropFilesFrom is dropFiles for files GoRex made itself, as a pasted image,
// when temp is set: they are removed once copied to an ssh host, whether it
// worked or not, as the copy is the one that matters. A local pane's agent
// reads the file by its path, so there it stays, and savePaste removes it
// later.
func (a *App) dropFilesFrom(p *Pane, files []string, temp bool) {
	if p.term == nil || p.closed || len(files) == 0 {
		return
	}
	conn, remote := isSSH(p.info.Program, p.info.Args)
	if !remote {
		p.term.Paste(pasteLine(files, pasteQuote))
		return
	}
	dir := p.info.Dir
	p.setDropNote(a, "Copying to "+conn.dest+"…", false, 0)
	go func() {
		paths, err := conn.uploadAll(dir, files)
		if temp {
			for _, f := range files {
				os.Remove(f)
			}
		}
		a.laterService(func() {
			if p.closed {
				return
			}
			if err != nil {
				p.setDropNote(a, "Drop failed: "+err.Error(), true, dropMessageTime)
				return
			}
			// The program the paste is for must still be the ssh the
			// files went to: after it ended the shell is the client's.
			if now, ok := isSSH(p.info.Program, p.info.Args); !ok || now.dest != conn.dest {
				p.setDropNote(a, "The ssh to "+conn.dest+" ended; copies are in ~/"+remoteDropDir+" there", true, dropMessageTime)
				return
			}
			p.term.Paste(pasteLine(paths, escapeUnix))
			p.setDropNote(a, "", false, 0)
		})
	}()
}

// setDropNote shows a message in the pane's header, and, with a duration,
// clears it after.
func (p *Pane) setDropNote(a *App, msg string, isErr bool, clear time.Duration) {
	p.dropSeq++
	seq := p.dropSeq
	p.dropNote, p.dropErr = msg, isErr
	if clear > 0 {
		time.AfterFunc(clear, func() {
			a.laterService(func() {
				if p.dropSeq == seq {
					p.dropNote = ""
				}
			})
		})
	}
}
