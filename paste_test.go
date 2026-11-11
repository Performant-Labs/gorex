package main

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

// testPNG is a small valid PNG.
func testPNG(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// withClipboardImage makes the clipboard hold png (none when nil) and the
// pasted images go in a directory of the test's own, which it returns.
func withClipboardImage(t *testing.T, png []byte) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "pastes")
	oldImg, oldDir := clipboardImage, pasteDir
	clipboardImage = func() []byte { return png }
	pasteDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { clipboardImage, pasteDir = oldImg, oldDir })
	return dir
}

// flat is the pane's text without its line breaks, for what wraps.
func flat(p *Pane) string { return strings.ReplaceAll(p.term.Text(), "\n", "") }

func TestPasteImageInLocalPane(t *testing.T) {
	want := testPNG(t)
	dir := withClipboardImage(t, want)
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })

	// With no text on the clipboard, its image is written and its path pasted.
	a.pasteClipboard(p, "")
	files, _ := filepath.Glob(filepath.Join(dir, "paste-*.png"))
	if len(files) != 1 {
		t.Fatalf("pasted images %q", files)
	}
	got, _ := os.ReadFile(files[0])
	if !bytes.Equal(got, want) {
		t.Error("the file is not the image")
	}
	if runtime.GOOS != "windows" { // its modes are not Unix's
		if st, _ := os.Stat(files[0]); st.Mode().Perm() != 0o600 {
			t.Errorf("mode %v", st.Mode())
		}
		if st, _ := os.Stat(dir); st.Mode().Perm() != 0o700 {
			t.Errorf("directory mode %v", st.Mode())
		}
	}
	waitFor(t, tt, "the path", func() bool { return strings.Contains(flat(p), pasteQuote(files[0])) })
}

func TestPasteTextWinsOverImage(t *testing.T) {
	dir := withClipboardImage(t, testPNG(t))
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	a.pasteClipboard(p, "echo-text-pasted")
	waitFor(t, tt, "the text", func() bool { return strings.Contains(flat(p), "echo-text-pasted") })
	if files, _ := filepath.Glob(filepath.Join(dir, "*")); len(files) != 0 {
		t.Errorf("an image was written with text on the clipboard: %q", files)
	}
}

func TestPasteNothing(t *testing.T) {
	dir := withClipboardImage(t, nil)
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	before := p.term.Text()
	a.pasteClipboard(p, "")
	if _, err := os.Stat(dir); err == nil || p.term.Text() != before || p.dropNote != "" {
		t.Error("an empty clipboard did something")
	}
}

func TestPasteImageTooLarge(t *testing.T) {
	dir := withClipboardImage(t, make([]byte, maxDropSize+1))
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	a.pasteClipboard(p, "")
	if !p.dropErr || !strings.Contains(p.dropNote, "larger than 20 MB") {
		t.Errorf("note %q err %v", p.dropNote, p.dropErr)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Error("a file too large was written")
	}
}

func TestPasteImageInSSHPane(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake ssh is a shell script")
	}
	want := testPNG(t)
	dir := withClipboardImage(t, want)
	home := t.TempDir()
	fakeSSH(t, home)
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	p.info.Program, p.info.Args, p.info.Idle = "ssh", []string{"ssh", "me@host"}, false

	a.pasteClipboard(p, "")
	waitFor(t, tt, "the remote path", func() bool {
		a.runPosted()
		return strings.Contains(flat(p), filepath.Join(home, remoteDropDir)+"/")
	})
	copies, _ := filepath.Glob(filepath.Join(home, remoteDropDir, "*-paste-*.png"))
	if len(copies) != 1 {
		t.Fatalf("copies %q", copies)
	}
	if got, _ := os.ReadFile(copies[0]); !bytes.Equal(got, want) {
		t.Error("the copy is not the image")
	}
	// The copy is the one that matters: the local file is gone, and the
	// pane got no local path.
	if left, _ := filepath.Glob(filepath.Join(dir, "*")); len(left) != 0 {
		t.Errorf("local files left: %q", left)
	}
	if strings.Contains(flat(p), dir) {
		t.Error("the pane got the path on this machine")
	}
}

