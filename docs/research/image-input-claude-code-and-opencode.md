# Image input in Claude Code and opencode: dropped paths, clipboard, `@`, ssh

Date: 2026-11-11 (all runs the morning of that day, Pacific Time). Written for GitHub issue #8
(Performant-Labs/gorex, "Drop or paste images onto a pane and deliver them to Claude Code and opencode,
locally or over ssh").

Every claim carries one label:

- **[D]** DOCUMENTED: official docs (URL and section given).
- **[S]** SOURCE: read in the code (repo, tag or commit, file and line or symbol).
- **[O]** OBSERVED: I ran it (version, exact input, what the screen showed).
- **[U]** UNVERIFIED: not settled; what would settle it is stated.

Versions tested:

| Tool | Where | Version |
| --- | --- | --- |
| Claude Code | local Mac and `jupiter` (Linux, headless, no `DISPLAY`) | 2.1.322 |
| opencode | local Mac (`/opt/homebrew/bin/opencode`) | v2.2.5 (tag commit `9bbb8ce18c80f099e0e448775f59d965ead40a21`) |
| opencode | `jupiter` (`/usr/local/bin/opencode`) | v2.0.29 (tag commit `042d0832f5f3a73eaa039a58a7d565776f6d4290`) |
| OpenTUI (opencode's TUI toolkit) | npm `@opentui/core` | 0.8.0 for opencode 2.2.5, 0.5.17 for 2.0.29 (pinned in each tag's root `package.json`); native clipboard source read at `sst/opentui` tag v0.8.0 (`738067f47e940dcec3f5aab74868908574335c6c`) |

Method for the [O] rows: each TUI ran inside a private tmux server (tmux 3.8) in a pseudo-terminal; for ssh
cases the pane ran `ssh -t jupiter 'cd ~/.gorex-research && exec <tool>'`, the same shape GoRex uses. "Bracketed
paste" means `tmux paste-buffer -p`, which wraps the text in `ESC[200~ ... ESC[201~` when the foreground app has
turned on mode 2004. "Typed" means `tmux send-keys -l`, i.e. plain keystrokes with no bracket. The screen was read
with `capture-pane`. The attachment indicator is `[Image #N]` in Claude Code and `[Image N]` in opencode. The test
images were 64x64 solid PNGs generated with Python (blue `plain.png`, red `red shot.png` in a directory with a
space in its name). Real Finder drops into a real terminal emulator were not run (headless), so how a given terminal
formats a drop is [U]; GoRex generates the paste itself, so that matters less.

## Summary

| | Claude Code 2.1.322 | opencode 2.2.5 (2.0.29 identical in every row tested) |
| --- | --- | --- |
| Pasted path to an image becomes an attachment | Yes, **only as a bracketed paste** [O]; typed keystrokes stay text [O] | Yes, **only as a bracketed paste** [O][S]; typed keystrokes stay text [O] |
| Path forms that attach | Absolute path; `\ `-escaped spaces; `'...'` or `"..."` quoted; raw unescaped spaces (whole paste is one path); trailing space or newline; several `\ `-escaped paths separated by spaces [O]. **Not** `file://`, relative, `~`, newline-separated lists, or a path inside other text [O] | Absolute; relative to the TUI's cwd; `\ `-escaped; quoted; raw spaces; `file://` (percent-decoded); trailing space or newline; several paths separated by whitespace (up to 32, each path escaped or quoted) [O][S]. **Not** `~`, newline-separated paths that contain unescaped spaces, or a path inside other text [O][S] |
| Clipboard image paste | Ctrl+V (Cmd+V in iTerm2, Alt+V on Windows and WSL) [D]. Over ssh to Linux it refuses and says "copy the file over, paste its path" [O][D] | Ctrl+V only [S]. Reads the clipboard of the machine the process runs on, with no terminal-side image read [S]; on headless `jupiter` Ctrl+V did nothing visible [O] |
| `@image.png` | Works: `@plain.png` showed "Read plain.png (168 bytes)" and the model answered "Blue" [O]. Pasting `@/abs/x.png` text is not special [O] | Works through the autocomplete menu (type `@plain`, Tab): a `file plain.png` attachment appeared [O]; a pasted `@/abs/x.png` stayed plain text with "No matching files" [O] |
| Formats | JPEG, PNG, GIF, WebP at the API level [D] | PNG, JPEG, GIF, WebP to the model, PDF; SVG as text; AVIF and BMP are not sent [D]; the paste route recognises avif, gif, jpeg, jpg, pdf, png, svg, webp by extension [S] |
| Limits | API: 8000x8000 px, 10 MB base64 on the first-party API (5 MB on Bedrock), long edge resized to 1568 px (2576 on Claude 4.7+) [D] | 20 MiB per attachment; resized above 2000x2000 px or 5 MiB base64 by default [D][S]; `media.image` config [D] |
| File on a remote machine | The tool reads files on the machine it runs on [O]; send the file there first | Same [O] |
| OSC 52 or other remote image clipboard | None found [U] | OSC 52 is used for **writing** text only; no image read [S] |
| Headless or flag input | `claude -p` with stream-json input and `image` blocks [D]; no `--image` flag found [D] | `opencode run -f/--file <path>` [D][O] |

Practical conclusion: for both tools the delivery that works over ssh is the same. Put the file on the machine
that runs the tool, then paste its **absolute path** as a **bracketed paste**, backslash-escaped if it has spaces,
alone in the paste, with an optional trailing space. Do not rely on `file://`, `~`, relative paths or `@` for both
tools. A clipboard-image route cannot work over ssh for either tool.

## 1. A dragged file pasted as a path

### Claude Code

- Claude Code's docs list "Drag and drop an image into the Claude Code window" and "Provide an image path to
  Claude" as ways to add an image. [D] code.claude.com/docs/en/common-workflows, section "Work with images".
  The docs do not say which path forms are recognised or that bracketed paste matters.
- The chip shown is `[Image #N]`, and Cmd+Click opens it. [D] same section. The counter keeps counting up within
  a session even after the prompt is cleared (it read #11 after earlier pastes were cleared). [O]
- Changelog: 2.1.311 "Fixed pasting text that mentions an image file, such as `! open /tmp/shot.png` or
  `/review look at /tmp/shot.png`, attaching the image and dropping its path from the prompt". [D]
  code.claude.com/docs/en/changelog, 2.1.311. This is consistent with the rule below that a path must be the
  whole paste.
- Path forms, all as a bracketed paste into the local TUI (cwd `/Users/andreangelantoni/Projects/gorex`) [O]:

  | Pasted text | Result |
  | --- | --- |
  | `/private/tmp/.../plain.png` | `[Image #N]` |
  | `/private/tmp/.../dir\ with\ space/red\ shot.png` | attached |
  | `'/.../dir with space/red shot.png'` | attached |
  | `"/.../dir with space/red shot.png"` | attached |
  | `/.../dir with space/red shot.png` (raw spaces) | attached |
  | `/.../plain.png ` (trailing space) | attached |
  | `/.../plain.png` + newline | attached |
  | two `\ `-escaped paths separated by a space | `[Image #18] [Image #19]` |
  | `file:///.../plain.png`, also with `%20` | stayed as text |
  | `docs/screenshot.png`, `./docs/screenshot.png` (exist in cwd) | stayed as text |
  | `~/Projects/gorex/docs/screenshot.png` | stayed as text |
  | two absolute paths separated by a newline | stayed as text |
  | `look at /.../plain.png` | stayed as text |
  | `/.../nope.png` (missing) | stayed as text |

- Same absolute, escaped and raw-space forms attached on `jupiter` over ssh (`ssh -t jupiter`, cwd
  `~/.gorex-research`), and `file://`, relative `plain.png`, `~/...` and a Mac path stayed text. [O]
- Typed rather than pasted: the same absolute path sent as plain keystrokes stayed text, locally and on
  `jupiter`. [O] So the path-to-attachment step is tied to the paste event. Whether a real person typing at
  normal speed behaves the same is [U] (the keystrokes arrived in one burst).
- Previously observed by the GoRex session on 2026-11-11: GoRex copied a PNG to `~/.gorex/drops` on `jupiter`
  and pasted the backslash-escaped absolute remote path in a bracketed paste; Claude Code 2.1.322 showed
  `[Image #1]` and described the image. [O] (reported, not re-run here beyond the matrix above).
- Claude Code source is not public, so the recognition rule behind this table is inferred from behaviour, not
  read. [U] What would settle it: Anthropic documenting it.

### opencode

- The paste handler is the textarea's `onPaste` in `packages/tui/src/component/prompt/index.tsx` (line 1798 at
  v2.2.5). It decodes the paste bytes, and for non-empty pastes calls `pasteInputText` (line 1447) on the pasted
  text. [S] `pasteInputText` strips ANSI sequences, trims the text, and calls `resolvePastedAttachments`; if that
  returns attachments each one becomes an attachment chip, otherwise the text is inserted (or collapsed to
  `[Pasted ~N lines]` at 3 or more lines or over 150 characters). [S]
- Because recognition lives only in the paste handler, text arriving as ordinary keystrokes is never examined. [S]
  Observed: typed absolute and escaped paths stayed text, locally on 2.2.5 and on `jupiter` on 2.0.29. [O]
- `resolvePastedAttachments` (`packages/tui/src/component/prompt/local-attachment.ts:30`) first treats the whole
  trimmed paste as ONE path after `normalizePastedFilepath` (line 93: strips leading and trailing quote
  characters, decodes `file://` URLs, un-escapes `\x` to `x` on non-Windows). If that file is not an image or
  PDF, it splits the paste with `parsePastedFilepaths` (line 119: whitespace separated, honours `'`, `"` and
  backslash escapes, `#` comment lines dropped when `file://` is present, at most 32 paths) and every piece must be
  an image or PDF, otherwise nothing attaches. [S]
- Consequences, each confirmed in the matrix below [O]: raw spaces work for a single path (the whole paste is the
  path); `file://` with `%20` works; `file://` with a host other than empty or `localhost` is refused [S]; `~` is
  never expanded (`Bun.file` gets the literal text, line 85) [S]; a path with other text around it fails the
  "every piece is an image" rule [S]; relative paths resolve against the TUI process's cwd (`Bun.file` default)
  [S] and worked. The type is decided from the file **extension** (`mimeTypes`, line 73: avif, gif, jpeg, jpg,
  pdf, png, svg, webp); the byte limit is 20 MiB (`MAX_LOCAL_ATTACHMENT_BYTES`, line 6). [S]
- Matrix [O], bracketed paste into the TUI, local v2.2.5 (cwd = the directory holding the test files):

  | Pasted text | Result |
  | --- | --- |
  | absolute path | `[Image 1]` |
  | `\ `-escaped spaces | attached |
  | `'...'` and `"..."` | attached |
  | raw unescaped spaces | attached |
  | `file:///.../dir%20with%20space/red%20shot.png` | attached |
  | trailing space, trailing newline | attached |
  | `plain.png`, `./plain.png` (cwd-relative) | attached |
  | two `\ `-escaped paths, space separated | `[Image 1] [Image 2]` |
  | two simple absolute paths, newline separated | attached (two chips, same file) |
  | newline-separated where one path has raw spaces | stayed text |
  | `~/...` | stayed text |
  | `look at /.../plain.png` | stayed text |
  | `/.../nope.png` (missing) | stayed text |
  | `@/.../plain.png` (pasted) | stayed text |

- `jupiter`, v2.0.29, over ssh, real remote files: absolute, escaped, single- and double-quoted, raw spaces,
  `file://...%20...`, trailing space, relative `plain.png`, two escaped paths: all attached. `~/...` and the Mac
  path stayed text. Typed absolute stayed text. [O] `local-attachment.ts` differs between 2.0.29 and 2.2.5 only by
  an added `path` field and `promptFilesForServer` helper (a diff of the two tags); the `onPaste` handler is the
  same shape. [S]
- A first-run "what's new" dialog on `jupiter` swallowed my first pastes; dismiss it before testing. (Noted so the
  matrix is not misread; the failing first rows were rerun and passed.) [O]
- What the attachment carries: the TUI reads the bytes and builds a `data:<mime>;base64,...` URI. On 2.2.5 it also
  keeps the file path and, when the server is local, passes a `file://` source (`promptFilesForServer`). [S] So
  the **bytes are read by the TUI process** on the machine it runs on.

## 2. Clipboard image paste

### Claude Code

- Key: Ctrl+V, or Cmd+V in iTerm2; Alt+V on Windows and WSL (both bound on WSL). It inserts an `[Image #N]`
  chip. [D] code.claude.com/docs/en/interactive-mode, keyboard shortcuts table, row "Paste image from
  clipboard"; common-workflows "Work with images" says Ctrl+V, Alt+V on Windows and WSL.
- Which OS tools it shells out to is not documented and the source is not public. [U] What would settle it: strace
  of the process on Linux, or Anthropic documentation.
- Over ssh: Ctrl+V on `jupiter` (a Linux machine reached over ssh, no clipboard tools) printed
  "no image over SSH · copy the file over, paste its path" and added nothing. [O] This matches changelog 2.1.314:
  "Fixed pasting an image over SSH to a Linux machine doing nothing silently; the prompt now says to copy the file
  over and paste its path". [D] The detection is of the ssh session, not of a missing clipboard tool (I did not
  test a Linux box with a display). [U]
