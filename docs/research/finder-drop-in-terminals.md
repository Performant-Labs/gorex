# What terminals send when a file is dragged onto them (Finder drop), and what Claude Code and opencode do with it

Date: 2026-11-11 (all work done the morning of that day, roughly 9:00 to 9:45 AM PST, Pacific Time). Written for
GitHub issue #8 (Performant-Labs/gorex, "Drop or paste images onto a pane and deliver them to Claude Code and
opencode, locally or over ssh"). It complements `docs/research/image-input-claude-code-and-opencode.md`, which
records which paste forms Claude Code 2.1.322 and opencode 2.2.5 accept; that is not repeated here except where a
result is reused.

Every claim carries one label:

- **[SOURCE]** read in the terminal's own code (repo, tag or commit, file, symbol).
- **[DOCUMENTED]** stated in official docs or a changelog (URL and section given).
- **[OBSERVED]** I ran it and saw the result (what was run is stated).
- **[UNVERIFIED]** not settled; what would settle it is stated.

Important limit, stated once: **no real Finder drag onto a real terminal window was performed.** That cannot be done
headlessly, and the standing rule here is not to drive any visible app. So every statement about what a terminal
emits for a drop comes from its source code or docs, never from watching bytes leave the terminal. Where I say
OBSERVED, it means: the exact bytes the source says the terminal emits were reproduced with the terminal's own
code where that was possible (see "How the bytes were reproduced") and then pasted as a bracketed paste into the real
Claude Code and opencode TUIs in a private tmux server, and the attachment chip was counted.

## Verdict table

"Bracketed" means the drop reaches the pty as `ESC[200~ ... ESC[201~` when the foreground program has turned on
mode 2004 (Claude Code and opencode both do; see the earlier research doc). Verdict column: does a dropped image file
attach in Claude Code 2.1.322 / opencode 2.2.5 with no workaround, for a local pane. "Replayed" means the bytes the
source says the terminal emits were pasted into the real TUIs; the drag itself was not run.

| Terminal | Version checked | What is inserted for a dropped file | Bracketed paste? | Several files | Image-specific behaviour | Attaches out of the box? |
| --- | --- | --- | --- | --- | --- | --- |
| Terminal.app | 2.15 (build 488.9), macOS 27.2.1. Closed source | Absolute path appears on the command line [DOCUMENTED]. Quoting, escaping, trailing space: not stated by Apple [UNVERIFIED] | UNVERIFIED. The binary has bracketed-paste mode support, and a drop entry point that appears to share the paste code path, but no drop-time byte was seen | UNVERIFIED | None documented | **UNKNOWN**: nothing verifies the bytes. Needs one real drag (steps below) |
| iTerm2 | v3.7.7 (also the installed app) | Backslash-escaped absolute path plus one trailing space. Characters escaped: `\ ()"&'!$<>;\|*?[]#` and backtick, tab, `{}^+=@~` (and CR, LF) [SOURCE]. Pref to use `"..."` instead, default off [SOURCE] | **Yes**, when 2004 is on, through the paste helper [SOURCE]. Exceptions: Cmd-drag sends `cd <dir>` + CR unbracketed; Option-drag uploads; a pane that shell integration says is on a remote host shows a dialog instead [SOURCE] | One paste, escaped paths joined by one space, one trailing space [SOURCE] | None for image files. A pure image drag with no file URL or string (e.g. from a browser) is refused [SOURCE] | **YES** for a local pane (replayed: attaches in both). Not for a remote-host pane with shell integration, which shows a dialog first |
| Ghostty | v1.4.0 (commit 83fb956); drop code in v1.3.1 and v1.2.3 compared. Not installed here | Backslash-escaped absolute path, no trailing space. Characters escaped: `\ ()[]{}<>"'` + backtick + `!#$&;\|*?` and tab [SOURCE] | **Yes**, when 2004 is on. v1.4.0 via `ghostty_surface_paste_text`; v1.2.3 and v1.3.1 via `insertText` then `sendText` then `ghostty_surface_text`; both end in the same paste encoder [SOURCE] | One paste, escaped paths joined by one space [SOURCE] | None for image files. Only `.string` and `.fileURL` drops are accepted, so an image-only drag is refused [SOURCE] | **YES** (replayed: attaches in both) |
| kitty | v0.50.0 | Plain absolute path. `file://` URLs are decoded to plain paths. NOT escaped or quoted unless the cursor is at a shell prompt with shell integration, in which case each path goes through `shlex.quote` [SOURCE]. No trailing space | **Yes**, when 2004 is on [SOURCE] | **Joined by newline** (not space) when not at a prompt; joined by space when at a prompt [SOURCE] | macOS file promises (typical for a browser image drag) are written to a temp dir and their path enters the same list, so such a drag may yield a path [SOURCE]; not run [UNVERIFIED] | **YES for one file; NO for several in Claude Code** (newline-separated paths do not attach, replayed). opencode attaches several only if no path has a space (replayed) |
| WezTerm | last tagged release 20240203-110809-5046fc22; also main at 458a100 (2026-11-08) | Absolute path with only spaces backslash-escaped (default `SpacesOnly`); everything else raw. Main adds one trailing space; the tagged release does not [SOURCE]. Setting `quote_dropped_files` [DOCUMENTED] | **Yes**, when 2004 is on [SOURCE] | One paste, paths joined by one space [SOURCE] | Only filename drops are accepted on macOS; text, URL and image-only drags are refused [SOURCE] | **YES** (replayed). One edge: opencode fails on several files when a name contains an unescaped quote character such as `'` |
| Alacritty | v0.17.0 (winit 0.30.13) | Raw absolute path, no quoting or escaping, one trailing space [SOURCE] | **Yes**, when 2004 is on; ESC and ETX are stripped [SOURCE] | **One separate paste per file**, each `path ` [SOURCE] | Only filename drops accepted on macOS; text and image-only drags refused [SOURCE] | **YES** (replayed: attaches in both, raw spaces included, several files as separate pastes) |
| Windows Terminal | v1.25.3131.0 | Path wrapped in `"..."` only if it contains a space; no other escaping. With a path-translation style set (WSL), single-quoted translated path. No trailing space [SOURCE] | **Yes**, when 2004 is on [SOURCE] | One paste, paths joined by one space [SOURCE] | Text, a web link or an app link on the drag win over the files, so a browser image drag usually pastes a URL, not a file [SOURCE] | **UNKNOWN** for native Windows paths (`C:\...`): Claude Code on Windows was not available. The quoted form with POSIX-style paths replayed fine in both |

