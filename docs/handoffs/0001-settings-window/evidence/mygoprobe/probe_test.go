package probe

import (
	"testing"
	"time"

	"github.com/egoist/mygo"
)

func within(t *testing.T, name string, fn func()) {
	done := make(chan struct{})
	go func() { fn(); close(done) }()
	select {
	case <-done:
		t.Logf("%s: returned", name)
	case <-time.After(3 * time.Second):
		t.Logf("%s: BLOCKED (>3s)", name)
	}
}

func TestProbe(t *testing.T) {
	var light, dark *mygo.MenuItem
	m := mygo.NewMenu([]*mygo.MenuItem{{Label: "Appearance", Submenu: []*mygo.MenuItem{
		{Label: "Light", Type: mygo.MenuItemRadio, Checked: true},
		{Label: "Dark", Type: mygo.MenuItemRadio},
	}}})
	sub := m.Items()[0]
	_ = sub
	within(t, "MenuItem.SetChecked after NewMenu", func() {
		items := m.Items()
		_ = items
	})
	// find the radio items through ItemByID-less traversal: rebuild with kept pointers
	light = &mygo.MenuItem{Label: "Light", Type: mygo.MenuItemRadio, Checked: true}
	dark = &mygo.MenuItem{Label: "Dark", Type: mygo.MenuItemRadio}
	mygo.NewMenu([]*mygo.MenuItem{{Label: "Appearance", Submenu: []*mygo.MenuItem{light, dark}}})
	within(t, "SetChecked(dark)", func() { dark.SetChecked(true) })
	t.Logf("after SetChecked: light=%v dark=%v", light.IsChecked(), dark.IsChecked())
	within(t, "Theme.SetSource(dark)", func() { mygo.Theme.SetSource(mygo.ThemeDark) })
	within(t, "(&mygo.Window{}).Focus()", func() { (&mygo.Window{}).Focus() })
	within(t, "(&mygo.Window{}).Invalidate()", func() { (&mygo.Window{}).Invalidate() })
}