- Cmd+V of an image into the pane is the terminal's job: a terminal pastes text, so an image-only clipboard gives
  Claude Code an empty or no paste, not an image (Claude Code listens for Ctrl+V). [U] I did not put an image on the
  Mac clipboard (it would overwrite the user's clipboard).

### opencode

- Key: `prompt.paste` is bound to `ctrl+v` with `preventDefault: false`
  (`packages/tui/src/config/keybind.ts:209`, v2.2.5; same in 2.0.29 at line 205). [S] The same action also runs
  when the terminal delivers an **empty** bracketed paste (Windows Terminal before 1.25 does this for an
  image-only clipboard): the `onPaste` handler checks `event.bytes.byteLength === 0` and dispatches
  `prompt.paste`. [S] (`index.tsx` around line 1811.)
- The handler calls `clipboard.read()` and, if the MIME is `image/*`, adds an attachment named `clipboard` with a
  `data:image/png;base64,...` URI; if `text/plain` it goes through the same path-recognition as a paste. [S]
  (`index.tsx:471-492`.) `read()` asks for `["image/png", "text/plain"]` in that order
  (`packages/tui/src/clipboard.ts:26`). [S]
- How it reads: through `@opentui/core`'s `createHostClipboard`, a **native in-process** backend, not a spawned
  `pngpaste`, `xclip`, `wl-paste` or PowerShell. The native sources are per platform: `macos.zig`, `x11.zig`,
  `wayland-protocol.zig` and `linux.zig`, `windows.zig` and `windows-dib.zig`, in
  `sst/opentui` `packages/native/src/clipboard/` at v0.8.0. [S] The Linux backend keys off `WAYLAND_DISPLAY` or
  `WAYLAND_SOCKET` and `DISPLAY` (`linux.zig:13-14`). [S] Defaults: 1 s timeout, 8 MiB read cap, 64 Mpx image cap
  (`host-clipboard.internal.ts:9-11`). [S]
- Over ssh: `createClipboard().read()` only calls `host.read` (`packages/core/src/lib/clipboard.ts:230`); the
  terminal adapter has `writeText` and `clear` (OSC 52) but **no read**. [S] So opencode's Ctrl+V reads the
  clipboard of the machine **running opencode** (the remote), never the ssh client's. On `jupiter` (no `DISPLAY`,
  no `WAYLAND_DISPLAY`) Ctrl+V in v2.0.29 produced no attachment and no visible message within 0.6 s and 3 s. [O]
  The 2.0.29 TUI uses `@opentui/core` 0.5.17, whose `createClipboard().read` is also `host.read(options)` only
  (checked in the published package). [S]