No terminal that I could read the source of delivers a file drop as plain keystrokes: all six send it through their
paste routine, so a program that has enabled 2004 sees the brackets. The only one that does not claim bracketing
for some drops is iTerm2's Cmd-drag, which is a different feature (`cd` plus Enter). Terminal.app is the one that is
not settled.

## How the bytes were reproduced

- Test files (64x64 solid PNGs generated with Python) in a directory whose path has no spaces:
  `plain.png` (call it P), `dir with space/red shot (1).png` (S), `special/it's [final] $x & y.png` (X, an apostrophe,
  brackets, a dollar sign, an ampersand and spaces).
- Ghostty: `macos/Sources/Ghostty/Ghostty.Shell.swift` from v1.4.0, copied verbatim and compiled with `swiftc` (with
  `-enable-bare-slash-regex`, which the project sets); its `Ghostty.Shell.escape` was run on P, S and X. [OBSERVED]
- iTerm2: `+shellEscapableCharacters`, `-escapeShellCharactersWithBackslashIncludingNewlines:`,
  `-escapeCharacters:` / `-escapeCharacters:forBash:` and `-quotedStringForPaste`, copied verbatim from
  `sources/Categories/NSStringITerm.m` at v3.7.7 into an Objective-C category and compiled with `clang`; run on P, S
  and X. The `escapeWithQuotes` advanced setting (default NO) was left at its default. [OBSERVED]
- kitty: `parse_uri_list` copied verbatim from `kitty/utils.py` at v0.50.0 and run on a CRLF-joined `file://` list
  built the way `glfw/cocoa_window.m` `create_uri_list` builds it (percent-encoded `absoluteString` per URL, joined by
  `\r\n`), then the `'\n'.join` from `Window.on_drop` and the `shlex.quote` step from `paste_with_actions`. The
  at-prompt/not-at-prompt switch itself (`self.at_prompt`) was not exercised. [OBSERVED]
- WezTerm: `DroppedFileQuoting::escape` copied verbatim from `config/src/config.rs` (main) into a Rust program using
  the `shlex` crate, with the `join(" ") + " "` from `WindowEvent::DroppedFile`. [OBSERVED]
- Alacritty: `path + " "` is one expression, written directly; no harness needed. [SOURCE]
- Windows Terminal: its quoting is C++/WinRT inside `TermControl`; I did not build it. I wrote the quoting rule (quote
  if the path contains a space) by hand in Python, with macOS-style paths, so this row is a re-implementation of the
  rule, not the terminal's code, and says nothing about backslash paths. [UNVERIFIED]
- Replay: each terminal's bytes were loaded with `tmux load-buffer` and sent with `tmux paste-buffer -p` (which adds
  `ESC[200~ ... ESC[201~` when the program has enabled 2004) into Claude Code 2.1.322 and opencode v2.2.5 running in a
  private tmux server (tmux 3.8, `-L termdrop`, 200x50), working directory the gorex repo. Nothing was submitted to a
  model. The chip (`[Image #N]` in Claude Code, `[Image N]` in opencode) was counted on the captured screen. The
  input was cleared between cases and the tool restarted when the clear left text behind. A control run of the same
  P path sent as plain keystrokes (`send-keys -l`) left plain text and no chip. [OBSERVED]
- Scratch files (not part of the repo): `h/` (the four harnesses), `run/drive.py` (builds the cases), `run/runner.py`
  (replays them), `run/cases.json`, `run/res-claude.json`, `run/res-opencode.json`, all under
  `/private/tmp/claude-501/-Users-andreangelantoni-Projects-gorex/2f9f4a80-b13a-42cd-a7f9-296bc8c73bb0/scratchpad/`.

