package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseSSH(t *testing.T) {
	for _, tc := range []struct {
		args []string
		dest string
		opts []string
		ok   bool
	}{
		{[]string{"ssh", "jupiter"}, "jupiter", nil, true},
		{[]string{"ssh", "-p", "2222", "me@host"}, "me@host", []string{"-p", "2222"}, true},
		{[]string{"ssh", "-p2222", "-i", "id", "host"}, "host", []string{"-p", "2222", "-i", "id"}, true},
		// A cluster whose last flag takes the value, and booleans before it.
		{[]string{"ssh", "-4tp", "22", "host", "ls"}, "host", []string{"-4", "-p", "22"}, true},
		{[]string{"ssh", "-J", "bastion", "-F", "cfg", "-o", "User=x", "host"}, "host", []string{"-J", "bastion", "-F", "cfg", "-o", "User=x"}, true},
		{[]string{"ssh", "--", "host"}, "host", nil, true},
		{[]string{"ssh", "ssh://me@host:2200"}, "ssh://me@host:2200", nil, true},
		// What acts on the other connection too is left out.
		{[]string{"ssh", "-L", "8080:localhost:80", "-R", "9:9", "-D", "1080", "-W", "h:1", "host"}, "host", nil, true},
		{[]string{"ssh", "-o", "LocalForward=1:2:3", "-o", "RemoteCommand=top", "-o", "RequestTTY=force", "host"}, "host", nil, true},
		{[]string{"ssh", "-o", "ControlMaster=yes", "-o", "ControlPath=/s", "host"}, "host", []string{"-o", "ControlPath=/s"}, true},
		// Options after the host count, as in OpenSSH; the first other word
		// starts the remote command, and what follows it is not options.
		{[]string{"ssh", "-A", "-t", "-v", "host", "-p", "9"}, "host", []string{"-p", "9"}, true},
		{[]string{"ssh", "host", "-p", "9"}, "host", []string{"-p", "9"}, true},
		{[]string{"ssh", "host", "-i", "id", "-J", "bastion"}, "host", []string{"-i", "id", "-J", "bastion"}, true},
		{[]string{"ssh", "host", "-4", "-L", "1:2:3", "-p2200"}, "host", []string{"-4", "-p", "2200"}, true},
		{[]string{"ssh", "host", "ls", "-p", "9"}, "host", nil, true},
		{[]string{"ssh", "host", "--", "-p", "9"}, "host", nil, true},
		{[]string{"ssh", "host", "-p"}, "", nil, false},
		// -P tag selects Match blocks of the config, and takes a value.
		{[]string{"ssh", "-P", "work", "host"}, "host", []string{"-P", "work"}, true},
		{[]string{"ssh"}, "", nil, false},
		{[]string{"ssh", "-p"}, "", nil, false},
	} {
		got, ok := parseSSH(tc.args)
		if ok != tc.ok || got.dest != tc.dest || !slices.Equal(got.opts, tc.opts) {
			t.Errorf("parseSSH(%q) = %q %q %v; want %q %q %v", tc.args, got.dest, got.opts, ok, tc.dest, tc.opts, tc.ok)
		}
	}
	if _, ok := isSSH("/usr/bin/ssh", []string{"ssh", "h"}); !ok {
		t.Error("/usr/bin/ssh is not ssh")
	}
	if _, ok := isSSH("ssh.EXE", []string{"ssh", "h"}); !ok {
		t.Error("ssh.EXE is not ssh")
	}
	if _, ok := isSSH("sleep", []string{"sleep", "3"}); ok {
		t.Error("sleep is ssh")
	}
	if _, ok := isSSH("ssh", nil); ok {
		t.Error("ssh without arguments is ssh")
	}
}

func TestDropName(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 30, 45, 0, time.UTC)
	for in, want := range map[string]string{
		"shot.png":                  "20261111-103045-0-shot.png",
		"Screenshot 2026-11-11.png": "20261111-103045-0-Screenshot_2026-11-11.png",
		"a;b$(rm -rf).png":          "20261111-103045-0-a_b__rm_-rf_.png",
		".hidden.png":               "20261111-103045-0-hidden.png",
		"é'\"`.JPG":                 "20261111-103045-0-____.JPG",
		"noext":                     "20261111-103045-0-noext",
	} {
		if got := dropName(now, 0, in); got != want {
			t.Errorf("dropName(%q) = %q; want %q", in, got, want)
		}
	}
	if got := dropName(now, 3, strings.Repeat("x", 200)+".png"); len(got) > 100 || !strings.HasSuffix(got, ".png") || !strings.Contains(got, "-3-") {
		t.Errorf("a long name became %q", got)
	}
}

