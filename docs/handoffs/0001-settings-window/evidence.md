# Evidence: issue #1, delivery 1

mygo paths are in the module directory of `github.com/egoist/mygo@v0.3.5-0.20261008155018-bad73d119dc8`, the version `go.mod` pins.

## F (Phase 6)

- **Fact:** `Window.Update` only queues `fn` on the main loop and returns without waiting. So `run`'s `a.do`, called on the zero `mygo.Window` the tests use as the main window, does not hang headless, and its `fn` never runs there.
  **Source:** mygo `content.go:165-170`
  **Verbatim excerpt:**
  > func (w *Window) Update(fn func()) {
  > 	postMain(func() {
  > 		fn()
  > 		w.contentChanged()
  > 	})
  > }

- **Fact:** `SetChecked(true)` on a radio item unchecks its siblings, and `SetChecked(false)` only clears the item. So `settingsChanged` can set all three kept Appearance items in any order.
  **Source:** mygo `menu.go:219-226`
  **Verbatim excerpt:**
  > func (it *MenuItem) SetChecked(v bool) {
  > 	it.update(func() {
  > 		if v && it.Type == MenuItemRadio {
  > 			it.uncheckRadioSiblings()
  > 		}
  > 		it.Checked = v
  > 	})
  > }

- **Fact:** A `nopConn` terminal's reader blocks forever. So the preview terminal is made once per `App` and kept.
  **Source:** `state.go:264`
  **Verbatim excerpt:**
  > func (nopConn) Read([]byte) (int, error)    { select {} }

- **Fact:** The main window's `OnClosed` closes the server connection and clears `a.win`. So `a.win == nil` means "no connection", which `quitAndEnd`'s new guard and the `OnActivate` change rely on.
  **Source:** `main.go:130-131`
  **Verbatim excerpt:**
  > 		a.client.Close()
  > 		a.win = nil

- **Fact:** The main view already closes its own window from inside its frame on Windows (its close button). `settingsKeys` does the same through `closeWindow` on Esc, Ctrl+W and Ctrl+Shift+W.
  **Source:** `view.go:124`
  **Verbatim excerpt:**
  > 			win.Close()

- **Fact:** Windows' main-window loop runs every command with an `Other` shortcut, so `cmdSettings` (in `paletteCommands`, and so in `commands`) gets Ctrl+, there with no hand-written case.
  **Source:** `commands.go:393`
  **Verbatim excerpt:**
  > 			a.later(c, func() { cmd.Run(a) })

- **Fact:** The dev debug hook runs commands through `a.run`, so the change to `run` reaches it too.
  **Source:** `debug.go:70`
  **Verbatim excerpt:**
  > 				a.run(cmd)