### Replay results (chips counted; every single-file case gave 1 chip in both tools)

Single file, forms P, S and X, in every terminal's form, one bracketed paste: **1 chip in Claude Code and 1 chip in
opencode for all 24 combinations.** [OBSERVED] That includes Ghostty's `\(`, `\'`, `\[`, `\$`, `\&` escapes,
iTerm2's escape set and its `"..."` pref form (with trailing space), kitty's raw path with raw spaces and an
apostrophe, WezTerm's spaces-only escaping, Alacritty's raw path plus trailing space, and the double-quoted form.

Several files (P+S, S+X, P+P; the number is chips, 2 means both files attached):

| Terminal form | Claude Code 2.1.322 (P+S, S+X, P+P) | opencode 2.2.5 (P+S, S+X, P+P) |
| --- | --- | --- |
| Ghostty (escaped, space-joined) | 2, 2, 2 | 2, 2, 2 |
| iTerm2 (escaped, space-joined, trailing space) | 2, 2, 2 | 2, 2, 2 |
| iTerm2 with the quote pref on | 2, 2, 2 | 2, 2, 2 |
| kitty, not at a prompt (newline-joined, raw) | **0, 0, 0** | **0, 0**, 2 (P+P has no spaces) |
| WezTerm main (spaces-only, trailing space) | 2, 2, 2 | 2, **0**, 2 |
| WezTerm 20240203 (no trailing space) | 2, 2, 2 | 2, **0**, 2 |
| Alacritty (one paste per file, `path `) | 2, 2, 2 | 2, 2, 2 |
| Double-quoted-if-space (Windows Terminal rule) | 2, 2, 2 | 2, 2, 2 |

Reading it [OBSERVED]: newline-separated paths never attach in Claude Code, and in opencode only when no path has a
space (this matches the earlier research doc). The WezTerm S+X zero in opencode is the unescaped apostrophe in X
opening a quote that never closes. A trailing space never hurt in any case.

## Terminal.app (macOS, closed source)

