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

	"github.com/egoist/mygo/ui"
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

// remoteAgent is a pane running an agent over a real ssh, driven through
// the app's own view.
type remoteAgent struct {
	t    *testing.T
	a    *App
	tt   *ui.Tester
	p    *Pane
	host string
}

// startRemoteAgent runs `ssh -t host 'cd ~ && exec <agent>'` in a pane and
// waits until the agent shows its prompt, which ready recognises on the
// screen. The drops it makes are removed from the remote afterwards.
func startRemoteAgent(t *testing.T, host, agent string, ready func(screen string) bool) *remoteAgent {
	t.Helper()
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	t.Cleanup(func() {
		// find with quoted patterns: the remote's login shell may refuse a
		// command line whose glob matches nothing.
		remoteRun(t, host, "find ~/"+remoteDropDir+" -type f \\( -name '*-gorex-drop.png' -o -name '*-paste-*.png' \\) -delete; rmdir ~/"+remoteDropDir+" ~/.gorex 2>/dev/null; true")
	})
	r := &remoteAgent{t: t, a: a, tt: tt, p: p, host: host}
	p.term.Send([]byte("ssh -t " + host + " 'cd ~ && exec " + agent + "'\r"))
	r.step("the pane's foreground to be ssh", 20*time.Second, func() bool {
		_, ok := isSSH(p.info.Program, p.info.Args)
		return ok
	})
	r.step(agent+"'s prompt", 60*time.Second, func() bool { return ready(r.screen()) })
	time.Sleep(1500 * time.Millisecond) // the input box takes the focus
	return r
}

func (r *remoteAgent) screen() string { return r.p.term.Text() }

// step runs frames until cond holds, or fails with what the pane shows.
func (r *remoteAgent) step(what string, d time.Duration, cond func() bool) {
	r.t.Helper()
	deadline := time.Now().Add(d)
	for !cond() {
		if time.Now().After(deadline) {
			r.t.Fatalf("timed out waiting for %s; the pane shows:\n%s", what, r.screen())
		}
		time.Sleep(100 * time.Millisecond)
		refresh(r.a)
		r.a.runPosted()
		r.tt.Frame()
	}
}

// noLocalPath fails when the pane shows a path of this machine.
func (r *remoteAgent) noLocalPath(local string) {
	r.t.Helper()
	if strings.Contains(strings.ReplaceAll(r.screen(), "\n", ""), filepath.Dir(local)) {
		r.t.Error("the path on this machine reached the agent")
	}
}

func claudeReady(s string) bool {
	return strings.Contains(s, "Claude Code") && strings.Contains(s, "❯")
}
func opencodeReady(s string) bool { return strings.Contains(s, "Ask anything") }

// TestDropIntoRemoteClaude is the whole path: a pane runs a real ssh to
// Claude Code on the remote, an image is dropped on it, and Claude Code
// shows it as an attachment and describes it.
func TestDropIntoRemoteClaude(t *testing.T) {
	host := remoteHost(t)
	src, _ := orangePNG(t, "gorex-drop.png")
	r := startRemoteAgent(t, host, "claude", claudeReady)

	r.a.dropFiles(r.p, []string{src})
	r.step("the image to be attached", 60*time.Second, func() bool { return strings.Contains(r.screen(), "[Image #1]") })
	r.noLocalPath(src)

	// Claude Code reads it: ask what color it is.
	r.p.term.Send([]byte(" Which single color fills this picture? Reply with just the color name."))
	time.Sleep(500 * time.Millisecond) // one write with the Enter would be taken for a paste
	r.p.term.Send([]byte("\r"))
	asked := strings.Count(strings.ToLower(r.screen()), "orange")
	r.step("Claude Code to name the color", 90*time.Second, func() bool {
		return strings.Count(strings.ToLower(r.screen()), "orange") > asked
	})
	for _, l := range strings.Split(r.screen(), "\n") {
		if strings.Contains(l, "Image #1") || strings.Contains(strings.ToLower(l), "orange") {
			t.Log(strings.TrimSpace(l))
		}
	}
	if ok, _ := regexp.MatchString(`\[Image #1\]`, r.screen()); !ok {
		t.Error("the attachment is gone from the screen")
	}
}

// TestDropIntoRemoteOpencode is the same for opencode, which shows an
// attachment as [Image 1]. Nothing is submitted, so no model is called and
// no session is made.
func TestDropIntoRemoteOpencode(t *testing.T) {
	host := remoteHost(t)
	src, _ := orangePNG(t, "gorex-drop.png")
	r := startRemoteAgent(t, host, "opencode", opencodeReady)

	r.a.dropFiles(r.p, []string{src})
	r.step("the image to be attached", 60*time.Second, func() bool { return strings.Contains(r.screen(), "[Image 1]") })
	r.noLocalPath(src)
}

// TestPasteImageIntoRemoteAgents is the screenshot route: an image on the
// clipboard, no text, pasted into a pane over ssh, reaches each agent.
func TestPasteImageIntoRemoteAgents(t *testing.T) {
	host := remoteHost(t)
	src, _ := orangePNG(t, "gorex-drop.png")
	shot, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, ag := range []struct {
		name  string
		ready func(string) bool
		chip  string
	}{{"claude", claudeReady, "[Image #1]"}, {"opencode", opencodeReady, "[Image 1]"}} {
		t.Run(ag.name, func(t *testing.T) {
			dir := withClipboardImage(t, shot)
			r := startRemoteAgent(t, host, ag.name, ag.ready)
			r.a.pasteClipboard(r.p, "")
			r.step("the pasted image to be attached", 60*time.Second, func() bool { return strings.Contains(r.screen(), ag.chip) })
			r.noLocalPath(src)
			// The local file made for the paste is gone: the copy is the one.
			if left, _ := filepath.Glob(filepath.Join(dir, "paste-*.png")); len(left) != 0 {
				t.Errorf("local files left: %q", left)
			}
		})
	}
}