func TestRemoteScriptHasNoSingleQuotes(t *testing.T) {
	if s := remoteScript(dropName(time.Now(), 0, "a.png")); strings.Contains(s, "'") {
		t.Errorf("the script is wrapped in single quotes and holds some: %s", s)
	}
}

func TestEscapeUnix(t *testing.T) {
	for in, want := range map[string]string{
		"/home/me/.gorex/drops/a.png": "/home/me/.gorex/drops/a.png",
		"/Users/me/My Shot (1).png":   `/Users/me/My\ Shot\ \(1\).png`,
		`/tmp/it's"q`:                 `/tmp/it\'s\"q`,
		"/tmp/é.png":                  "/tmp/é.png",
	} {
		if got := escapeUnix(in); got != want {
			t.Errorf("escapeUnix(%q) = %q; want %q", in, got, want)
		}
	}
}

// fakeSSH writes an ssh that runs the command locally as if it were the
// remote, in the home directory home, and logs its arguments.
func fakeSSH(t *testing.T, home string) (log string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake ssh is a shell script")
	}
	dir := t.TempDir()
	log = filepath.Join(dir, "args")
	script := `#!/bin/sh
printf '%s\n' "$@" > "` + log + `"
while [ "$1" != "--" ]; do shift; done
shift; shift
case "$GOREX_FAKE_FAIL" in ?*) echo "$GOREX_FAKE_FAIL" >&2; exit 255;; esac
HOME="` + home + `" exec /bin/sh -c "$*"
`
	path := filepath.Join(dir, "ssh")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	old := sshBin
	sshBin = path
	t.Cleanup(func() { sshBin = old })
	return log
}

func TestUpload(t *testing.T) {
	home := t.TempDir()
	log := fakeSSH(t, home)
	src := filepath.Join(t.TempDir(), "My Shot (1).png")
	want := []byte("\x89PNG\r\n\x1a\n binary \x00\x01\xff data")
	if err := os.WriteFile(src, want, 0o644); err != nil {
		t.Fatal(err)
	}

	conn, _ := parseSSH([]string{"ssh", "-p", "2222", "-L", "1:2:3", "me@host"})
	paths, err := conn.uploadAll("", []string{src, src})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0] == paths[1] {
		t.Fatalf("paths %q", paths)
	}
	for _, p := range paths {
		if filepath.Dir(p) != filepath.Join(home, remoteDropDir) || strings.ContainsAny(filepath.Base(p), " ()") {
			t.Errorf("the copy is at %q", p)
		}
		got, err := os.ReadFile(p)
		if err != nil || string(got) != string(want) {
			t.Errorf("the copy at %q is %q, %v", p, got, err)
		}
		if st, _ := os.Stat(p); st == nil || st.Mode().Perm() != 0o600 {
			t.Errorf("the copy at %q has mode %v", p, st.Mode())
		}
	}
	b, _ := os.ReadFile(log)
	args := strings.Split(strings.TrimSpace(string(b)), "\n")
	for _, w := range []string{"BatchMode=yes", "ControlMaster=no", "RemoteCommand=none", "ClearAllForwardings=yes", "PermitLocalCommand=no", "-T", "-p", "2222", "me@host"} {
		if !slices.Contains(args, w) {
			t.Errorf("ssh did not get %q: %q", w, args)
		}
	}
	if slices.Contains(args, "-L") || slices.Contains(args, "1:2:3") {
		t.Errorf("ssh got the pane's forward: %q", args)
	}
	if i := slices.Index(args, "--"); i < 0 || args[i+1] != "me@host" {
		t.Errorf("the destination does not follow --: %q", args)
	}
}

func TestUploadRemovesOldCopies(t *testing.T) {
	home := t.TempDir()
	fakeSSH(t, home)
	dir := filepath.Join(home, remoteDropDir)
	os.MkdirAll(dir, 0o700)
	old, fresh := filepath.Join(dir, "old.png"), filepath.Join(dir, "fresh.png")
	os.WriteFile(old, []byte("x"), 0o600)
	os.WriteFile(fresh, []byte("x"), 0o600)
	stale := time.Now().Add(-(remoteDropHours + 1) * time.Hour)
	os.Chtimes(old, stale, stale)

	src := filepath.Join(t.TempDir(), "a.png")
	os.WriteFile(src, []byte("y"), 0o644)
	conn, _ := parseSSH([]string{"ssh", "h"})
	if _, err := conn.uploadAll("", []string{src}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); err == nil {
		t.Error("a copy older than a day stays")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("a recent copy was removed")
	}
}