- Version here: Terminal 2.15, build 488.9, macOS 27.2.1. [OBSERVED] (`defaults read` on the app's Info.plist)
- Apple's User Guide says: "Enter the path of a file or folder: Drag the file or folder to the Terminal window. The
  item's absolute path appears on the command line." [DOCUMENTED]
  https://support.apple.com/guide/terminal/drag-items-into-a-terminal-window-trml106/mac (section "Drag items into a
  Terminal window"). The page says dragging a text clipping "may add the escape equivalent of certain special text
  characters" and that Command-drag avoids that. [DOCUMENTED] It says nothing about quoting a path, about spaces,
  about several files or about bracketed paste. [DOCUMENTED]
- The Terminal Keyboard shortcuts page lists "Paste escaped text: Control-Command-V". [DOCUMENTED]
  https://support.apple.com/guide/terminal/keyboard-shortcuts-trmlshtcts/mac. That is a menu paste, not the drop.
- Whether the drop path is escaped, quoted or raw: Apple's page does not say. [UNVERIFIED]
- Static inspection of the app binary (`strings`, `otool -oV` on `/System/Applications/Utilities/Terminal.app`):
  it contains the strings "You cannot send the text to the terminal, because it contains invalid content for
  Bracketed Paste Mode", the property `bracketedPasteMode` and `setBracketedPasteMode:`, and a text-view class
  with `performDragOperation:`, `draggingEntered:`, `readSelectionFromPasteboard:isDragAndDrop:isCommand:sender:`,
  `pasteEscaped:` and `folderItemsFromPasteboard:`. [OBSERVED] That shows Terminal.app implements mode 2004 and
  that its drop and paste handlers share a `readSelectionFromPasteboard:` family of methods. It does not show that a
  drop is sent bracketed, quoted or escaped; I did not disassemble the method. [UNVERIFIED]
- Several files, image files, non-file drags, any setting that changes the drop: nothing documented. [UNVERIFIED]
- Verdict: UNKNOWN. What would settle it: one real drag with the capture in "Open questions".

## iTerm2

- Version: tag v3.7.7 (ref `e0025340b49a2fc6415f490d77bbead5d3e0583e`), the same version as the installed
  `/Applications/iTerm.app` (3.7.7). [OBSERVED] for the installed version; the code below is read at that tag.
- Drop target: `PTYTextView` registers `NSPasteboardTypeFileURL`, `NSPasteboardTypeString`, `NSPasteboardTypePNG` and
  `NSPasteboardTypeTIFF`; the image types are there "so a program that accepts Kitty drag-and-drop image drops
  receives pure-image drags (with no file URL or string)". [SOURCE] `sources/TerminalView/PTYTextView.m`, init, and
  `-performDragOperation:`.
- A drop that no program claimed through the Kitty drag-and-drop protocol goes to the legacy path in
  `-performDragOperation:`. [SOURCE] A program only claims drops after it opts in to that protocol; Claude Code and
  opencode do not appear to (opencode's OpenTUI repo had no hits for `dnd` or `5522` in code search; Claude Code is
  closed source). [UNVERIFIED]
- Legacy path, files, local pane: `-pasteValuesOnPasteboard:cdToDirectory:` calls
  `-[NSPasteboard filenamesOnPasteboardWithShellEscaping:YES forPaste:YES]`, appends the escaped names joined by a
  single space, then appends one more space, and calls `[_delegate pasteString:]`. [SOURCE]
  `PTYTextView.m` `-pasteValuesOnPasteboard:cdToDirectory:`; `sources/Pasting/NSPasteboard+iTerm.m`.
- Escaping: by default backslash escaping via `stringWithEscapedShellCharactersIncludingNewlines:YES`, over
  `+shellEscapableCharacters` = `\ ()"&'!$<>;|*?[]#` backtick, tab, `{}^+=@~`, plus CR and LF. [SOURCE]
  `sources/Categories/NSStringITerm.m`. The result for the three test names, from the verbatim code, was
  `.../dir\ with\ space/red\ shot\ \(1\).png` and `.../special/it\'s\ \[final\]\ \$x\ \&\ y.png`. [OBSERVED]
- Settings that change it: the preference `kPreferenceKeyWrapDroppedFilenamesInQuotesWhenPasting` (default `@NO`)
  switches to `quotedStringForPaste`, which wraps the path in `"..."` and escapes `"` and `\` [SOURCE]
  (`sources/Settings/iTermPreferences.m`, `NSPasteboard+iTerm.m`); the experimental advanced setting `escapeWithQuotes`
  (default NO) switches to POSIX `'...'` quoting [SOURCE] (`iTermAdvancedSettingsModel.m`); the advanced setting
  `fileDropCoprocess`, when set, runs a silent coprocess instead of pasting the names [SOURCE]
  (`PTYSession.m` `-textViewPasteFiles:`; description in `iTermAdvancedSettingsModel.m`).
- Bracketing: `-[PTYSession pasteString:]` calls `pasteString:flags:0`, which passes `allowBracketing` to the paste
  helper; the helper brackets when `pasteHelperShouldBracket` is true, and that returns
  `_screen.terminalBracketedPasteMode`. [SOURCE] `PTYSession.m` `-pasteString:flags:`, `-pasteHelperShouldBracket`;
  `sources/Pasting/iTermPasteHelper.m`, `PasteEvent.m` `-addPasteBracketing`. So a drop is a bracketed paste when the
  foreground program has turned 2004 on, and plain text otherwise.
- Several files: all names escaped, joined by single spaces, in one paste. [SOURCE] (same function.)
- Image files: no image-specific branch for a file drop; it is the same path. [SOURCE] A browser image drag that
  carries only image data (no file URL, no string) falls through both branches and the drop is refused;
  the source comment says exactly this ("an image from a browser ... a drop that cannot land"). [SOURCE]
  `PTYTextView.m` `-legacyDraggingOperationForSender:` and `-dragOperationForSender:numberOfValidItems:`.
- Modifier and context differences, all [SOURCE] in `-performDragOperation:` and `-pasteValuesOnPasteboard:`:
  Option held on drop uploads the files (scp) instead of pasting; Command-drag (source mask exactly Generic) pastes
  `cd <dir>` plus CR through `pasteStringWithoutBracketing:`, so that one is not bracketed and presses Enter;
  and when shell integration reports a remote host (`textViewIsOnLocalhost` false) a dialog offers actions such as
  "Paste Path", "Upload and Paste Path(s)" and "Paste Base64-Encoded Contents" (`iTermNonTextPasteHelper.swift`,
  `FilePasteAction`). Without shell integration the host is unknown and the drop is pasted as a local path even in an
  ssh session, which a remote program cannot read.
- Replay: the iTerm2 bytes (escaped, trailing space; and the quote pref form) attached in both tools for single and
  several files. [OBSERVED]
- Verdict: YES for a local pane (and for an ssh pane without shell integration, but then the path is the Mac's).

## Ghostty

- Version: tag v1.4.0 (commit `83fb95624b0439ee660c2679b43b38c8b63769cd`). Also read v1.3.1 and v1.2.3. Ghostty is not
  installed on this Mac. [SOURCE]
- v1.4.0 drop handler: `Ghostty.SurfaceView.performDragOperation` in
  `macos/Sources/Ghostty/Surface View/SurfaceView_AppKit.swift` gets `pb.getOpinionatedStringContents()` and calls
  `surfaceModel?.pasteText(content)`, with the comment "Dropped text is untrusted, so it goes through the same paste
  protection as a clipboard paste". [SOURCE]
- Content: `NSPasteboard.getOpinionatedStringContents()` in `macos/Sources/Helpers/Extensions/NSPasteboard+Extension.swift`
  maps each pasteboard item: if it has a `.fileURL`, the item becomes `Ghostty.Shell.escape(fileURL.path)`, otherwise
  its `.string`; the pieces are joined with a single space; no trailing space. [SOURCE]
- Escaping: `Ghostty.Shell.escape` backslash-prefixes each of `\ ()[]{}<>"'` backtick `!#$&;|*?` and tab
  (`macos/Sources/Ghostty/Ghostty.Shell.swift`, `escapeCharacters`); the project's `ShellTests.swift` pins it, for
  example `/tmp/my file (1).txt` becomes `/tmp/my\ file\ \(1\).txt`. [SOURCE] Run from the verbatim file, the three
  test names came out as `.../dir\ with\ space/red\ shot\ \(1\).png` and `.../special/it\'s\ \[final\]\ \$x\ \&\ y.png`.
  [OBSERVED]
- Delivery: `Ghostty.Surface.pasteText` calls `ghostty_surface_paste_text`; in `src/apprt/embedded.zig` `pasteText`
  builds a paste request and ends in `Surface.completeClipboardPaste`, which calls `input.paste.encode`; with
  `bracketed` taken from the terminal's mode 2004 it prefixes `ESC[200~` and suffixes `ESC[201~`, otherwise it turns
  `\n` into `\r`. Bytes such as ESC, DEL, Ctrl-C, Ctrl-U are replaced by spaces in both cases. [SOURCE]
  `src/Surface.zig` `completeClipboardPaste`; `src/input/paste.zig` `encode`.
- Paste protection: `clipboard-paste-protection` and `clipboard-paste-bracketed-safe` both default to true, so a
  bracketed drop of a normal path is not held for confirmation; an unbracketed drop containing CR or LF would be.
  [SOURCE] `src/config/Config.zig`; `src/input/paste.zig` `isSafe`.
- Older tags: v1.3.1 and v1.2.3 `performDragOperation` first checks `pb.string(forType: .URL)` (a URL, escaped
  as-is), then file URLs (each `Shell.escape(path)`, joined by a space), then `.string`, and calls
  `self.insertText(content, replacementRange: ...)`. `insertText` calls `surfaceModel.sendText`, which calls
  `ghostty_surface_text`, whose core `textCallback` runs `completeClipboardPaste(text, true)` with the doc comment "if
  bracketed mode is on this will do a bracketed paste". [SOURCE] (v1.2.3 `SurfaceView_AppKit.swift`,
  `Ghostty.Surface.swift` `sendText`; v1.2.3 `src/Surface.zig` `textCallback`.) So for a file drop all three tags are
  bracketed when 2004 is on and give the same escaped path; v1.4.0 dropped the `.URL` branch, so a link drag now goes
  through `.string`.
