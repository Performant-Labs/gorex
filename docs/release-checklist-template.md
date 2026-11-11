# Release checklist (template)

**This is the reusable template. It carries no version number.** For each release, copy this
file's content into a new GitHub issue titled `Release checklist: vX.Y.Z` (the real version),
and fill in and tick the boxes on that copy, never here. Where a step has a `___`, write the
result there: a result you can read later is what makes "done" verifiable.

The repository today (checked when this template was written): the version is `"version"` in
[`mygo.json`](../mygo.json) and nowhere else; there are no git tags and no GitHub Releases;
there is no `CHANGELOG.md`; `main` is not a protected branch; CI is
[`.github/workflows/test.yml`](../.github/workflows/test.yml) (`go vet` and the tests with the
race detector on Linux and macOS, and on Windows with no test allowed to skip), and nothing
publishes anything. Where this template says "today", it describes that state; update it when
the state changes.

## Decide before the first release

The repository does not say these things, so the first release stops at each line until it
records a decision. Once decided, replace "Decision: ___" with the decision, and the step that
depends on it stops being conditional.

- [ ] **D1. Signing and notarization.** Today `mygo.json` sets neither, so `go tool mygo build`
  signs ad hoc (`codesign -dv` shows `Signature=adhoc`, `TeamIdentifier=not set`) and macOS
  warns whoever opens the app. `macos.signingIdentity` (a Developer ID) and `macos.notarize` in
  `mygo.json` change that. Decision: `___`
- [ ] **D2. Where releases are published**: GitHub Releases on `Performant-Labs/gorex`, the
  upstream repository, or somewhere else; and whether a Homebrew cask is wanted. Decision: `___`
- [ ] **D3. Auto-update.** MyGo can make signed update archives and `update-<platform>.json`
  (the `updates` setting in `mygo.json`, and the key of `go tool mygo keygen` in
  `MYGO_UPDATER_PRIVATE_KEY`), to be published where the app's updates point. Not configured
  today. Decision: `___`
- [ ] **D4. Platforms that ship.** `go tool mygo build` builds `darwin/arm64` by default
  (`-platform` takes more: `darwin/universal`, `linux/amd64`, `windows/amd64`...). CI tests
  Linux, macOS and Windows. Decision: `___`
- [ ] **D5. A `CHANGELOG.md`, or release notes only.** None exists. Decision: `___`
- [ ] **D6. Tag signing** (`git tag -s`), and which key. Decision: `___`

## Pre-flight

