package main

import (
	"runtime/debug"

	"github.com/egoist/mygo"
)

// The About panel shows the version of mygo.json twice, as "0.1.0 (0.1.0)":
// the version is the same for every build until a release. The parentheses
// show the commit the build was made from instead.

// buildCommit tells the commit a build was made from, as `go build` stamped
// it: seven characters of the revision, with "+dirty" when the tree had
// changes, and "dev" when the build carries no version control information.
func buildCommit() string { return commitOf(debug.ReadBuildInfo()) }

// commitOf is buildCommit's formatting, for a test.
func commitOf(info *debug.BuildInfo, ok bool) string {
	if !ok {
		return "dev"
	}
	var rev string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	rev = rev[:min(7, len(rev))]
	if dirty {
		rev += "+dirty"
	}
	return rev
}

// showAbout opens the standard About panel with the commit in place of the
// build number; the name, the version and the copyright are the bundle's.
func showAbout() {
	showAboutPanel(mygo.AboutPanelOptions{Version: buildCommit()})
}

// showAboutPanel opens the panel; a variable for tests, which cannot open one.
var showAboutPanel = func(o mygo.AboutPanelOptions) { mygo.App.ShowAboutPanel(o) }