- Non-file and image drops: `dropTypes` is `[.string, .fileURL]`; `draggingEntered` returns no operation when neither
  is present, so an image-only drag is refused. [SOURCE] A browser image drag that also carries a text URL would paste
  that URL text; what Safari or Chrome actually put on the pasteboard was not inspected. [UNVERIFIED]
- Settings that change it: no drop-specific option; the paste-protection options above. [SOURCE]
- Replay: attaches in both tools for single and several files. [OBSERVED]
- Verdict: YES.

## kitty

- Version: tag v0.50.0 (ref `5e0197fef01c1141925d2f7be08851b87af0b454`). The new drag-and-drop protocol and kitten
  arrived in 0.47 [DOCUMENTED] `docs/changelog.rst` ("Drag and drop for terminal programs [0.47]") and
  `docs/dnd-protocol.rst` ("versionadded:: 0.47.0"). Only v0.50.0 was read; pre-0.47 behaviour is not covered.
  [UNVERIFIED]
- When a program in the window has enabled the drag-and-drop protocol (`w->drop.wanted`), the drop goes to that
  program through the protocol and no text is pasted. Otherwise kitty handles it itself. [SOURCE]
  `kitty/glfw.c` `drop_dest_callback` (`is_client_drop`). Claude Code and opencode have no reason to enable it.
  [UNVERIFIED]
- What macOS offers: `glfw/cocoa_window.m` `draggingEntered:` advertises `text/uri-list` if the pasteboard can read
  `NSURL` objects (which includes web URLs and file promises) and `text/plain` if it can read a string;
  `create_uri_list` builds the list from `url.absoluteString` joined by `\r\n`. [SOURCE] The view registers many
  pasteboard types including `NSPasteboardTypeFileURL`, string, URL, PNG, TIFF, and the file-promise types.
  [SOURCE]
- What is read: `kitty/glfw.c` `is_droppable_mime` ranks `text/uri-list` (3) over `text/plain;charset=utf-8` (2) over
  `text/plain` (1); only those are read for a plain drop. [SOURCE]
- Conversion: `Window.on_drop` in `kitty/window.py`: if `text/uri-list` is present, `parse_uri_list` turns each
  `file://` line into the percent-decoded path and passes other lines through, and the lines are joined with
  `'\n'`; otherwise `text/plain` is used. [SOURCE] `parse_uri_list` is in `kitty/utils.py`. Run from the verbatim
  function, the three test files came out as plain absolute paths with raw spaces, parens and apostrophe, separated
  by `\n`. [OBSERVED]