- [ ] 1. **Identify the release commit** (usually `main`'s tip)
  - Commit SHA: `___`
- [ ] 2. **CI is green on that exact commit**, in all three jobs: `test (ubuntu-latest)`,
  `test (macos-latest)`, `test (windows-latest)`
  - `gh run list -R Performant-Labs/gorex --workflow test.yml --commit <sha>`
  - Link the specific run, not "CI passes in general": `___`
- [ ] 3. **No real, ready work left unmerged**
  - `gh pr list -R Performant-Labs/gorex --state open`
  - `main` is not protected, so nothing stops a direct push: everything in this release,
    including the version bump (step 14), goes through a pull request by convention
- [ ] 4. **If tags are signed (D6): the tag can be signed and GitHub can verify it.** Check now,
  not at step 18
  - The machine that will tag holds the secret key: `gpg --list-secret-keys <key id>`
  - The key's public half is on the GitHub account that owns the tag: `gh api user/gpg_keys --jq '.[].key_id'`
  - The tagger email is a verified email of that account (`gh api user/emails`) and a user id
    of the key, or GitHub shows the tag as "Unverified"

## Testing beyond CI

CI proves `go vet` and `go test` pass. It cannot run what needs a remote machine or a window.

- [ ] 5. **The unit suite on the release commit, locally**, as CI runs it
  - `go vet ./... && go test -count=1 -race ./...` (without `-race` on Windows)
  - Result (the `ok` lines): `___`
- [ ] 6. **The real-remote tests, against the newest Claude Code and opencode** (for releases
  that include the image drop and paste feature, [#8](https://github.com/Performant-Labs/gorex/issues/8);
  before it, write "not in this release")
  - Look up the newest versions today: opencode with `brew info opencode`; Claude Code with
    `npm view @anthropic-ai/claude-code dist-tags.latest` (`claude --version` is only what is
    installed here). Record them: `___`
  - The remote host must reach this Mac by key without a password, have both agents installed,
    Claude Code signed in and trusting the home directory
  - `GOREX_TEST_SSH=<host> go test -count=1 -v -run 'ToRemote|IntoRemote' ./...`. One test sends
    one short model turn to Claude Code
  - **An all-skipped run is not a pass.** These tests skip when `GOREX_TEST_SSH` is unset, and
    `go test` still exits 0. Read the `-v` output: every test must say `--- PASS`, none
    `--- SKIP`. Paste the lines: `___`
  - If an agent changed its version since the last release, say whether the paste form still
    attaches the image (`[Image #1]` in Claude Code, `[Image 1]` in opencode). A new agent
    version that stops attaching is a red test (step 9)
- [ ] 7. **The native-window checks that CI never sees**, on the app *built in step 19*, not on
  `go tool mygo dev` (the build is what ships; this step is done after step 19, with the
  result written here)
  - Tabs, splits, zoom, and dragging the gaps between panes
  - Quit with ⌘Q and reopen: every tab and pane comes back, programs still running. Then ⌥⌘Q:
    sessions end
  - A Finder drop onto a local pane pastes the path; onto a pane running `ssh`, the file is
    copied and the remote path pasted. ⌘V with a screenshot on the clipboard does the same
    (#8). Verified in the native window, not only headlessly: `___`
  - Light, dark and System appearance; text size (⌘+ ⌘− ⌘0); the Settings window (⌘,)
  - A long command finishing while the window is in the background posts a notification
  - **Upgrade over the previous release with sessions running** (see step 15 for what to
    expect): `___`
- [ ] 8. **If Linux or Windows builds ship (D4): run the app once on each.** CI tests them; it
  does not open the app. Not shipped? Write "not shipped: D4"
- [ ] 9. **Every failure is accounted for**: fixed, or knowingly overridden. The override is
  recorded here (which test, why), and also under Known issues if users will see it: `___`

## Known issues

This is a young project, so shipping with documented issues is acceptable. They must be checked
and named, not left out. Gathered here, **before** the release notes are written below.

- [ ] 10. **Skim the open `bug` issues**
  - `gh issue list -R Performant-Labs/gorex --label bug --state open`
- [ ] 11. **Write one line per real, still-open issue relevant to this release, with a link**
  - This exact list goes into the release notes in step 16, verbatim: `___`

## Version and release notes

**Steps 12 to 17 happen on ONE branch, `release/vX.Y.Z`**: one pull request, one merge.

- [ ] 12. **Decide the version** (`X.Y.Z`) from everything merged since the last tag (or, for
  the first release, since the start)
- [ ] 13. **Bump `"version"` in `mygo.json`** to `X.Y.Z`
  - It is the only place the version lives. Check: `git grep -n '<previous version>'` should
    find only what is meant to keep it (release notes)
- [ ] 14. **Say what an upgrade does to the session server**, and put it in the release notes
  - `rex.ProtocolVersion` in `internal/rex/proto.go` changes whenever the wire protocol does.
    Did it change since the last release? `___`
  - If it changed: the new app shows "The session server is of another version of GoRex: quit and
    end all sessions to restart it." until the user quits with ⌥⌘Q and reopens, which ends
    their sessions. The notes say so, up front
  - If it did not: the new app talks to the **old** session server. Only a development build
    restarts a stale server (`main.go`, `staleServer`); a production build never does. The
    server's own fixes reach the user only once sessions end and the server restarts. The notes
    say so when the release contains a change under `internal/rex`
- [ ] 15. **Write the release notes** (D5: in `CHANGELOG.md`, or in the draft Release only)
  - Breaking changes and upgrade notes first (step 14), then what is new and fixed, then
    **Known issues: the list from step 11, verbatim**
  - Read it once as a stranger would: does it stand alone?
- [ ] 16. **Update `README.md`** for anything user-facing that changed, or write "nothing
  README-worthy" here: `___`
- [ ] 17. **Push `release/vX.Y.Z`, open the pull request, confirm CI is green on it, merge it**
  - Then confirm CI is green on the **merge commit** on `main`: that commit is the release
    commit for step 18

## Tag and build

> **Order is part of the checklist.** Nothing here starts until steps 1 to 17 are done. Create the tag with git,
> **before** touching the GitHub Releases page: `gh release create` on a tag that does not exist
> yet makes a lightweight tag on the spot.

- [ ] 18. **Tag the merge commit on `main`** (not the commit on the release branch, which `main`
  does not contain after a squash merge)
  - `git tag -a vX.Y.Z -m "GoRex X.Y.Z" <merge commit sha>` (`-s` if D6 says sign)
  - `git cat-file -t vX.Y.Z` must print `tag`; `commit` means a lightweight tag: delete it and redo
  - `git push origin vX.Y.Z`. If signed: `gh api repos/Performant-Labs/gorex/git/tags/$(git rev-parse vX.Y.Z) --jq .verification`
    must say `"verified": true`
- [ ] 19. **Build from the tag, in a clean tree**
  - `git status` is clean and `git describe --tags --exact-match` prints `vX.Y.Z`
  - `go tool mygo build -platform <the platforms from D4>`, with the signing and notarization
    from D1 set in `mygo.json` or `-sign`
  - Output goes to `build/<platform>/` (the directory is ignored by git). On macOS: `GoRex.app`
    and `GoRex X.Y.Z.dmg`. Other platforms' files are named as `go tool mygo build -h` describes
- [ ] 20. **Verify the macOS build before attaching anything**
  - The app's version is the release's:
    `/usr/libexec/PlistBuddy -c "Print :CFBundleShortVersionString" build/darwin-arm64/GoRex.app/Contents/Info.plist` prints `X.Y.Z`
  - `codesign --verify --deep --strict --verbose=2 build/darwin-arm64/GoRex.app` passes, and
    `codesign -dv build/darwin-arm64/GoRex.app 2>&1 | grep -E 'Signature|TeamIdentifier'` shows
    what D1 decided (today: `adhoc` and `not set`)
  - If notarized: `spctl --assess --type execute -vv build/darwin-arm64/GoRex.app` says
    `accepted`, and `xcrun stapler validate "build/darwin-arm64/GoRex X.Y.Z.dmg"` passes
  - No builder's home directory is embedded:
    `strings build/darwin-arm64/GoRex.app/Contents/MacOS/GoRex | grep -c "$HOME"` prints `0`
- [ ] 21. **Verify any other platform's build** (D4): the file exists, is named with `X.Y.Z`, and
  the installed app reports `X.Y.Z`. Not shipped? Write "not shipped: D4"
- [ ] 22. **If auto-update is on (D3)**: the update archives and `update-<platform>.json` exist
  for each shipped platform, and `MYGO_UPDATER_PRIVATE_KEY` was the one in use. Off? Write "off: D3"
- [ ] 23. **Now do step 7** on the build from step 19
- [ ] 24. **Write the checksums** from the verified files, in the `<sha256>  <name>` form:
  `cd build/<platform> && shasum -a 256 "GoRex X.Y.Z.dmg" > SHA256SUMS` (add every asset)

## Publish

- [ ] 25. **Create the GitHub Release as a draft on the existing tag** (D2: on the repository decided there)
  - Notes from step 15, assets from steps 19 to 22, and `SHA256SUMS`:
    `gh release create vX.Y.Z --draft --verify-tag -R Performant-Labs/gorex --title "GoRex X.Y.Z" --notes-file <notes file> <assets> SHA256SUMS`
    (`--verify-tag` refuses to create a tag)
- [ ] 26. **Read the draft's page**: assets and checksums present, the notes render, the Known
  issues section is visible

## Post-release smoke checks

Verify what was **uploaded**, not the local build that produced it: a local build passing steps
19 to 23 is not evidence that the uploaded file works.

- [ ] 27. **Download the draft's assets into an empty directory** and check them
  - `gh release download vX.Y.Z -R Performant-Labs/gorex -D <empty dir>` (as the owner, which
    can read a draft; if that fails for a draft, publish first and do this immediately)
  - `shasum -a 256 -c SHA256SUMS` prints `OK` for every asset
- [ ] 28. **Install it as a user would, on a Mac that has never run a development build**
  - Download the `.dmg` in a browser (that gives it the quarantine flag a user's download
    has), open it, drag the app to Applications, open it
  - What macOS shows matches D1: with ad hoc signing it asks the user to approve the app
    in System Settings ▸ Privacy & Security; a notarized build opens
  - GoRex ▸ About, or the built app's version (step 20), says `X.Y.Z`
  - Open a tab, start a program, quit with ⌘Q and reopen: the pane is back. Result: `___`
- [ ] 29. **Publish the draft**: `gh release edit vX.Y.Z -R Performant-Labs/gorex --draft=false --latest`,
  then `gh release list -R Performant-Labs/gorex` shows it as **Latest**
- [ ] 30. **Install it on your own Macs** (replace the app; if step 14 said the protocol changed,
  quit with ⌥⌘Q first and reopen). `GoRex ▸ About` reports `X.Y.Z` on each: `___`
- [ ] 31. **Close out**: link this checklist issue from the GitHub Release, tick every box above,
  and close it