## 3. `@` references

### Claude Code

- Docs: "Use `@` to quickly include files or directories"; a referenced file's content is included when it fits the
  Read tool's limit (25,000 tokens by default); a text file over 256 KB is not included; paths can be relative or
  absolute; Tab or Enter accepts the highlighted path. [D] code.claude.com/docs/en/common-workflows, "Reference
  files and directories". Nothing there mentions images.
- `@plain.png` (typed `@plain.png`, Tab to accept, then a question) on `jupiter`: the transcript showed
  `Read plain.png (168 bytes)` and the answer `Blue.`. [O] So an `@` reference to an image reaches the model, via
  the Read tool. Whether it shows as `[Image #N]` is not the case (no chip). [O]
- A bracketed paste of `@/abs/x.png` text (no autocomplete) stayed plain text in the prompt; I did not submit it,
  so whether the model would read it is [U].

### opencode

- Docs: `@` does a fuzzy file search in the current working directory; "The content of the file is added to the
  conversation automatically". [D] opencode.ai/docs/tui, "File references". Nothing there about images.
- Typing `@plain` showed a `plain.png` suggestion; Tab produced `@plain.png`; after adding text and Enter, the sent
  message showed a `file plain.png` attachment. [O] (v2.2.5 local; the model call itself failed on a dead
  provider tunnel, so model receipt through this exact route is [U].)
