"""Generate low-fi wireframes for the GoRex Settings/Profiles issue.

    python3 -I gen.py <outdir>
    for f in <outdir>/*.svg; do rsvg-convert -z 2 "$f" -o "${f%.svg}.png"; done
"""
import os, sys
from xml.sax.saxutils import escape as esc

OUT = sys.argv[1]
os.makedirs(OUT, exist_ok=True)

INK = "#1d1d1f"; MUTED = "#636366"; FAINT = "#8e8e93"; LINE = "#c7c7cc"
FILL = "#f2f2f4"; PANEL = "#ffffff"; SEL = "#e4e4e8"; ACC = "#2f6fde"; NOTE = "#c2410c"
CHIP = "#e3ecfb"; RED = "#d70015"
SANS = "Helvetica Neue, Helvetica, Arial, sans-serif"; MONO = "Menlo, monospace"


class S:
    def __init__(s, w, h, title, sub):
        s.w, s.h, s.p = w, h, []
        s.rect(0, 0, w, h, fill="#fafafa")
        s.text(32, 40, title, 20, INK, 700)
        s.text(32, 64, sub, 13, MUTED)

    def rect(s, x, y, w, h, fill=PANEL, stroke=None, r=0, dash=None, sw=1):
        a = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
        if dash: a += f' stroke-dasharray="{dash}"'
        s.p.append(f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="{r}" fill="{fill}"{a}/>')

    def line(s, x1, y1, x2, y2, c=LINE, sw=1, dash=None):
        d = f' stroke-dasharray="{dash}"' if dash else ""
        s.p.append(f'<line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" stroke="{c}" stroke-width="{sw}"{d}/>')

    def circle(s, x, y, r, fill=PANEL, stroke=None, sw=1):
        a = f' stroke="{stroke}" stroke-width="{sw}"' if stroke else ""
        s.p.append(f'<circle cx="{x}" cy="{y}" r="{r}" fill="{fill}"{a}/>')

    def text(s, x, y, t, size=13, c=INK, weight=400, anchor="start", font=SANS, italic=False):
        st = ' font-style="italic"' if italic else ""
        s.p.append(f'<text x="{x}" y="{y}" font-family="{font}" font-size="{size}" fill="{c}" '
                   f'font-weight="{weight}" text-anchor="{anchor}"{st}>{esc(t)}</text>')

    def pill(s, x, y, t, fg=ACC, bg=CHIP, size=11):
        w = len(t) * size * 0.58 + 14
        s.rect(x, y - size - 3, w, size + 8, fill=bg, r=(size + 8) / 2)
        s.text(x + 7, y, t, size, fg, 600)
        return w

    def button(s, x, y, w, t, primary=False, h=28, danger=False):
        fill = RED if danger else ACC if primary else PANEL
        s.rect(x, y, w, h, fill=fill, stroke=None if primary or danger else LINE, r=6)
        s.text(x + w / 2, y + h / 2 + 4.5, t, 13, "#fff" if primary or danger else INK, 500, "middle")

    def note(s, n, x, y, ax=None, ay=None):
        if ax is not None:
            s.line(x, y, ax, ay, NOTE, 1.2, "3 3")
            s.circle(ax, ay, 3, NOTE)
        s.circle(x, y, 11, NOTE)
        s.text(x, y + 4.5, str(n), 12, "#fff", 700, "middle")

    def notes(s, x, y, items):
        s.text(x, y, "Notes", 14, INK, 700)
        y += 26
        for i, lines in enumerate(items, 1):
            s.circle(x + 11, y - 4, 11, NOTE)
            s.text(x + 11, y + 0.5, str(i), 12, "#fff", 700, "middle")
            for j, ln in enumerate(lines):
                s.text(x + 30, y + j * 17, ln, 12.5, INK if j == 0 else MUTED, 600 if j == 0 else 400)
            y += len(lines) * 17 + 14

    def window(s, x, y, w, h, title):
        s.rect(x + 2, y + 4, w, h, fill="#00000014", r=12)
        s.rect(x, y, w, h, fill=PANEL, stroke=LINE, r=12)
        for i, c in enumerate(("#ff5f57", "#febc2e", "#28c840")):
            s.circle(x + 20 + i * 20, y + 18, 6, c)
        s.text(x + w / 2, y + 23, title, 13, INK, 600, "middle")
        s.line(x, y + 38, x + w, y + 38)

    def menu(s, x, y, w, rows, hi=None):
        """rows: list of (label, key) or None for a separator. Returns height."""
        h = 10 + sum(26 if r else 10 for r in rows)
        s.rect(x, y, w, h, fill=PANEL, stroke=LINE, r=8)
        yy = y + 5
        for r in rows:
            if r is None:
                s.line(x + 8, yy + 5, x + w - 8, yy + 5); yy += 10; continue
            lbl, key = r
            on = lbl == hi
            if on: s.rect(x + 6, yy + 1, w - 12, 24, fill=ACC, r=4)
            s.text(x + 18, yy + 18, lbl, 13, "#fff" if on else RED if lbl.startswith("Delete") else INK)
            if key: s.text(x + w - 14, yy + 18, key, 13, "#fff" if on else FAINT, anchor="end")
            yy += 26
        return h

    def svg(s, name):
        body = "\n".join(s.p)
        doc = f'<svg xmlns="http://www.w3.org/2000/svg" width="{s.w}" height="{s.h}" viewBox="0 0 {s.w} {s.h}">\n{body}\n</svg>\n'
        open(os.path.join(OUT, name + ".svg"), "w").write(doc)


WX, WY, WW, WH = 32, 92, 760, 520
MX = WX + 226


def sidebar(s, profiles, selected):
    x, y, h = WX, WY, WH
    s.rect(x + 1, y + 39, 200, h - 40, fill=FILL)
    s.line(x + 201, y + 39, x + 201, y + h)
    yy = y + 52
    if selected == "General":
        s.rect(x + 8, yy, 186, 30, fill=SEL, r=6)
    s.text(x + 20, yy + 20, "General", 13, INK, 600 if selected == "General" else 400)
    s.text(x + 16, yy + 62, "PROFILES", 11, FAINT, 700)
    yy += 74
    for name, uses in profiles:
        if name == selected:
            s.rect(x + 8, yy, 186, 30, fill=SEL, r=6)
        s.text(x + 20, yy + 20, name, 13, INK, 600 if name == selected else 400)
        if name == "Default":
            s.text(x + 20 + len(name) * 7.4 + 6, yy + 20, "built-in", 11, FAINT)
        if uses:
            s.text(x + 184, yy + 20, str(uses), 11.5, FAINT, anchor="end")
        yy += 34
    s.line(x + 1, y + h - 44, x + 201, y + h - 44)
    s.text(x + 16, y + h - 17, "+  New Profile…", 13, ACC, 500)


def appearance_cards(s, x, y, chosen):
    for i, (lbl, kind) in enumerate([("System", "sys"), ("Light", "light"), ("Dark", "dark")]):
        cx = x + i * 112
        on = lbl == chosen
        s.rect(cx, y, 100, 64, fill=FILL, stroke=ACC if on else LINE, r=8, sw=2.5 if on else 1)
        if kind == "sys":
            s.rect(cx + 10, y + 10, 40, 44, fill="#ffffff", stroke=LINE, r=4)
            s.rect(cx + 50, y + 10, 40, 44, fill="#3a3340", r=4)
        elif kind == "light":
            s.rect(cx + 10, y + 10, 80, 44, fill="#ffffff", stroke=LINE, r=4)
        else:
            s.rect(cx + 10, y + 10, 80, 44, fill="#3a3340", r=4)
        s.text(cx + 50, y + 82, lbl, 12.5, INK, 600 if on else 400, "middle")


def text_size(s, x, y, val, frac, default_frac=None, inherited=False):
    s.rect(x, y, 28, 28, fill=PANEL, stroke=LINE, r=6); s.text(x + 14, y + 19, "−", 16, INK, 500, "middle")
    sx, sw = x + 40, 220
    s.rect(sx, y + 12, sw, 4, fill=LINE, r=2)
    s.rect(sx, y + 12, sw * frac, 4, fill=FAINT if inherited else ACC, r=2)
    if default_frac is not None:
        s.line(sx + sw * default_frac, y + 6, sx + sw * default_frac, y + 22, MUTED, 1.5)
    s.circle(sx + sw * frac, y + 14, 9, PANEL, FAINT if inherited else LINE)
    s.rect(sx + sw + 12, y, 28, 28, fill=PANEL, stroke=LINE, r=6); s.text(sx + sw + 26, y + 19, "+", 16, INK, 500, "middle")
    s.text(sx + sw + 52, y + 19, val, 13, FAINT if inherited else INK, 600)
    s.text(sx, y + 40, "8", 10.5, FAINT); s.text(sx + sw, y + 40, "32", 10.5, FAINT, anchor="end")


def preview(s, x, y, w, h, size):
    s.rect(x, y, w, h, fill="#fbf7f9", stroke=LINE, r=8)
    s.text(x + 12, y + 20, "PREVIEW · in the app's current appearance", 10, MUTED, 700)
    yy = y + 30 + size
    for a, b in [("~/Projects/gorex", "main"), ("$ go test ./...", None), ("ok   gorex           0.42s", None)]:
        s.text(x + 14, yy, a, size, INK, 400, font=MONO)
        if b: s.text(x + 14 + len(a) * size * 0.61 + 10, yy, b, size, "#34a853", 600, font=MONO)
        yy += size * 1.55


PROFILES1 = [("Default", 0)]
PROFILES3 = [("Default", 3), ("Presenting", 2), ("Late night", 0)]

# ---------- 1: General ----------
s = S(1240, 640, "1 · Settings ▸ General — app-wide settings (first launch)",
      "GoRex ▸ Settings… (⌘,)   ·   Windows / Linux: Edit ▸ Settings… (Ctrl+,)")
s.window(WX, WY, WW, WH, "Settings")
sidebar(s, PROFILES1, "General")
s.text(MX, WY + 76, "General", 20, INK, 700)
s.text(MX, WY + 98, "These apply to the whole app, whatever profile a pane uses.", 12.5, MUTED)
s.line(MX, WY + 114, WX + WW - 24, WY + 114)
s.text(MX, WY + 142, "Appearance", 13, INK, 700)
appearance_cards(s, MX, WY + 154, "System")
s.text(MX, WY + 268, "Later: notifications, confirm before quitting, … — app behaviour, not terminals.", 12, FAINT, italic=True)
s.text(MX, WY + WH - 18, "Changes apply and save as you make them.", 11.5, FAINT, italic=True)
s.note(1, WX + 150, WY + 112, WX + 120, WY + 67)
s.note(2, MX + 360, WY + 186, MX + 320, WY + 186)
s.note(3, WX + 150, WY + 180, WX + 130, WY + 150)
s.note(4, WX + 120, WY + WH - 60)
s.note(5, 452, 60)
s.notes(830, 120, [
    ["Two layers, as in iTerm2", "General = the app. Profiles = how a", "terminal pane looks and behaves."],
    ["Appearance is app-wide", "Light/Dark is the app's look, not a", "pane's (iTerm2: Appearance ▸ Theme)."],
    ["Default is the only profile at first", "Built in: every pane uses it, every other", "profile inherits from it. Count = panes using it."],
    ["New Profile… pinned to the bottom", ""],
    ["macOS 13+ calls it “Settings…”", "GoRex's minimum is 13.0. Windows and", "Linux have no app menu, so: Edit."],
])
s.svg("1-general")

# ---------- 2: profile with override ----------
s = S(1240, 640, "2 · Settings ▸ a profile — overrides and inheritance",
      "A profile stores only what it changes; everything else follows Default, live.")
s.window(WX, WY, WW, WH, "Settings")
sidebar(s, PROFILES3, "Presenting")
s.text(MX, WY + 76, "Presenting", 20, INK, 700)
s.text(MX + 116, WY + 76, "✎", 15, FAINT)
s.rect(WX + WW - 56, WY + 56, 32, 28, fill=PANEL, stroke=LINE, r=6); s.text(WX + WW - 40, WY + 75, "⋯", 15, INK, 700, "middle")
s.text(MX, WY + 98, "Used by 2 panes · 1 override · everything else follows Default", 12.5, MUTED)
s.line(MX, WY + 114, WX + WW - 24, WY + 114)
s.circle(MX - 10, WY + 138, 4, NOTE)
s.text(MX, WY + 142, "Text size", 13, INK, 700)
s.text(MX + 74, WY + 142, "Overrides Default (11.6 pt)  ·", 12, MUTED)
s.text(MX + 252, WY + 142, "Reset", 12, ACC, 600)
text_size(s, MX, WY + 154, "18 pt", (18 - 8) / 24, default_frac=(11.6 - 8) / 24)
preview(s, MX, WY + 216, WW - 250, 120, 17)
# inherited inset
IX, IY = MX, WY + 360
s.rect(IX - 10, IY - 4, WW - 230, 120, fill="none", stroke=LINE, r=8, dash="5 4")
s.text(IX, IY + 14, "SAME ROW WHEN INHERITED", 10, FAINT, 700)
s.text(IX, IY + 40, "Text size", 13, INK, 700)
s.text(IX + 74, IY + 40, "from Default", 12, FAINT, italic=True)
text_size(s, IX, IY + 52, "11.6 pt", (11.6 - 8) / 24, inherited=True)
# ⋯ menu
dx, dy = 1010, 470
s.text(dx, dy - 10, "⋯ menu", 12, MUTED, 700)
s.menu(dx, dy, 190, [("New Tab with Profile", None), ("Rename", None), ("Duplicate…", None), ("Reset All to Default", None), None, ("Delete…", None)])
s.note(1, MX + 330, WY + 138, MX + 292, WY + 138)
s.note(2, MX + 112, WY + 200, MX + 103, WY + 176)
s.note(3, MX + 245, WY + 392, MX + 175, WY + 392)
s.note(4, WX + WW - 80, WY + 70, WX + WW - 58, WY + 70)
s.note(5, MX + 470, WY + 98, MX + 440, WY + 94)
s.notes(830, 120, [
    ["Overridden: dot + Default's value + Reset", "Reset removes the override; the row then", "follows Default again."],
    ["Tick on the track marks Default's value", ""],
    ["Inherited: greyed, “from Default”", "Moving the slider creates the override.", "Changing Default updates it live."],
    ["Profile actions behind ⋯", "Default has no Rename or Delete.", "Delete always asks (screen 5)."],
    ["Used by N panes", "Profiles belong to panes, not to the app.", "No “active profile”, no “Make Active”."],
])
s.svg("2-profile-overrides")

# ---------- 3: new profile sheet ----------
s = S(1240, 640, "3 · New Profile sheet",
      "From “+ New Profile…”, ⋯ ▸ Duplicate… (copy pre-selected), or the palette's “New Profile…”.")
s.window(WX, WY, WW, WH, "Settings")
sidebar(s, PROFILES3[:2], "Presenting")
s.rect(WX + 1, WY + 39, WW - 2, WH - 40, fill="#00000022")
SX, SY, SW, SH = WX + 170, WY + 39, 420, 330
s.rect(SX + 2, SY + 4, SW, SH, fill="#0000001f", r=10)
s.rect(SX, SY, SW, SH, fill=PANEL, stroke=LINE, r=10)
s.text(SX + 24, SY + 36, "New Profile", 16, INK, 700)
s.text(SX + 24, SY + 68, "Name", 12.5, INK, 600)
s.rect(SX + 24, SY + 76, SW - 48, 30, fill=PANEL, stroke=ACC, r=6, sw=2)
s.text(SX + 34, SY + 96, "Late night", 13, INK)
s.text(SX + 24, SY + 124, "Must be unique and not empty.", 11.5, FAINT)
s.text(SX + 24, SY + 152, "Start from", 12.5, INK, 600)
s.circle(SX + 32, SY + 172, 7, PANEL, ACC, 5)
s.text(SX + 48, SY + 176, "Default — follow it for everything", 13, INK)
s.text(SX + 48, SY + 193, "Later changes to Default flow through.", 11.5, MUTED)
s.circle(SX + 32, SY + 218, 7, PANEL, LINE, 1.5)
s.text(SX + 48, SY + 222, "A copy of “Presenting”", 13, INK)
s.text(SX + 48, SY + 239, "Starts with its overrides (text size 18 pt).", 11.5, MUTED)
s.rect(SX + 24, SY + 258, 14, 14, fill=ACC, r=3); s.text(SX + 31, SY + 270, "✓", 11, "#fff", 700, "middle")
s.text(SX + 46, SY + 270, "Open a new tab with it", 13, INK)
s.button(SX + SW - 200, SY + SH - 44, 84, "Cancel")
s.button(SX + SW - 106, SY + SH - 44, 84, "Create", primary=True)
s.note(1, SX + SW + 22, SY + 92, SX + SW - 24, SY + 92)
s.note(2, SX + SW + 22, SY + 196, SX + SW - 120, SY + 196)
s.note(3, SX + SW + 22, SY + 266, SX + 210, SY + 266)
s.note(4, SX + SW + 22, SY + SH - 30, SX + SW - 22, SY + SH - 30)
s.notes(830, 120, [
    ["Name, focused, pre-filled “New Profile”", "Return = Create, Esc = Cancel.", "Duplicate name → inline error, Create disabled."],
    ["Two starting points", "Inherit is the default choice; “copy” names", "the profile it copies (the selected one).", "One level only: profiles inherit from", "Default, never from each other."],
    ["Optional: try it right away", "Profiles belong to panes, so this opens", "one that uses it."],
    ["The only modal step in Settings", "Everything else saves as you change it."],
])
s.svg("3-new-profile-sheet")

# ---------- 4: profiles on panes ----------
s = S(1240, 640, "4 · Profiles on panes — what a pane uses, and changing it",
      "Every pane has a profile, kept with the layout. New tabs use Default; splits use the focused pane's profile.")
# two pane cards
s.rect(32, 92, 700, 300, fill="#efe1e6", r=12)
for i, (prog, d, prof) in enumerate([("zsh", "~/Projects/gorex", None), ("Node", "~/Projects/site", "Presenting")]):
    px = 44 + i * 344
    s.rect(px, 104, 332, 276, fill=PANEL, stroke=ACC if i else LINE, r=10, sw=2 if i else 1)
    s.rect(px + 12, 114, 18, 18, fill=SEL, r=4)
    s.text(px + 38, 128, prog, 12.5, INK, 600); s.text(px + 38 + len(prog) * 7.6 + 8, 128, d, 12.5, MUTED)
    if prof: s.pill(px + 220, 128, prof)
    s.line(px, 142, px + 332, 142)
    sz = 11 if not prof else 15
    s.text(px + 14, 168, "$ npm run dev" if prof else "$ git status", sz, INK, font=MONO)
    s.text(px + 14, 168 + sz * 1.6, "ready on :3000" if prof else "nothing to commit", sz, MUTED, font=MONO)
# context menu on header
h = s.menu(470, 144, 210, [("Split Right", "⌘D"), ("Split Down", "⇧⌘D"), ("Zoom", "⇧⌘↩"), None, ("Profile", "▸"), ("Edit “Presenting”…", None)], hi="Profile")
F = "\u2007\u2007"
s.menu(684, 236, 170, [(F + "Default", None), ("✓ Presenting", None), (F + "Late night", None), None, (F + "New Profile…", None)])
# shell menu
s.rect(32, 420, 300, 26, fill=FILL, stroke=LINE, r=4)
for i, t in enumerate(["GoRex", "Shell", "Edit", "View", "Tabs"]):
    s.text(48 + i * 56, 438, t, 13, INK, 600 if t == "Shell" else 400)
s.menu(32, 448, 250, [("New Tab", "⌘T"), ("New Tab with Profile", "▸"), None, ("Split Right", "⌘D"), ("Split Down", "⇧⌘D")], hi="New Tab with Profile")
s.menu(286, 474, 160, [("Default", None), ("Presenting", None), ("Late night", None)])
s.note(1, 600, 112, 560, 124)
s.note(2, 870, 270, 856, 270)
s.note(3, 460, 540, 446, 520)
s.note(4, 700, 410, 690, 372)
s.notes(930, 120, [
    ["Profile chip in the pane header", "Shown only when it isn't Default, so a", "one-profile setup looks as it does today."],
    ["Pane ▸ Profile", "Header context menu, and View ▸ Profile", "for the focused pane. Applies at once.", "Kept in layout.json; survives restarts."],
    ["New Tab with Profile", "⌘T stays on Default. Splits inherit", "the focused pane's profile."],
    ["⌘+ ⌘− ⌘0 edit the focused pane's", "profile — every pane using it changes,", "as all panes do today. ⌘0 clears the", "override (on Default: back to 11.6)."],
])
s.svg("4-profiles-on-panes")

# ---------- 5: palette + delete ----------
s = S(1240, 640, "5 · Command palette and deleting a profile",
      "The palette reaches everything the menus do; delete says which panes are affected.")
PX, PY = 32, 100
s.rect(PX, PY, 460, 252, fill=PANEL, stroke=LINE, r=10)
s.text(PX + 16, PY + 30, "profile", 15, INK)
s.line(PX, PY + 44, PX + 460, PY + 44)
rows = [("Pane Profile: Default", ""), ("Pane Profile: Presenting", "this pane"), ("Pane Profile: Late night", ""),
        ("New Tab with Profile: Late night", ""), ("New Profile…", ""), ("Edit Profile “Presenting”…", ""), ("Settings…", "⌘,")]
for i, (t, k) in enumerate(rows):
    yy = PY + 70 + i * 26
    if i == 2: s.rect(PX + 8, yy - 17, 444, 24, fill=SEL, r=5)
    s.text(PX + 20, yy, t, 13, INK)
    if k: s.text(PX + 444, yy, k, 12, FAINT, anchor="end")
s.note(1, 516, PY + 96, 494, PY + 96)
DX, DY = 32, 390
s.rect(DX, DY, 360, 196, fill=PANEL, stroke=LINE, r=12)
s.text(DX + 24, DY + 38, "Delete “Presenting”?", 15, INK, 700)
s.text(DX + 24, DY + 64, "2 panes use it. They switch to Default.", 12.5, MUTED)
s.text(DX + 24, DY + 82, "Its 1 override is lost. This cannot be undone.", 12.5, MUTED)
s.button(DX + 360 - 200, DY + 146, 84, "Cancel")
s.button(DX + 360 - 106, DY + 146, 84, "Delete", danger=True)
s.note(2, DX + 384, DY + 70, DX + 360, DY + 70)
s.notes(600, 120, [
    ["Palette entries", "“Pane Profile: <name>” applies to the focused", "pane (its current one marked). Plus New Tab", "with Profile, New Profile…, Edit Profile…,", "and Settings…"],
    ["Delete always confirms", "Auto-save means no Cancel later. It says how", "many panes use the profile and that they", "fall back to Default. Shown over Settings."],
])
s.svg("5-palette-and-delete")
print("ok")