func TestUploadErrors(t *testing.T) {
	home := t.TempDir()
	fakeSSH(t, home)
	conn, _ := parseSSH([]string{"ssh", "host"})

	// ssh fails: its last line of complaint is the message.
	src := filepath.Join(t.TempDir(), "a.png")
	os.WriteFile(src, []byte("y"), 0o644)
	t.Setenv("GOREX_FAKE_FAIL", "Warning: x\nme@host: Permission denied (publickey).")
	if _, err := conn.uploadAll("", []string{src}); err == nil || !strings.Contains(err.Error(), "Permission denied (publickey).") || strings.Contains(err.Error(), "Warning") {
		t.Errorf("a failed ssh gave %v", err)
	}
	// Debug output before the complaint does not push it out.
	t.Setenv("GOREX_FAKE_FAIL", strings.Repeat("debug1: noise noise noise noise\n", 400)+"me@host: Connection refused")
	if _, err := conn.uploadAll("", []string{src}); err == nil || !strings.Contains(err.Error(), "Connection refused") || strings.Contains(err.Error(), "debug1") {
		t.Errorf("a failed ssh with 4 KB of debug output gave %v", err)
	}
	t.Setenv("GOREX_FAKE_FAIL", "")

	// A directory, a missing file and a file too large are not sent.
	if _, err := conn.uploadAll("", []string{t.TempDir()}); err == nil || !strings.Contains(err.Error(), "not a file") {
		t.Errorf("a directory gave %v", err)
	}
	if _, err := conn.uploadAll("", []string{filepath.Join(t.TempDir(), "nope.png")}); err == nil {
		t.Error("a missing file was sent")
	}
	big := filepath.Join(t.TempDir(), "big.bin")
	f, _ := os.Create(big)
	f.Truncate(maxDropSize + 1)
	f.Close()
	if _, err := conn.uploadAll("", []string{big}); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Errorf("a large file gave %v", err)
	}
	entries, _ := os.ReadDir(filepath.Join(home, remoteDropDir))
	if len(entries) != 0 {
		t.Errorf("failed drops left copies: %v", entries)
	}
}

func TestDropOnPane(t *testing.T) {
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })

	// On a pane that does not run ssh, the local path is pasted, escaped.
	local := filepath.Join(t.TempDir(), "my shot.png")
	os.WriteFile(local, []byte("png"), 0o644)
	a.dropFiles(p, []string{local})
	want := strings.ReplaceAll(local, " ", `\ `)
	if runtime.GOOS == "windows" {
		want = `"` + local + `"`
	}
	waitFor(t, tt, "the local path", func() bool { return strings.Contains(strings.ReplaceAll(p.term.Text(), "\n", ""), want) })
	if p.dropNote != "" {
		t.Errorf("a local drop left the note %q", p.dropNote)
	}
	if runtime.GOOS == "windows" {
		return
	}

	// On a pane that runs ssh, the file is copied and the remote path
	// pasted, whatever the local path was.
	home := t.TempDir()
	fakeSSH(t, home)
	secondDir := filepath.Join(t.TempDir(), "CLIENTONLY")
	os.MkdirAll(secondDir, 0o755)
	local = filepath.Join(secondDir, "shot.png")
	os.WriteFile(local, []byte("png"), 0o644)
	p.info.Program, p.info.Args, p.info.Idle = "ssh", []string{"ssh", "me@host"}, false
	a.dropFiles(p, []string{local})
	if !strings.Contains(p.dropNote, "me@host") || p.dropErr {
		t.Errorf("no note of the copy: %q", p.dropNote)
	}
	tt.Type("") // frames, so the posted paste runs
	waitFor(t, tt, "the remote path", func() bool {
		a.runPosted()
		return strings.Contains(strings.ReplaceAll(p.term.Text(), "\n", ""), filepath.Join(home, remoteDropDir)+"/")
	})
	if strings.Contains(p.term.Text(), "CLIENTONLY") {
		t.Error("the pane got the path on this machine")
	}
	if p.dropNote != "" {
		t.Errorf("the note stays after the paste: %q", p.dropNote)
	}

	// A failed copy shows its error in the header and pastes nothing.
	before := p.term.Text()
	t.Setenv("GOREX_FAKE_FAIL", "Connection refused")
	a.dropFiles(p, []string{local})
	waitFor(t, tt, "the error", func() bool { a.runPosted(); return p.dropErr })
	if !strings.Contains(p.dropNote, "Connection refused") || p.term.Text() != before {
		t.Errorf("note %q; pane changed: %v", p.dropNote, p.term.Text() != before)
	}

	// When the ssh ended before the copy was done, the shell is the
	// client's: nothing is pasted.
	t.Setenv("GOREX_FAKE_FAIL", "")
	a.dropFiles(p, []string{local})
	p.info.Program, p.info.Args, p.info.Idle = "sh", nil, true
	waitFor(t, tt, "the ended note", func() bool { a.runPosted(); return strings.Contains(p.dropNote, "ended") })
	if p.term.Text() != before {
		t.Error("a path was pasted after the ssh ended")
	}
}