- Server side: the mention becomes an attachment with a `file://` URI (`createFilePart`,
  `packages/tui/src/component/prompt/autocomplete.tsx:261`). The server reads it
  (`packages/core/src/session/prompt.ts:171` `readFileAttachment`, only `file:` URIs; HTTP and HTTPS are
  rejected), detects the type **from the bytes** (`Mime.detect`, `packages/core/src/mime.ts`, line 125 in
  `prompt.ts`), and normalises images. [S] Documented: media type is detected from bytes, so renaming does not
  help. [D] opencode.ai/docs/attachments, "Formats". So an `@` image reference is an image attachment, read by
  the server on its own machine. [S]
- A pasted `@/abs/x.png` stayed `@/private/.../plain.png` text with "No matching files, agents, or references"
  in the menu. [O] No `@` parse happens in the paste path. [S]

## 4. Formats, limits, resizing, token cost

### Claude Code (and the Claude API under it)

- Formats: JPEG, PNG, GIF, WebP; animation unsupported, first frame only. [D]
  platform.claude.com/docs/en/build-with-claude/vision, "Supported formats".
- Limits: 8000x8000 px per image; 100 images per request for 200k-context models, 600 otherwise; 10 MB per image
  (base64) on the first-party API, 5 MB on Bedrock; 128 MB request cap on the first-party API; stricter
  per-image dimensions above 20 images per request. [D] same page, "Request limits".
