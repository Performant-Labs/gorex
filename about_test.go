package main

import (
	"runtime/debug"
	"testing"
)

func TestCommitOf(t *testing.T) {
	set := func(kv ...string) *debug.BuildInfo {
		b := &debug.BuildInfo{}
		for i := 0; i < len(kv); i += 2 {
			b.Settings = append(b.Settings, debug.BuildSetting{Key: kv[i], Value: kv[i+1]})
		}
		return b
	}
	for name, tc := range map[string]struct {
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		"clean":       {set("vcs.revision", "9be1911d32fc3253127170c77a9305bd43ca9a11", "vcs.modified", "false"), true, "9be1911"},
		"dirty":       {set("vcs.revision", "9be1911d32fc3253127170c77a9305bd43ca9a11", "vcs.modified", "true"), true, "9be1911+dirty"},
		"short":       {set("vcs.revision", "abc"), true, "abc"},
		"no vcs":      {set("-trimpath", "true"), true, "dev"},
		"no build id": {nil, false, "dev"},
	} {
		if got := commitOf(tc.info, tc.ok); got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
	// The test binary itself is built by `go test`, which stamps no VCS
	// information: not an empty string either way.
	if buildCommit() == "" {
		t.Error("buildCommit is empty")
	}
}
