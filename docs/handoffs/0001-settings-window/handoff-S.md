# Handoff-S: Phase 8 - Settings window, General and Default's text size, settings v2

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Issue:** #1 (delivery 1)
**Handoff-T reviewed:** docs/handoffs/0001-settings-window/handoff-T-green.md (and handoff-T-red.md)
**Handoff-A reviewed:** docs/handoffs/0001-settings-window/handoff-A.md (round 5, PASS) and handoff-A-dup.md (anti-duplication, PASS)
**Handoff-F reviewed:** docs/handoffs/0001-settings-window/handoff-F.md
**Spec:** docs/handoffs/0001-settings-window/brief.md (at `6238bf7`), issue #1, wireframes `docs/wireframes/settings-profiles/1-general.png` and `2-profile-overrides.png`
**Operator-facing report:** [`handoff-S-report.html`](handoff-S-report.html) (side-by-side and wipe-slider comparison of the live render and the wireframes)

## A precondition

Confirmed. handoff-A.md (round 5) returned PASS, with 3 advisory warns that F handled (decisions in handoff-F.md "Design decisions"). handoff-A-dup.md returned PASS with 2 minor warns and no block.

## T precondition

Confirmed. handoff-T-green.md reports "Blocking issues: None". I re-ran `go build ./...`, `go vet ./...` and `go test -count=1 -race ./...` in the worktree: all pass, matching T's report. I did not re-run T's Tier 1/2 matrix beyond that.

## Visual-diff-tool precondition (visual / web projects only)

**Not applicable as written; substituted. Read this paragraph.** GoRex is a native desktop app (Go and the `mygo` UI toolkit). It has no web surface: no URL to open, no DOM, no CSS breakpoints at 360/768/1280. The Claude-in-Chrome tools are not in my tool list, and Playwright cannot drive a native window, so the browser-tool and Playwright preconditions cannot be met for this project at all. The role document limits them to "visual / web projects", and the pipeline's verdict schema has no CANNOT-AUDIT value, so I did not stop. I did the nearest honest equivalent, headless and with nothing opened on the desktop:

- In a scratch copy of the tree (session scratchpad, not the worktree, nothing committed except the PNGs below), I rendered `App.settingsView` with `ui.NewTester(...).Image()`, the same renderer F used. Pages: General and Default, light and dark, at the default window size 760×560, and Default at the minimum window size 640×460.
- I compared those renders with wireframes 1 and 2 by eye, side by side (`+append`), and read each PNG.
- ImageMagick `compare -metric AE` is on PATH (`/opt/homebrew/bin/compare`), but **I did not compute a pixel-delta percentage**. The wireframes are annotated SVG mock-ups, drawn at another aspect ratio (1.57 against 1.36), in another typeface, with a macOS title bar and callout badges. A pixel count against them would be noise, and the role's 2% / 5% thresholds do not mean anything here. The brief itself says to read wireframe 1 with its footer ignored. So the comparison below is by element, not by pixel count.
- What this does not cover, and U must walk live: the native title bar and traffic lights, the Appearance selection ring under the real theme, real slider dragging, and how the window sits next to the main window.

## Visual diff results (visual / web projects only)

No pixel-delta table: see the paragraph above for why. Evidence files, all under `docs/handoffs/0001-settings-window/screenshots/cycle-1/`:

| View | Live render | Wireframe (cropped to the window's content) | Side by side |
|---|---|---|---|
| General, light, 760×560 | `s-general-760-live-20261109.png` | `s-general-760-reference-20261109.png` | `s-general-760-composite-20261109.png` |
| General, dark, 760×560 | `s-general-dark-760-live-20261109.png` | none (no dark wireframe) | n/a |
| Default, light, 760×560 | `s-default-760-live-20261109.png` | `s-default-760-reference-20261109.png` (wireframe 2, a non-Default profile: only the text-size row and preview apply) | `s-default-760-composite-20261109.png` |
| Default, dark, 760×560 | `s-default-dark-760-live-20261109.png` | none | n/a |
| Default, light, 640×460 (minimum size) | `s-default-640-live-20261109.png` | none | n/a |

### Per-section delta description

| Section | Views | What differs from the wireframe | Documented by F? | Status |
|---|---|---|---|---|
| Sidebar: General, PROFILES, Default "built-in" | all | Matches: same order, same labels, General highlighted on the General page, Default highlighted on the Default page. No "+ New Profile…" footer, no counts, as the brief says (delivery 2). | Yes (brief §2) | MATCH |
| General title and line, rule | General | Matches in wording and layout. | n/a | MATCH |
| Appearance choices | General | Three swatches (System half-and-half, Light, Dark) with labels; System has the blue selection ring on first launch, as wireframe 1. The dark half is a neutral near-black (the terminal's dark background) where the wireframe draws a purple-grey one. | Yes (F: reuses `darkTerm.Background`, no new colour) | MATCH (colour shade differs, intentionally) |
| "Later: notifications, …" annotation | General | Absent in the live window. It is a note about future work in the mock-up. | Yes (handoff-F "Deviations") | MATCH |
| Text size row | Default | Title "Default", row "Text size" with − slider + and "11.6 pt". The 8 and 32 tick labels under the slider are not shipped. The wireframe's override marker, "Overrides Default" text, Reset link, Default tick on the track and the "…" menu are delivery 2. | Yes (handoff-F "Deviations") | MATCH (as scoped) |
| Live preview | Default | Three lines of terminal text (`~/Projects/gorex main`, `$ go test ./...`, `ok   gorex  0.42s`) in a box that follows light and dark, at 11.6 pt. In the wireframe the text is larger only because that profile is at 18 pt. | n/a | MATCH |
| Footer | all | "Changes apply and save as you make them." at the bottom of the page. | n/a | MATCH |
| Minimum window size 640×460 | Default | The slider shortens and the value stays visible, nothing clips or overlaps. | n/a | MATCH |

I looked for gestalt problems the way the role asks: overlapping elements, anything hidden, anything clipped, an unreadable dark mode. I found none. The dark renders read clearly; the selected swatch ring, the highlighted sidebar row, the slider fill and the preview box all show in both appearances.

## Spec compliance

Checked against brief.md, section by section and against the 13 acceptance criteria, by reading the code, running the suite, and the probes noted below.

| # | Requirement | Compliant | How checked / notes |
|---|---|---|---|
| 1 | First launch: General, Appearance System, Default "built-in" | YES | `TestSettingsFirstLaunch`; render of General at first launch. |
| 2 | ⌘, / Ctrl+, / menu item / palette open Settings or focus the open one; never a second | YES (darwin and headless); Windows Ctrl+, code-read only | `TestSettingsMenuItem`, `TestSettingsOpensOnce`. `openSettings` focuses `a.settingsWin` when set. Windows: Settings' only Ctrl+, path is the app-wide loop in `settingsKeys` (`preferences.go`); I read it and it matches §2. The main window's Windows path is the existing `shortcuts()` loop, which now sees `cmdSettings`. |
| 3 | Appearance from Settings changes the app, and the menu and Settings follow the menu and palette | YES | `TestSettingsAppearanceInStep`. Both writers end in `settingsChanged()`: `SetChecked` on the kept radio items, `Invalidate` on the Settings window. System is spelled `"system"` in memory, file, menu and palette. |
| 4 | Text size: slider, −, + whole points; saved; ⌘+ ⌘− ⌘0 as today; Settings follows | YES | `TestSettingsTextSize`. − from 11.6 gives 11 and + gives 12 (`ceil-1`, `floor+1`); ⌘+ keeps today's ±1. I probed that opening Default's page does not change the size or write the file (the slider does not snap a non-whole value on open): passes. |
| 5 | 8–32 clamp everywhere, except a v1 file with no size (11.6) | YES | `TestSettingsMigratesV1`, `TestSettingsV2Load`; `clampFontSize` is the one clamp, used by `complete()` and `setFontSize`. |
| 6 | v1 migrates at load; byte-for-byte backup; `""` becomes `"system"` | YES | `TestSettingsMigratesV1`, `TestSettingsMigrationOverwritesOldBackup`. The backup is written first and the v2 rewrite follows in the same `loadSettings`. |
| 7 | Default cannot be removed or renamed by editing the file | YES | `TestSettingsV2Load`; I also probed `profiles.default: null`, a missing size, a size of 0 (becomes 8) and no `profiles`: all restored, always named "Default". |
| 8 | Survives restart; missing, invalid and newer files never crash or lose the file, including after a later change; failed backup leaves the v1 file; footer says why | YES | `TestSettingsSurviveRestart`, `TestSettingsReadOnly` (3 cases). `saveSettings` returns when `readOnly != ""`, for the whole run. Footer strings match the brief verbatim. I also probed `[]`, `""`, `{"version":"2"}`, `1e999`: each is "could not be read", no crash. |
| 9 | Windows: Esc, Ctrl+W, Ctrl+Shift+W close Settings; Ctrl+, keeps one window | NOT RUN; code-read | `TestSettingsKeysOnWindows` skips on darwin and has never run on Windows. `settingsKeys` does what §2 says, and it compiles and vets for `GOOS=windows`. Left to U. |
| 10 | From Settings, close shortcuts close only Settings; ⌘D, ⌘K, ⌘1–9 do not reach the hidden main window; New Tab brings the main window forward | YES (headless part); Show Tab 1–9 code-read | `TestCommandsFromSettings`, `TestMenuClicksFromSettings`. Show Tab 1–9 are now built with `item()`, so they pass through `click` → `runsIn`; I read that path. Close Pane from Settings goes to `closeWindow`. |
| 11 | Settings works with the main window closed; Dock click with only Settings open reopens the main window | PARTIAL: code-read | `TestSettingsOpensOnce` runs Settings with no main window. `main.go` now reads `if a.win == nil { a.open() }`, which is what §4 prescribes; no seam to run it. Left to U. |
| 12 | `go vet ./...` and `go test -count=1 -race ./...` pass; testable behaviour has failing-if-removed tests | YES | I ran both. T's 7 mutations show the tests fail when the behaviour goes. |
| 13 | README: Ctrl+, in the Shortcuts table, named with Ctrl+= Ctrl+- Ctrl+0 as a Shift-rule exception; Layout lists the new file | YES | `TestReadmeSettings`; I read the README diff. `preferences.go` is a one-word lowercase name, not `settings.go`. |

Other brief requirements checked:

- **§1** one v2 format; `profile.FontSize` is `*float32` so a missing field is representable (§7); `"default"` is a named constant; a failed backup is read-only with the v1 file untouched; a newer-version file is read as well as possible and never written.
- **§2** sidebar, General and Default pages, copy, no OK or Apply, footer wording, Windows keys: as above. The menu places Settings… after About on darwin and under Edit elsewhere.
- **§3** `setAppearance` and `setFontSize` are the only writers. Settings never touches `prefs`, `termFont` or `saveSettings`: I grepped `preferences.go`, it only reads `prefs` and calls the writers.
- **§4** one `AppWide` field and one helper (`runsIn`); `item()`, the Show Tab items and `run` use it; Quit and End All Sessions guards in `quitAndEnd` only; a nil click window works; Close Pane from Settings closes Settings; Settings' own queue is not used.
- **§5** one preview terminal per `App`, made lazily, kept outside `reset()`; `SetFont` on change; no second copy of the terminal colours; no new colours in `colors`. `newSettingsWindow` is set in `init` to break an initialization cycle: a documented, justified departure from "a package-level func variable" that keeps the seam's behaviour.
- **§6** seams: `newSettingsWindow`, `focusWindow`, `closeWindow`, `setThemeSource`; the pure `parseSettings` with a thin `loadSettings`; the write gate carries its reason.
- **Source-of-truth sanity check (the two minutes before the visual pass):** the brief and wireframes show no defect that warrants ADVISORY-HOLD. One unverified claim in the brief (Ctrl+, sends nothing in legacy terminal encoding) is flagged as unverified in the brief itself. One convention note, advisory only: see "Accessibility" below.

## Quality audit

| Area | Result | Notes |
|------|--------|-------|
| API consistency | N/A | No API. The session-server protocol is untouched. |
| Error handling | PASS | Missing, invalid, unreadable, empty, newer-version and failed-backup cases all end in a defined state with no crash and no lost file. Advisory: `saveSettings` still ignores a failed write (disk full), so the footer would claim "save" when a write failed. Unchanged from today and outside the brief's three reasons; F and T both noted it. |
| UI/UX match to spec | PASS | Elements, copy and order match wireframes 1 and 2 as scoped (see the table above). Differences are documented deviations. |
| Accessibility | PASS with advisories | Slider is labelled "Text size"; − and + are labelled "Smaller Text" and "Bigger Text" with tooltips; the Appearance choices are radios named by their text. Advisory 1: the sidebar's General and Default entries are plain clickable rows with no role and no selected state (the main window's tabs use `ui.RoleTab` and `.Selected`). I did not run the window live to see whether they take keyboard focus. Advisory 2: the app's `textFaint` grey (#8e8e93) on the light panel is about 3.2:1, under 4.5:1, for the 12 px footer, the "PROFILES" heading and "built-in". It is the app's existing token and the wireframe uses a similar grey; in dark it is about 4.6:1. Neither is a spec requirement. |
| Architecture gate | PASS | A round 5 PASS, A-dup PASS. No quality finding here conflicts with A. |
| Code organization | PASS | `settings.go` holds the format, migration and writers; `preferences.go` the window; `commands.go` the table and one window check. Comments explain why (the initialization cycle, the preview terminal). No dead code, commented-out code or TODOs in the diff. `hasVisibleWindows` is an unused parameter of the `OnActivate` closure, as the brief specifies. |
| Security | PASS | Input is only `settings.json`, parsed with `encoding/json` and clamped; nothing is executed or interpolated from it. Files are written 0o600, including the backup. No secrets. |
| Performance | PASS | Settings calls `SetFont` only on a change; the one preview terminal is reused, so the never-ending `nopConn` reader goroutine is created once per app, not per open. |
| Visual regression | N/A (no Playwright, no web surface) | Replaced by the headless render comparison above. No baselines exist in this repo. |
| Naming consistency | PASS | `preferences.go` (one word, lowercase), `AppWide`, `settingsView`, `openSettings`, `cmdSettings` follow the file's conventions. The window's visible name is "Settings", as the wireframe. |
| Test quality (`testing/test-quality.md` §7) | PASS with advisories | 15 test functions in `settings_test.go` for about 1,300 changed lines is proportionate to a feature story. Each fails in isolation on a feature assertion (T's RED output) and T's 7 mutations confirm they fail when the behaviour goes. They sit at the cheapest sufficient tier (plain `go test` with `ui.Tester`, temp dirs, recorders). They assert behaviour (file contents, theme applied, menu checked, texts shown), not implementation. Advisories, none a "delete" finding: (a) `TestSettingsMenuItem` checks four behaviours in one function (menu placement, the shortcut string, command-table membership, the palette's key hint): a failure names one assertion, but splitting it would give clearer signals; (b) `TestReadmeSettings` finds the table and prose by string index (`"\| ⌘T"`, `"## Develop"`), so it breaks on an unrelated README restructuring, though it can fail for the right reason; (c) `TestSettingsKeysOnWindows` has never run on a real Windows host; (d) the "newer version" read-only case does not pin which size is loaded (the brief says "what it could read"). |

## Scope check

F delivered delivery 1 and nothing from delivery 2: no "+ New Profile…", no counts, no per-pane profiles, no override or Reset UI, and `layout.json` is unchanged. The changes outside the Settings window all come from the brief: `AppWide` and `runsIn`, the Show Tab items through `item()`, `quitAndEnd`'s guard, `OnActivate`, the README, and the keep-in-step sync. `debug.go:70` calls `a.run`, so it inherits the new `run` (New Tab also focuses the main window there); F flagged this and it is harmless. Nothing under-delivered except the criteria that cannot run headless (9, parts of 10 and 11), which U has to walk live.

Process note: some tool output in this session showed `[REDACTED]` in place of string literals in `commands.go`, `main.go` and `preferences.go` (for example `StateKey`). I checked the files with `grep -c REDACTED` (0 hits) and `xxd`: the files contain the real values (`StateKey: "settings"`). It was a display artefact, not a defect.

## Verdict

PASS — all acceptance criteria are met or, where they cannot run headless (9, parts of 10 and 11), the code reads as the brief requires and is left to U. The implementation is spec-compliant, the suite is green, the quality is acceptable. Ready for O to commit.

Verification I did myself: `go build ./...`, `go vet ./...` and `go test -count=1 -race ./...` in the worktree (all pass); a headless render of both pages in light and dark, and at the minimum window size, in a scratch copy; a parser probe (`null`, `[]`, `{}`, `profiles.default: null`, `appearance: "auto"`, `version: "2"`, size 0, `1e999`, empty file) and a probe that opening Default's page changes nothing on disk. Both probes were in the scratch copy and are not committed.

## Advisory notes

Non-blocking, for O to carry to U or to delivery 2:

1. **U must walk these live** (no headless seam): the Dock reopen with only Settings open; ⌘1–9, ⌘D and ⌘K doing nothing from Settings; the Appearance selection ring and the preview under the real theme; Linux Ctrl+W (`RoleClose`); opening Settings from the palette (`NewWindow` inside the main view's frame); and everything Windows-specific (Esc, Ctrl+W, Ctrl+Shift+W, Ctrl+, in Settings, Ctrl+Shift+T with no main window), since `TestSettingsKeysOnWindows` has not run on a Windows host.
2. Give the sidebar entries a role and a selected state (`ui.RoleTab` and `.Selected(...)`, as `tabs.go` does) before delivery 2 adds more entries. Check keyboard access when U runs it.
3. A settings file that is the JSON `null` is treated as v1: it is backed up byte for byte, then rewritten as the defaults. Nothing is lost. Treating it as unreadable would leave it alone; either is defensible.
4. `saveSettings` ignores a failed write. If delivery 2 adds more to the file, consider surfacing it in the footer's read-only state.
5. Consider splitting `TestSettingsMenuItem`, and making `TestReadmeSettings` less dependent on README layout.
6. The footer, "PROFILES" and "built-in" use the faint grey, about 3.2:1 on the light panel. Worth a look in a later pass on the whole app's colour tokens, not just this window.
7. Editorial, carried from A: the brief has two `### 8.` headings.