- Resizing and cost: long edge capped at 1568 px (1,568 visual tokens) on standard models, 2,576 px (4,784 tokens)
  on Claude 4.7 and later; cost is `ceil(w/28) * ceil(h/28)` visual tokens before capping; a 1000x1000 image is
  1,296 tokens. [D] same page, "Resolution and token cost".
- Claude Code specifics from the changelog: pasted and attached images are saved where Claude can open them as
  files without a permission prompt (2.1.275); Read deny rules apply to pasted or dragged image paths (2.1.290);
  the Read tool refuses image sources over 128 MB (2.1.302); a remote session sends a scaled-down copy of PNG,
  JPEG, WebP over 8000 px (2.1.287). [D] code.claude.com/docs/en/changelog. HEIC, SVG, TIFF and BMP attachments in the
  Slack integration arrive as files, not images (2.1.302). [D]

### opencode

- Types sent to the model: PNG, JPEG, GIF, WebP; PDF (needs a PDF-capable model); SVG is treated as text; AVIF, BMP,
  audio and video are not sent; type is detected from the bytes. [D] opencode.ai/docs/attachments, "Formats".
- Limits: 20 MiB decoded per attachment (`MAX_ATTACHMENT_BYTES`,
  `packages/core/src/session/prompt.ts:223`); the pasted-path route also refuses over 20 MiB
  (`local-attachment.ts:6`). [S][D]
- Resizing: on by default; images over 2000x2000 px or 5 MiB base64 are resized (`packages/core/src/image.ts:65-67`);
  config under `media.image` with `auto_resize`, `max_width`, `max_height`, `max_base64_bytes`. [S][D]
  opencode.ai/docs/attachments, "Images". The docs and the repo disagree on the key name: the website attachments
  page and `config/plugin/image.ts` use `media.image`, while `packages/web/src/content/docs/config.mdx:426` at
  v2.2.5 says `attachment.image`. [D][S] I did not test which is honoured. [U]
- Token cost is the provider's; opencode adds none beyond the resize. [D] (docs: "Provider image limits apply after
  OpenCode processing").
- Needs a model with image input; a text-only model "may reject the request". [D] same page, warning.