- Quoting: `Window.paste_with_actions` applies the `paste_actions` option, whose default is
  `quote-urls-at-prompt,confirm`; with `quote-urls-at-prompt`, **only if `self.at_prompt` is true** and the text came
  from a URI list, every line goes through `shlex.quote` and the lines are joined with a single space. When a program
  other than the shell is running (the case for Claude Code and opencode), `at_prompt` is false and the paths stay
  raw and newline-separated. [SOURCE] (`kitty/window.py` `paste_with_actions`, `at_prompt` at line 960;
  `kitty/options/definition.py` `paste_actions`; the option text says quoting "needs shell_integration".) Whether
  `at_prompt` is really false while those two programs run was not tested with a live kitty shell integration.
  [UNVERIFIED]
- Delivery: `paste_with_actions` ends in `Window.paste_text`, which strips any `ESC[201~` when 2004 is on, then
  `screen.paste`, and `Screen.paste_` writes `ESC[200~`, the data and `ESC[201~` when the `BRACKETED_PASTE` mode is
  set; otherwise it writes the bytes with `\n` turned into `\r`. [SOURCE] `kitty/window.py` `paste_text`;
  `kitty/screen.c` `paste_`. No trailing space is added. [SOURCE]
- The `confirm` default only asks when the text contains control codes, or contains newlines while 2004 is off; a
  newline-separated path list in a 2004-on program is not held. [SOURCE] (`paste_with_actions`, the `'confirm'`
  branch). The option text in `kitty/options/definition.py` says the same: "Drops into kitty follow the receiving
  window's paste_actions settings. Multiline text or text containing control codes can therefore require
  confirmation." [SOURCE]
- Image and non-file drags: a file promise (what Safari-style and many browser image drags provide) is fulfilled
  into a temp directory, its path is added to the URI list, and kitty keeps that directory for 600 seconds
  (`DROPPED_FILE_PROMISES_LIFETIME`). [SOURCE] `glfw/cocoa_window.m` (`file_promise_is_for_a_file`, the promise
  receiver code) and `kitty/boss.py` `dropped_file_promises_dir`. So an image dragged out of a browser can reach the
  program as a temp-file path. A dropped web URL arrives as a `http(s)://` line (`parse_uri_list` passes non-`file://`
  lines through). [SOURCE] Which browsers actually do this: not run. [UNVERIFIED]
