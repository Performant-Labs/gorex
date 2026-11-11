package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// These tests use a real remote, named by GOREX_TEST_SSH (a host that
// ssh reaches without a password): they skip without it. The second one
// also needs Claude Code on the remote, signed in, and trusting the home
// directory; it makes one short request.

func remoteHost(t *testing.T) string {
	t.Helper()
	h := os.Getenv("GOREX_TEST_SSH")
	if h == "" || runtime.GOOS == "windows" {
		t.Skip("GOREX_TEST_SSH names no remote")
	}
	return h
}

// orangePNG writes a PNG that is one color, orange.
func orangePNG(t *testing.T, name string) (path string, sum [32]byte) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 320, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			img.Set(x, y, color.RGBA{255, 140, 0, 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, sha256.Sum256(b.Bytes())
}

// remoteRun runs a command on the remote.
func remoteRun(t *testing.T, host, script string) string {
	t.Helper()
	out, err := exec.Command("ssh", "-o", "BatchMode=yes", host, script).CombinedOutput()
	if err != nil {
		t.Fatalf("ssh %s %q: %v\n%s", host, script, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestUploadToRemote(t *testing.T) {
	host := remoteHost(t)
	src, sum := orangePNG(t, "Screen Shot (1).png")
	conn, _ := parseSSH([]string{"ssh", host})
	paths, err := conn.uploadAll("", []string{src})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		remoteRun(t, host, "rm -f "+escapeUnix(paths[0])+"; rmdir ~/"+remoteDropDir+" ~/.gorex 2>/dev/null; true")
	})
	home := remoteRun(t, host, "printf %s \"$HOME\"")
	if want := home + "/" + remoteDropDir + "/"; !strings.HasPrefix(paths[0], want) || !strings.HasSuffix(paths[0], "Screen_Shot__1_.png") {
		t.Errorf("path %q; want under %q", paths[0], want)
	}
	// The copy is the same bytes, private, and readable at that path.
	got := remoteRun(t, host, "sha256sum "+escapeUnix(paths[0])+" | cut -d' ' -f1; stat -c %a "+escapeUnix(paths[0]))
	lines := strings.Fields(got)
	if len(lines) != 2 || lines[0] != fmt.Sprintf("%x", sum) || lines[1] != "600" {
		t.Errorf("remote sha256 and mode %q; want %s 600", got, fmt.Sprintf("%x", sum))
	}
}

// TestDropIntoRemoteClaude is the whole path: a pane runs a real ssh to
// Claude Code on the remote, an image is dropped on it, and Claude Code
// shows it as an attachment and describes it.
func TestDropIntoRemoteClaude(t *testing.T) {
	host := remoteHost(t)
	src, _ := orangePNG(t, "gorex-drop.png")
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	t.Cleanup(func() {
		remoteRun(t, host, "rm -f ~/"+remoteDropDir+"/*-gorex-drop.png; rmdir ~/"+remoteDropDir+" ~/.gorex 2>/dev/null; true")
	})

	step := func(what string, d time.Duration, cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(d)
		for !cond() {
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for %s; the pane shows:\n%s", what, p.term.Text())
			}
			time.Sleep(100 * time.Millisecond)
			refresh(a)
			a.runPosted()
			tt.Frame()
		}
	}
	screen := func() string { return p.term.Text() }

	p.term.Send([]byte("ssh -t " + host + " 'cd ~ && exec claude'\r"))
	step("the pane's foreground to be ssh", 20*time.Second, func() bool {
		_, ok := isSSH(p.info.Program, p.info.Args)
		return ok
	})
	step("Claude Code's prompt", 60*time.Second, func() bool {
		s := screen()
		return strings.Contains(s, "Claude Code") && strings.Contains(s, "❯")
	})
	time.Sleep(1500 * time.Millisecond) // the input box takes the focus

	a.dropFiles(p, []string{src})
	step("the image to be attached", 60*time.Second, func() bool { return strings.Contains(screen(), "[Image #1]") })

	// What is pasted is the path on the remote, not the one here.
	flat := strings.ReplaceAll(screen(), "\n", "")
	if strings.Contains(flat, filepath.Dir(src)) {
		t.Error("the path on this machine reached Claude Code")
	}

	// Claude Code reads it: ask what color it is.
	p.term.Send([]byte(" Which single color fills this picture? Reply with just the color name."))
	time.Sleep(500 * time.Millisecond) // one write with the Enter would be taken for a paste
	p.term.Send([]byte("\r"))
	asked := strings.Count(strings.ToLower(screen()), "orange")
	step("Claude Code to name the color", 90*time.Second, func() bool {
		return strings.Count(strings.ToLower(screen()), "orange") > asked
	})
	for _, l := range strings.Split(screen(), "\n") {
		if strings.Contains(l, "Image #1") || strings.Contains(strings.ToLower(l), "orange") {
			t.Log(strings.TrimSpace(l))
		}
	}
	if ok, _ := regexp.MatchString(`\[Image #1\]`, screen()); !ok {
		t.Error("the attachment is gone from the screen")
	}
}