## 5. The file is on a remote machine

- Claude Code on `jupiter` pasted a **Mac** path (`/private/tmp/.../plain.png`): stayed text, no attachment. [O]
  Same for opencode 2.0.29. [O] A remote path to a real file on `jupiter` attached. [O] So both tools read the
  filesystem of the machine they run on, and a dropped file's local path is useless to an agent over ssh. [S] for
  opencode (`Bun.file` in the TUI process; server `readFileAttachment` on the server's machine).
- Clipboard over ssh: Claude Code says it cannot (message above). [O][D] opencode reads only the remote host's
  clipboard; OSC 52 is write-only in its adapter. [S] I found no tool or flag in either that fetches an image from
  the ssh client's clipboard (OSC 52 read for images, or a forwarded clipboard). [U] What would settle it:
  Anthropic stating it, or a test with an X11-forwarded `ssh -X` session on a Linux host with a display.
- Using opencode against a remote server (`opencode --server <url>` or `attach`): the TUI reads local files and
  sends `data:` URIs, so a **pasted path** attachment travels with its bytes. [S] A `file://` mention is read by the
  **server**, on the server's disk. [S] Not tested. [U]

## 6. Everything else

- Claude Code headless: `--input-format stream-json` (with `--print`) accepts user messages whose content has
  `type: "image"` blocks with `source: {type: "base64", media_type, data}`; an image block with no `source` is
  replaced by a text note. [D] code.claude.com/docs/en/agent-sdk/streaming-vs-single-mode; `claude --help` lists
  `--input-format` and a `--file file_id:path` flag for cloud file resources (not a local image attach). [O]
  `claude -p "look at /path/x.png"` relies on the model calling the Read tool. [U]
- Claude Code `CLAUDE_CODE_DISABLE_FILE_MENTIONS=1` turns off `@` file inclusion. [D] common-workflows.
- opencode headless: `opencode run -f, --file <path> "prompt"` attaches a file. [D] `opencode run --help`
  (v2.2.5). Run locally with `-m zai/glm-4.6v-flash -f plain.png`, the model answered "blue" (the image was blue);
  run on `jupiter` over ssh with the same model and a remote file, the model answered "blue" for `plain.png` and
  "orange" for the pure-red `red shot.png` (so it received an image but its colour naming is loose). [O] This
  proves the attachment reaches a model in the `run` route; I did not submit the pasted-path TUI route to a
  model because the only vision-capable configured providers were not usable without changing settings (the
  Anthropic key stored locally returned "invalid x-api-key"). The TUI route's wire format is a `data:` URI that the
  server decodes in `materializeAttachment` (`prompt.ts:92-140`). [S]
- opencode config: `media.image.*` (above). Keybind `input_paste` / `prompt.paste` can be rebound. [S]
  (`config/keybind.ts`, `config/v1/keybind.ts:175`.)
- opencode shows an image preview when `config.prompt.image_preview` is on and an attachments dialog on
  `<leader>i` (`prompt.images.view`). [S]

## Side by side

| | Claude Code | opencode |
| --- | --- | --- |
| Recognition trigger | Paste event only [O] | Paste event only [S][O] |
| Bracketed paste needed | Yes [O] | Yes [S][O] |
| Absolute path | Yes | Yes |
| Escaped, quoted, raw spaces | Yes | Yes |
| `file://` | No | Yes (percent-decoded; empty or localhost host only) |
| Relative path | No | Yes (to the TUI's cwd) |
| `~` | No | No |
| Several paths in one paste | Space-separated, escaped | Whitespace-separated (32 max), each escaped or quoted |
| Path inside other text | No | No |
| `@image` | Yes (typed, via autocomplete, Read tool) | Yes (typed, via autocomplete, server reads) |
| Clipboard key | Ctrl+V (Cmd+V iTerm2, Alt+V Windows and WSL) | Ctrl+V |
| Clipboard over ssh | Refused, tells you to paste a path | Reads the remote machine's clipboard only |
| Headless | stream-json image blocks | `run -f` |

## Implications for GoRex and issue #8

**What to paste, per tool.** One form serves both: the absolute path of the file on the machine running the agent,
backslash-escaped (spaces, quotes and other shell-special characters), as the only content of one bracketed
paste, with an optional trailing space. Verified on Claude Code 2.1.322 and opencode 2.0.29 on `jupiter`, and on
opencode 2.2.5 locally (matrix above).

- **Path form:** absolute, escaped. Do not use `file://` (Claude Code ignores it), relative or `~` paths (neither
  tool handles `~`; only opencode handles relative). For a drop of several images, one paste with the escaped paths
  separated by single spaces attached all of them in both tools. Do not separate with newlines (Claude Code
  ignores them; opencode needs every path escaped or quoted).
- **No `@`:** `@` is a typing and autocomplete feature in both tools; a pasted `@/abs/path` is just text in both
  and was not shown to attach. Open question 1 in #8 is answered: one form is enough, the agent does not need to be
  told apart.
- **Bracketed paste is required, not just safe:** neither tool attaches a path that arrives as keystrokes. GoRex
  must send `ESC[200~ ... ESC[201~` when the pane's program has enabled mode 2004 (both do, and the existing
  "bracketed when the program asks" rule in #8 is right). If a pane has not enabled 2004 (a plain shell), the path
  goes in as text, which is the right result for a shell.
- **Trailing space:** harmless to both. A trailing newline was also accepted but must not be sent: it risks
  submitting the prompt in a pane without bracketed paste.
- **One path per paste, nothing else in the paste.** Text around the path defeats recognition in both tools. If GoRex
  ever wants to add a caption, send it as a separate paste.
- **Remote file lifetime and location (settled by the follow-up experiments below).** Neither tool needs the original
  file once the paste has been delivered: Claude Code copies it at the paste [O], and opencode keeps the bytes in the
  prompt [S][O]. The file must exist at the instant the paste lands, and not after. GoRex keeps `~/.gorex/drops`
  copies for 24 hours, which only covers a paste that did not happen (for example the ssh ended during the copy);
  the earlier 7 days is not needed.
- **Clipboard-image route cannot work over ssh in either tool.** Claude Code refuses over ssh; opencode reads only
  the remote host's clipboard (none on a headless box). GoRex's design in #8 (read the image on the Mac, copy it to
  the host, paste the path) is the only route, and the right one. No OSC 52 image read exists in either tool as
  far as this research found.
- **Local panes:** paste the escaped local absolute path in a bracketed paste. Same result.
- **Spaces in the drops directory:** do not create one; use names without spaces where GoRex controls the
  name (already "sanitized names").
- **`file://` or `~` fallbacks are not needed.** Always send the absolute path GoRex wrote on the remote (resolve
  `~` to the remote home when creating the path; `$HOME` from the ssh copy step).

**Suggested changes to the issue's scope.**

1. Resolve the "UNVERIFIED" opencode claim: close the opencode item as verified for 2.0.29 and 2.2.5 (path in
   bracketed paste attaches; `@` is not required; the clipboard route is out). Update acceptance criteria to name
   the form: "absolute, backslash-escaped path, one bracketed paste".
2. Drop open question 1 (tool detection by foreground program or title). No per-tool form is needed; a per-pane
   profile override for the form is not justified by this evidence.
3. Add an acceptance test that a path pasted with **no** bracketed paste is not claimed to attach, and document that
   a pane whose program never enabled mode 2004 gets text.
4. In the README limits, state the limits of the tools themselves: Claude Code needs JPEG, PNG, GIF or WebP; opencode
   adds PDF and recognises avif and svg on the paste route but does not send AVIF or BMP to the model; both resize
   large images (opencode above 2000 px or 5 MiB base64 by default).
5. Consider a size note for large PNGs: opencode refuses over 20 MiB at the paste step; GoRex's 64 MB cap is above
   that, so a large drop could copy and then fail to attach in opencode with no error shown by GoRex. Either lower
   the cap for images, or add a README line. [S]
6. For the `drop` debug verb: the headless matrix here used `tmux paste-buffer -p`; the same `ESC[200~...ESC[201~`
   sequence is what the verb needs to emit to a pane.

## Follow-up experiments: when is the file read?

Run later on 2026-11-11 (about 10:25 AM PST), same method as above (private tmux server, `ssh -t jupiter 'cd ~ && exec
<tool>'`, bracketed paste of the absolute path, screen read with `capture-pane`). Test images were 96x96 solid PNGs
generated on `jupiter` (purple, brown, teal).

**Claude Code 2.1.322 on `jupiter`.**
- **[O] Copied at the paste.** Pasting `.../e1.png` gave `[Image #1]`, and within about a second a copy existed at
  `~/.cache/tmp/claude/claude-<uid>/<cwd-with-dashes>/<session-id>/images/1.png` (here
  `claude-1000/-home-andre-angelantoni/<session-id>/images/1.png`). No other file of that name or size appeared.
- **[O] Original not needed after the paste.** Deleted `e1.png` after `[Image #1]` appeared and before submitting; the
  prompt "Which single color fills this picture? ..." was answered "purple".
- **[O] Not needed after submit either.** Pasted `e2.png`, submitted, got "brown", deleted the file, then asked "In one
  word, what color was that picture?"; it answered "brown" from the conversation.
- Conclusion: for Claude Code the original must exist only until the paste is delivered. [U] Whether the copy is
  removed at the end of the session was not checked (the directory is under `~/.cache`).

**opencode 2.0.29 on `jupiter`.**
- **[O] Attachment survives deletion of the file.** Pasted `e3.png` (`[Image 1]` appeared), deleted the file, and the
  attachment was still shown; submitting produced no file error, and the message showed `[Image 1]` and `file e3.png`.
- **[S]** Matches the source reading above: the bytes are read when the paste arrives and held in the prompt.
- **[O] A text-only model still attaches.** The model configured on `jupiter` is GLM-5.3 (Z.AI), which answered "this
  model does not support image input". So the TUI route to a vision model was not shown to deliver the pixels in this
  run; `opencode run -f` was shown to earlier. [U] Settle by running the paste route with a vision model selected.
  This is a configuration matter for the user, not for GoRex: GoRex gets the image into the prompt either way.

**What this means.** The retention of the remote copies has no agent-side requirement beyond the paste itself. Keep
them short (24 hours) because they are private screenshots left on another machine.

## Open questions

1. A real Finder drag into Terminal.app, iTerm2, Ghostty and kitty: does each emit a bracketed paste of an escaped
   path or a quoted path, or send keystrokes? Not run (headless). Does not affect GoRex (it writes its own paste)
   but affects what a user dropping onto a terminal outside GoRex sees. Settle with a real drop and a bytes capture
   (`cat -v`).
2. Claude Code's exact path-recognition rule (why `file://`, relative and `~` are not accepted). Closed source;
   settle with Anthropic documentation or a later version's changelog.
3. Whether a typed `@/abs/x.png` (via autocomplete) behaves like `@plain.png` in Claude Code, and whether a
   submitted `@abs` pasted text is read by the Read tool. Settle by submitting it.
4. opencode TUI route to a vision model end to end (paste then submit): proven for the `run -f` route only.
   Settle with a TUI model switch to a working vision model, which I did not do because it changes settings.
5. Claude Code's clipboard backend on macOS and Linux (which OS tool, and the exact over-ssh detection). Settle with
   `strace` on Linux and `dtruss` or `fs_usage` on macOS while pressing Ctrl+V.
6. Remote opencode (`--server` or attach) with `file://` versus `data:` handling; and whether `ssh -X`
   changes opencode's clipboard read on a host with a display.
7. opencode `media.image` versus `attachment.image` config key; which is honoured in 2.2.5.
8. (Answered: see "Follow-up experiments". Claude Code copies a pasted image at the paste; the original is not read again.)
9. Terminal emulators that send an empty bracketed paste for an image-only clipboard (Windows Terminal before 1.25):
   opencode handles it by reading the clipboard; Claude Code behaviour [U].

## Cleanup performed

Test files were created in `~/.gorex-research` on `jupiter` and `/private/tmp/claude-501/r/` locally; the remote
directory, the Claude Code project record it created under `~/.claude/projects/`, and the five opencode sessions
the runs created (3 local, 2 remote) were deleted. No settings, auth or config were changed.

The follow-up experiments created two Claude Code sessions, an opencode session and `~/.gorex-research2` on `jupiter`;
the two session files and their cache directories under `~/.cache/tmp/claude/`, the opencode session and the test
directory were deleted afterwards.