func TestSavePasteRemovesOldImages(t *testing.T) {
	dir := withClipboardImage(t, nil)
	os.MkdirAll(dir, 0o700)
	old, fresh, other := filepath.Join(dir, "paste-old.png"), filepath.Join(dir, "paste-new.png"), filepath.Join(dir, "keep.txt")
	for _, f := range []string{old, fresh, other} {
		os.WriteFile(f, []byte("x"), 0o600)
	}
	stale := time.Now().Add(-pasteKeepTime - time.Hour)
	os.Chtimes(old, stale, stale)
	os.Chtimes(other, stale, stale)
	if _, err := savePaste(testPNG(t), time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); err == nil {
		t.Error("an image older than a day stays")
	}
	for _, f := range []string{fresh, other} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("%s was removed", filepath.Base(f))
		}
	}
}

// TestPasteKey presses the paste shortcut in the view: the pane takes it
// before its terminal does, and pastes the image where the terminal alone
// would paste nothing.
func TestPasteKey(t *testing.T) {
	dir := withClipboardImage(t, testPNG(t))
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	for i := 0; i < 3; i++ {
		tt.Frame()
	}
	mods, key := pasteKey()

	tt.SetClipboard("")
	tt.Key(mods, key)
	tt.Frame()
	files, _ := filepath.Glob(filepath.Join(dir, "paste-*.png"))
	if len(files) != 1 {
		t.Fatalf("the shortcut with an image on the clipboard wrote %q", files)
	}
	waitFor(t, tt, "the image's path", func() bool { return strings.Contains(flat(p), pasteQuote(files[0])) })

	tt.SetClipboard("text-from-the-clipboard")
	tt.Key(mods, key)
	tt.Frame()
	waitFor(t, tt, "the text", func() bool { return strings.Contains(flat(p), "text-from-the-clipboard") })
	if files2, _ := filepath.Glob(filepath.Join(dir, "paste-*.png")); len(files2) != 1 {
		t.Errorf("an image was written with text on the clipboard: %q", files2)
	}
}

func TestPasteImageCommand(t *testing.T) {
	dir := withClipboardImage(t, testPNG(t))
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	count := func() int { f, _ := filepath.Glob(filepath.Join(dir, "paste-*.png")); return len(f) }

	// A field of the app has the focus: not a pane to paste into.
	a.paletteOpen = true
	a.pasteImageFocused()
	a.paletteOpen = false
	a.renaming = a.tab()
	a.pasteImageFocused()
	a.renaming = nil
	if n := count(); n != 0 {
		t.Fatalf("%d images pasted when they should not", n)
	}
	// The command is in the menus and the palette, and runs the same.
	if !slices.Contains(commands, &cmdPasteImage) || !slices.Contains(paletteCommands, &cmdPasteImage) {
		t.Error("Paste Image is not a command of the palette")
	}
	cmdPasteImage.Run(a)
	waitFor(t, tt, "the pasted image", func() bool { a.runPosted(); return count() == 1 })

	// With no image it says so, and pastes nothing.
	clipboardImage = func() []byte { return nil }
	a.pasteImageFocused()
	if !p.dropErr || !strings.Contains(p.dropNote, "No image") {
		t.Errorf("header note %q (error %v)", p.dropNote, p.dropErr)
	}
	if n := count(); n != 1 {
		t.Errorf("%d images after an empty clipboard", n)
	}
}

func TestDebugDrop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell prompt of the test is Unix's")
	}
	a, tt := newTestApp(t)
	p := a.tab().Focus
	waitFor(t, tt, "the shell", func() bool { return p.term != nil && strings.Contains(p.term.Text(), testShell.prompt) })
	dir := t.TempDir()
	a.debugDrop(filepath.Join(dir, `my\ shot.png`) + " " + filepath.Join(dir, "b.png"))
	waitFor(t, tt, "both paths", func() bool {
		s := flat(p)
		return strings.Contains(s, escapeUnix(filepath.Join(dir, "my shot.png"))) && strings.Contains(s, filepath.Join(dir, "b.png"))
	})
}
