# Handoff-T-green: Phase 7 - Settings window, General and Default's text size, settings v2

**Date:** 2026-11-09
**Branch:** 0001-settings-window
**Issue:** #1 (delivery 1)
**Handoff-F reviewed:** docs/handoffs/0001-settings-window/handoff-F.md
**Handoff-T-red:** docs/handoffs/0001-settings-window/handoff-T-red.md

## GREEN confirmation

F reported "Tests that look wrong (for T): None", so no test was changed. `settings_test.go` is untouched.

    $ go test -count=1 -race ./...
    ok  	gorex	5.291s
    ok  	gorex/internal/rex	1.424s
    ?   	gorex/tools/mkicon	[no test files]
    $ go test -count=3 -race .      # flakiness check
    ok  	gorex	13.367s

All 14 darwin-runnable new tests and the 5 existing ones pass. `TestSettingsKeysOnWindows` skips here (Windows only).

**Spot-check that the tests pin behaviour.** In a scratch copy of the tree (session scratchpad, not the worktree), I removed one behaviour at a time and ran `go test .`. Each mutation failed the named tests, and the copy was restored after each:

| Behaviour removed | Failing test |
|---|---|
| the write gate (`saveSettings` ignores `readOnly`) | `TestSettingsReadOnly` |
| the 8-32 clamp | `TestSettingsMigratesV1`, `TestSettingsV2Load` |
| the `a.win == nil` guard in `quitAndEnd` | `TestEndAllNeedsMainWindow` |
| the menu sync (`SetChecked` in `settingsChanged`) | `TestSettingsAppearanceInStep` |
| Close Pane from Settings closing Settings | `TestMenuClicksFromSettings` |
| Bigger Text being app-wide | `TestCommandsFromSettings`, `TestMenuClicksFromSettings` |
| restoring Default's name | 6 tests, among them `TestSettingsV2Load` |

## Tier 1 results

| Check | Command | Result |
|---|---|---|
| gofmt | `gofmt -l .` | no output, PASS |
| vet, 3 platforms | `go vet ./... && GOOS=linux go vet ./... && GOOS=windows go vet ./...` | PASS |
| cross-compile tests | `GOOS=windows go test -c -o /dev/null . && GOOS=linux go test -c -o /dev/null .` | PASS |
| build | `go build ./...` | PASS |
| suite, race | `go test -count=1 -race ./...` | PASS (matches F's report) |
| server starts / API smoke | `go test` runs the session server in `TestEndAllNeedsMainWindow` and the existing suite (socket, panes) | PASS |

No discrepancy with F's reported results.

## Tier 2 results

- **Test coverage per criterion:** PASS (table below). Criterion 11's Dock reopen and ⌘1-9 doing nothing from Settings have no headless seam; they are U's.
- **Test quality (test-quality.md §7):** PASS. Each test names a behaviour and fails alone for a feature assertion, and the mutation table shows they fail when the behaviour goes. They sit at unit/headless tier. 15 test functions for ~1300 changed lines is proportionate, and I found none to delete or merge.
- **Type safety:** PASS. No `any` casts in the changed files. `profile.FontSize` is `*float32`, so a missing value is representable (§7).
- **Error handling:** PASS. Newer, invalid and unreadable files, a failed backup and a missing file are tested (`TestSettingsReadOnly`, `TestSettingsFirstLaunch`). Advisory: `saveSettings` ignores a failed write (unchanged behaviour, F noted it).
- **Data integrity:** PASS. The v1 backup is written before the v2 rewrite, and a failed backup leaves the v1 file alone. A retry overwrites an old backup (`TestSettingsMigrationOverwritesOldBackup`). `loadSettings` resets `prefs` and `readOnly` first, so nothing leaks between loads.
- **API contract / spec:** PASS. The v2 JSON shape matches §1. The footer strings, the `"%v pt"` value and the menu placement match the brief.
- **Security:** N/A beyond the files being written 0o600 as before.
- **Migration safety:** PASS. It is not reversible in place, but `settings.v1.json` holds the original bytes.
- **Playwright:** the project does not use Playwright. N/A.
- **Evidence appendix:** added one entry (`evidence.md`, "T (Phase 7)"): the slider `Step` fact that `TestSettingsTextSize` relies on, with a verbatim excerpt of mygo `ui/base.go:122-124`.
- **Cross-check of F's results:** I re-ran F's commands; the results are identical.

## Acceptance criteria status

| # | Result | Backed by |
|---|---|---|
| 1 | PASS | `TestSettingsFirstLaunch` |
| 2 | PASS (darwin/headless); Windows Ctrl+, in Settings only compiled | `TestSettingsMenuItem`, `TestSettingsOpensOnce`, `TestSettingsKeysOnWindows` (skipped) |
| 3 | PASS | `TestSettingsAppearanceInStep` |
| 4 | PASS | `TestSettingsTextSize` |
| 5 | PASS | `TestSettingsMigratesV1` (too_big, too_small), `TestSettingsV2Load` |
| 6 | PASS | `TestSettingsMigratesV1`, `TestSettingsMigrationOverwritesOldBackup` |
| 7 | PASS | `TestSettingsV2Load` |
| 8 | PASS | `TestSettingsSurviveRestart`, `TestSettingsReadOnly` |
| 9 | NOT RUN (Windows only) | `TestSettingsKeysOnWindows` compiles and vets for windows |
| 10 | PASS for the command window check and Close Pane; Show Tab 1-9 not reachable headless | `TestCommandsFromSettings`, `TestMenuClicksFromSettings`, `TestEndAllNeedsMainWindow` |
| 11 | PARTIAL: Settings opens with no main window (`TestSettingsOpensOnce`); the `OnActivate` change is read in `main.go` (`if a.win == nil { a.open() }`), not run | U live |
| 12 | PASS | the suite and vet above |
| 13 | PASS | `TestReadmeSettings` |

## Blocking issues

None.

## Advisory notes

- Coverage holes for U to walk live: the Dock reopen with only Settings open; ⌘1-9, ⌘D and ⌘K doing nothing from Settings; the Appearance selection ring and the live preview's look; Linux Ctrl+W (`RoleClose`); opening Settings from the palette (`NewWindow` inside the main view's frame); and everything Windows-specific (Esc, Ctrl+W, Ctrl+Shift+W, Ctrl+, in Settings), since `TestSettingsKeysOnWindows` has never run on a Windows host.
- An empty (0-byte) `settings.json` is now "could not be read": read-only for the run, and the file is left alone. Today it would be overwritten. That matches "an invalid file must not be destroyed".
- `saveSettings` ignores a failed write, so the footer would still say changes save. Unchanged from today and not in the brief's three reasons.