- Settings that change it: `paste_actions` (`quote-urls-at-prompt`, `replace-dangerous-control-codes`,
  `replace-newline`, `confirm`, `confirm-if-large`, `filter`, `no-op`). [SOURCE] `kitty/options/definition.py` at
  v0.50.0 (the option's own doc text; I did not fetch the website copy of it).
- Replay: a single file (P, S or X) as a raw path attached in both tools; several files newline-separated did not
  attach in Claude Code at all and attached in opencode only when no path had a space. [OBSERVED]
- Verdict: YES for one file; NO for several files in Claude Code, and only for space-free names in opencode.
  Workaround for several files: the default `quote-urls-at-prompt` does not help because it needs a shell prompt;
  a `filter` paste action (a `filter_paste()` function in `paste-actions.py`) could in principle turn the newline
  list into an escaped, space-joined one, but this was not tried. [UNVERIFIED]

## WezTerm

- Versions: the last tagged release is 20240203-110809-5046fc22 (published 2024-02-03); main was read at commit
  `458a100c1bf4134d36cb5c42dd90e4c8266d9e1a` (2026-11-08). They differ in one respect, below. [SOURCE]
- Drop target (macOS): `window/src/os/macos/window.rs` registers only `NSFilenamesPboardType`; `performDragOperation`
  reads that list and dispatches one `WindowEvent::DroppedFile(Vec<PathBuf>)` with all the paths; with no filenames
  it returns NO. So text, URL and image-only drags are refused. [SOURCE] (The changelog also records "macOS:
  dragging non-filename data over wezterm could cause it to crash. #4771".) [DOCUMENTED] `docs/changelog.md`.
- Handling: `wezterm-gui/src/termwindow/mod.rs` `WindowEvent::DroppedFile` maps each path through
  `config.quote_dropped_files.escape(...)`, joins with `" "`, and calls `pane.send_paste`. On main there is a
  `+ " "` after the join (one trailing space); at 20240203-110809-5046fc22 there is not. [SOURCE]
- Quoting: `DroppedFileQuoting::escape` in `config/src/config.rs`: `None` raw; `SpacesOnly` (default off Windows)
  `s.replace(" ", "\\ ")`; `Posix` via `shlex::try_quote`; `Windows` double-quote if the name has space, tab, LF, VT or
  `"`; `WindowsAlwaysQuoted`. The default is `Windows` on Windows and `SpacesOnly` elsewhere. [SOURCE] The doc
  `docs/config/lua/config/quote_dropped_files.md` shows the same table. [DOCUMENTED] Run from the verbatim function:
  `.../special/it's\ [final]\ $x\ &\ y.png` (only the spaces escaped, the apostrophe, brackets, `$` and `&` raw).
  [OBSERVED]
- Delivery: `Pane::send_paste` reaches `TerminalState::send_paste` in `term/src/terminalstate/mod.rs`, which writes
  `ESC[200~`, the text with embedded `ESC[200~` and `ESC[201~` removed, and `ESC[201~` when `bracketed_paste` is on;
  otherwise it canonicalises newlines per config and writes the bytes. [SOURCE]
- Platforms: macOS, Windows, Wayland since 20220624-141144-bd1b7c5d; X11 only on nightly. [DOCUMENTED]
  `docs/config/lua/config/quote_dropped_files.md`.
- Replay: attached in both for single files; several files attached in both except opencode with name X (unescaped
  apostrophe), 0 chips. [OBSERVED]
- Verdict: YES for ordinary names. Setting `quote_dropped_files = "Posix"` quotes the name with `'...'`, which
  both tools accept (earlier doc), though its output for a name with an apostrophe was not replayed.

## Alacritty

- Version: tag v0.17.0 (ref `dc946285396bd3e251b6942af7e9961b74f76bee`); `Cargo.lock` pins winit 0.30.13. [SOURCE]
- Drop target (macOS, via winit): `src/platform_impl/macos/window_delegate.rs` at winit v0.30.13 (ref
  `45a6f047c7eb0d68f37138cbdd38eec9cd4c1575`) registers only `NSFilenamesPboardType` and, in `performDragOperation`,
  queues **one `WindowEvent::DroppedFile(path)` per filename**. Text and image-only drags are refused. [SOURCE]
- Handling: `alacritty/src/event.rs`: `WindowEvent::DroppedFile(path)` does
  `let path: String = path.to_string_lossy().into(); self.ctx.paste(&(path + " "), true);`. No escaping or quoting,
  one trailing space (changelog: "Dragging files into terminal now adds a space after each path"). [SOURCE]
  [DOCUMENTED] `CHANGELOG.md` at v0.17.0.
- Delivery: `ActionContext::paste(text, bracketed)`: when `bracketed` and the terminal mode has `BRACKETED_PASTE`
  it writes `ESC[200~`, the text with ESC and ETX removed, `ESC[201~`; otherwise it writes the text with CRLF and LF
  turned into CR. [SOURCE]
- Consequences: a name with spaces arrives as raw spaces; several files arrive as separate consecutive pastes, each
  one path plus a space, so a program sees N pastes, not one. [SOURCE]
- Settings: none for drops found. [SOURCE] (I searched `event.rs` and the changelog; no config key.)
- Replay: `path ` as its own paste, once per file, attached in both tools for every case including S and X. [OBSERVED]
- Verdict: YES.

## Windows Terminal

- Version: tag v1.25.3131.0 (ref `82b360d0e011a3d1a9ec4b4ef079518b5b2c1b26`), the latest release on 2026-11-11.
  [SOURCE]
- `src/cascadia/TerminalControl/TermControl.cpp` `TermControl::_DragDropHandler` checks the drag data in this order:
  application link, web link, text, and only then storage items (files). So a drag that carries text or a link is
  pasted as that text (the link's `AbsoluteUri`), and a file drop from Explorer (storage items only) is pasted as
  paths; the code comment says storage items must be last because some apps put hybrid formats in a drop. [SOURCE]
- Files: the full paths are collected (a workaround reads the `FileDrop` format when exactly 16 items are returned,
  for archives), each optionally translated by the profile's `PathTranslationStyle` (WSL and similar), quoted, and
  joined by a single space. Quoting: with no translation, wrap in `"` only if the path contains a space; with a
  translation style, always wrap in `'`. No other character is escaped; no trailing space. [SOURCE]
- Delivery: `_pasteTextWithBroadcast` then `ControlCore::PasteText`, which filters the string (CR/LF and control codes)
  and, when `BracketedPasteEnabled()`, wraps it in `ESC[200~ ... ESC[201~`. [SOURCE]
  `src/cascadia/TerminalControl/ControlCore.cpp`.
- Image and non-file: no image-specific branch. An image dragged from a browser normally carries a web link or text,
  so the URL would be pasted, not a file; this follows from the order above, but what a given browser puts in the
  drag data was not inspected. [UNVERIFIED]
- Settings: the profile's path translation style. [SOURCE]
- The earlier research doc (section 2) notes Windows Terminal before 1.25 sends an empty bracketed paste for an
  image-only clipboard; that is a paste, not a drop, and is not changed by anything here.
- Replay: only a POSIX-path stand-in for the quoting rule, as flagged above; double-quoted paths with spaces attached
  in both tools. Native `C:\...` paths in Claude Code and opencode on Windows were not run. [UNVERIFIED]
- Verdict: UNKNOWN for native Windows paths; likely YES in WSL with the translated single-quoted path, since
  single-quoted paths attach in both tools. [UNVERIFIED]

## What this means for GoRex and issue #8

- GoRex generates its own paste, so none of this changes its design. The recommended GoRex form (absolute,
  backslash-escaped path, one bracketed paste, optional trailing space; earlier doc) is the same shape that Ghostty
  and iTerm2 produce natively, and WezTerm, Alacritty and iTerm2 also add a trailing space, which the replay showed
  harmless in every case. That is extra confirmation for the trailing-space choice.
- A drop onto a local pane in iTerm2, Ghostty, WezTerm or Alacritty should therefore already attach in Claude Code
  and opencode, when the path is local and has no exotic characters (replayed, not dragged). kitty works for one
  file only. Terminal.app and native Windows Terminal paths are not established.
- None of these terminals copies the file anywhere. In an ssh pane they paste the Mac's local path, which the remote
  Claude Code or opencode cannot read (earlier doc, section 5), so GoRex's copy-then-paste-remote-path flow remains the
  thing that makes remote attachment work. The one terminal with a built-in equivalent is iTerm2 with shell
  integration: its remote-host dialog offers "Upload and Paste Path(s)" [SOURCE], worth a mention in the README as
  the closest native feature.
- Several images in one GoRex paste: use space-separated escaped paths in one paste (matches Ghostty and iTerm2). Do not
  use kitty's newline separation (zero attachments in Claude Code, observed). Alacritty's one-paste-per-file also
  works in both tools, so a GoRex fallback of one paste per file is valid if a combined paste is ever a problem.
- Escape every shell-special character, not only spaces. WezTerm's spaces-only form and Alacritty's raw form fail in
  opencode when a name has an apostrophe and several files share a paste (observed: 0 chips). GoRex already plans
  "sanitized names"; escaping covers the rest.
- Non-file drags (an image dragged out of a browser): the terminals differ (Ghostty pastes a URL string, Windows
  Terminal a URL, kitty may hand over a temp-file path, WezTerm and Alacritty refuse, iTerm2 refuses image-only data).
  GoRex should not assume a file path arrives; handle the cases it can read (file URL, or image data it saves to
  `~/.gorex/drops`) and document the rest as unsupported. These browser cases were not run. [UNVERIFIED]
- What to say in the README (suggested wording, limited to what the evidence above supports): "Dropping a file on
  a pane in GoRex copies it to the host and pastes its escaped path as a bracketed paste. Other terminals insert the
  local path of a dropped file and, for iTerm2, Ghostty, WezTerm and Alacritty, that also attaches in Claude Code and
  opencode when the program runs on the same machine; it does not work over ssh, where the path points at your Mac."
  The "not verified with a real drag" caveat belongs with it until the open questions are run.

## Open questions

Each needs a real drag on a real window and cannot be settled headlessly. Capture command, the same in every
terminal, from a shell prompt: `printf '\033[?2004h'; cat -v` then drop the file, press Return, then press Ctrl-D
(twice if needed). Turning mode 2004 on first matters: `cat -v` alone does not enable it, so the terminal would send
the unbracketed form. A bracketed drop shows `^[[200~...^[[201~`; an unbracketed one shows the bare text. Also run
`printf '\033[?2004l'` afterwards. For exact bytes use `printf '\033[?2004h'; xxd` instead and compare.
Test files: one PNG named `plain.png`, one named `red shot (1).png`, one named `it's [final] $x & y.png`; drop each
alone, then all three together.

1. Terminal.app 2.15: run the capture above for each of the three single files and the group. Settles: path form,
   quoting, separator, trailing space, bracketed or not. (Everything about Terminal.app is open until this is done.)
2. iTerm2 3.7.7, local shell: same capture; then with Option held, then with Command held (expect `cd ...` + CR
   unbracketed); then in an ssh session with shell integration installed (expect the dialog). Settles the
   `[SOURCE]` claims above on a real build.
3. Ghostty (record the version): same capture. Also settles whether v1.4.0 and v1.3.x really match for a file drop.
4. kitty 0.50.0: same capture, once in a shell prompt with kitty shell integration (expect `shlex.quote`d,
   space-joined) and once while `cat -v` runs (expect raw, newline-joined). This settles whether `at_prompt` is false
   while Claude Code or opencode runs.
5. WezTerm: tagged 20240203 and a recent nightly: same capture; expect the trailing space only on the nightly.
6. Alacritty 0.17.0: same capture with three files; expect three separate pastes.
7. Windows Terminal 1.25: in WSL run the capture; from Explorer drop a file whose path has a space; and drop an
   image from a browser. Also paste a native Windows path into Claude Code on Windows and into opencode on Windows to
   see whether `"C:\Users\me\a b.png"` attaches.
8. Browser image drag into each terminal (Safari, Chrome): run the capture, and note whether a path, a URL or nothing
   arrives. This is the only way to learn what the non-file branches really do.
9. kitty, whether a `filter` paste action could turn newline-separated drops into a space-joined, escaped list for
   Claude Code. Settle by a `paste-actions.py` and one multi-file drop.
10. Whether Claude Code or opencode ever enables the kitty drag-and-drop protocol (which would divert drops in kitty
    and iTerm2 from the paste path). Settle by recording the program's terminal output (for example
    `script -q out.txt claude`) and searching it for the accept request, `OSC <dnd code> ; t=a ... ST` per
    `docs/dnd-protocol.rst` (the numeric code is `DND_CODE` in kitty's C sources, which I did not resolve; search the
    capture for `;t=a` after an `ESC ]`).
