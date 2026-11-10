"""Generate low-fi wireframes for the GoRex Settings/Profiles issue."""
import os, sys
from xml.sax.saxutils import escape as esc

OUT = sys.argv[1]
os.makedirs(OUT, exist_ok=True)

INK = "#1d1d1f"; MUTED = "#636366"; FAINT = "#8e8e93"; LINE = "#c7c7cc"
FILL = "#f2f2f4"; PANEL = "#ffffff"; SEL = "#e4e4e8"; ACC = "#2f6fde"; NOTE = "#c2410c"
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

    def pill(s, x, y, t, fg, bg, size=11):
        w = len(t) * size * 0.58 + 14
        s.rect(x, y - size - 3, w, size + 8, fill=bg, r=(size + 8) / 2)
        s.text(x + 7, y, t, size, fg, 600)
        return w

    def button(s, x, y, w, t, primary=False, h=28):
        s.rect(x, y, w, h, fill=ACC if primary else PANEL, stroke=None if primary else LINE, r=6)
        s.text(x + w / 2, y + h / 2 + 4.5, t, 13, "#fff" if primary else INK, 500, "middle")

    def note(s, n, x, y, ax=None, ay=None):
        """Numbered callout at (x,y); optional leader to (ax,ay)."""
        if ax is not None:
            s.line(x, y, ax, ay, NOTE, 1.2, "3 3")
            s.circle(ax, ay, 3, NOTE)
        s.circle(x, y, 11, NOTE)
        s.text(x, y + 4.5, str(n), 12, "#fff", 700, "middle")

    def notes(s, x, y, items, w=330):
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

    def svg(s, name):
        body = "\n".join(s.p)
        doc = f'<svg xmlns="http://www.w3.org/2000/svg" width="{s.w}" height="{s.h}" viewBox="0 0 {s.w} {s.h}">\n{body}\n</svg>\n'
        open(os.path.join(OUT, name + ".svg"), "w").write(doc)


# ---------- shared pieces ----------
def sidebar(s, x, y, h, profiles, selected, active):
    s.rect(x + 1, y + 39, 200, h - 40, fill=FILL)
    s.line(x + 201, y + 39, x + 201, y + h)
    s.text(x + 16, y + 64, "PROFILES", 11, FAINT, 700)
    yy = y + 76
    for name in profiles:
        if name == selected:
            s.rect(x + 8, yy, 186, 30, fill=SEL, r=6)
        s.text(x + 20, yy + 20, name, 13, INK, 600 if name == selected else 400)
        if name == "Default":
            s.text(x + 20 + len(name) * 7.4 + 6, yy + 20, "built-in", 11, FAINT)
        if name == active:
            s.circle(x + 180, yy + 15, 4, ACC)
        yy += 34
    s.line(x + 1, y + h - 44, x + 201, y + h - 44)
    s.text(x + 16, y + h - 17, "+  New Profile…", 13, ACC, 500)


def appearance_cards(s, x, y, chosen, inherited=False):
    labels = [("System", "sys"), ("Light", "light"), ("Dark", "dark")]
    for i, (lbl, kind) in enumerate(labels):
        cx = x + i * 112
        on = lbl == chosen
        s.rect(cx, y, 100, 64, fill=FILL, stroke=(FAINT if inherited else ACC) if on else LINE, r=8,
               sw=2.5 if on else 1, dash="4 3" if on and inherited else None)
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
    s.circle(sx + sw * frac, y + 14, 9, PANEL, LINE)
    s.rect(sx + sw + 12, y, 28, 28, fill=PANEL, stroke=LINE, r=6); s.text(sx + sw + 26, y + 19, "+", 16, INK, 500, "middle")
    s.text(sx + sw + 52, y + 19, val, 13, MUTED if inherited else INK, 600)
    s.text(sx, y + 40, "8", 10.5, FAINT); s.text(sx + sw, y + 40, "32", 10.5, FAINT, anchor="end")


def preview(s, x, y, w, h, dark, size):
    bg, fg, dim = ("#231e27", "#f1eef3", "#a39aa8") if dark else ("#fbf7f9", INK, MUTED)
    s.rect(x, y, w, h, fill=bg, stroke=LINE, r=8)
    s.text(x + 12, y + 20, "PREVIEW", 10, dim, 700)
    lines = [("~/Projects/gorex", "main"), ("$ go test ./...", None), ("ok   gorex           0.42s", None)]
    yy = y + 30 + size
    for a, b in lines:
        s.text(x + 14, yy, a, size, fg, 400, font=MONO)
        if b: s.text(x + 14 + len(a) * size * 0.61 + 10, yy, b, size, "#34a853", 600, font=MONO)
        yy += size * 1.55


# ---------- 1: first launch ----------
s = S(1240, 640, "1 · Settings on first launch — only Default",
      "GoRex ▸ Settings… (⌘,)   ·   Windows / Linux: Edit ▸ Settings… (Ctrl+,)")
WX, WY, WW, WH = 32, 92, 760, 520
s.window(WX, WY, WW, WH, "Settings")
sidebar(s, WX, WY, WH, ["Default"], "Default", "Default")
MX = WX + 226
s.text(MX, WY + 76, "Default", 20, INK, 700)
s.pill(MX + 84, WY + 75, "Active", ACC, "#e3ecfb")
s.text(MX, WY + 98, "Every other profile follows these values unless it overrides them.", 12.5, MUTED)
s.line(MX, WY + 114, WX + WW - 24, WY + 114)
s.text(MX, WY + 142, "Appearance", 13, INK, 700)
appearance_cards(s, MX, WY + 154, "System")
s.text(MX, WY + 278, "Text size", 13, INK, 700)
text_size(s, MX, WY + 290, "11.6 pt", (11.6 - 8) / 24)
s.text(MX + 368, WY + 309, "Actual Size  ⌘0", 12, FAINT)
preview(s, MX, WY + 352, WW - 250, 110, False, 12)
s.text(MX, WY + WH - 18, "Changes apply and save as you make them.", 11.5, FAINT, italic=True)
s.note(1, WX + 120, WY + 150, WX + 150, WY + 115)
s.note(2, WX + 120, WY + WH - 60)
s.note(3, MX + 340, WY + 196, MX + 310, WY + 196)
s.note(4, MX + 420, WY + 262, MX + 340, WY + 286)
s.note(5, WX + WW - 40, WY + 372)
s.notes(830, 120, [
    ["Default is built in", "No rename, no delete, no ⋯ menu.", "Blue dot = the active profile."],
    ["New Profile… pinned to the bottom", "Always visible, even with many profiles."],
    ["Appearance as preview cards", "Selected card: solid blue ring."],
    ["Slider with − and + over 8–32 pt", "Steps of 1 pt; value shown; ⌘0 hint."],
    ["Live preview of the SELECTED profile", "Matters when it is not the active one:", "the main window would not show it."],
    ["macOS 13+ names it “Settings…”", "(was “Preferences…” before Ventura;", "GoRex's minimum is 13.0)."],
])
s.note(6, 452, 60)
s.svg("1-first-launch")

# ---------- 2: several profiles ----------
s = S(1240, 640, "2 · Several profiles — a selected profile that is not the active one",
      "Inherited values are shown greyed with their source; overridden ones get a marker and a reset link.")
s.window(WX, WY, WW, WH, "Settings")
sidebar(s, WX, WY, WH, ["Default", "Presenting", "Late night"], "Presenting", "Default")
s.text(MX, WY + 76, "Presenting", 20, INK, 700)
s.text(MX + 116, WY + 76, "✎", 15, FAINT)
s.button(WX + WW - 168, WY + 56, 104, "Make Active", primary=True)
s.rect(WX + WW - 56, WY + 56, 32, 28, fill=PANEL, stroke=LINE, r=6); s.text(WX + WW - 40, WY + 75, "⋯", 15, INK, 700, "middle")
s.text(MX, WY + 98, "1 override · everything else follows Default", 12.5, MUTED)
s.line(MX, WY + 114, WX + WW - 24, WY + 114)
# appearance: inherited
s.text(MX, WY + 142, "Appearance", 13, INK, 700)
s.text(MX + 90, WY + 142, "from Default", 12, FAINT, italic=True)
appearance_cards(s, MX, WY + 154, "System", inherited=True)
# text size: overridden
s.circle(MX - 10, WY + 274, 4, NOTE)
s.text(MX, WY + 278, "Text size", 13, INK, 700)
s.text(MX + 74, WY + 278, "Overrides Default (11.6 pt)  ·", 12, MUTED)
s.text(MX + 252, WY + 278, "Reset", 12, ACC, 600)
text_size(s, MX, WY + 290, "18 pt", (18 - 8) / 24, default_frac=(11.6 - 8) / 24)
preview(s, MX, WY + 352, WW - 250, 130, False, 17)
# dropdown of ⋯
dx, dy = WX + WW - 196, WY + 88
s.rect(dx, dy, 172, 102, fill=PANEL, stroke=LINE, r=8)
for i, t in enumerate(["Rename", "Duplicate", "Reset All to Default"]):
    s.text(dx + 14, dy + 22 + i * 22, t, 12.5, INK)
s.line(dx + 8, dy + 74, dx + 164, dy + 74)
s.text(dx + 14, dy + 92, "Delete…", 12.5, "#d70015")
s.note(1, WX + 180, WY + 150, WX + 150, WY + 150)
s.note(2, WX + WW - 116, WY + 46)
s.note(3, MX + 205, WY + 143, MX + 175, WY + 143)
s.note(4, MX - 24, WY + 236, MX - 10, WY + 268)
s.note(5, MX + 112, WY + 330, MX + 103, WY + 310)
s.note(6, dx - 16, dy + 50)
s.notes(830, 120, [
    ["Selected ≠ active", "Selection = what you edit. Blue dot = what", "the terminals use. One active profile, app-wide."],
    ["“Make Active” only when selected ≠ active", "Becomes an “Active” pill once it is."],
    ["Inherited setting", "Greyed + dashed ring + “from Default”.", "Clicking a card creates the override."],
    ["Overridden setting", "Marker dot + what Default is + Reset.", "Replaces a badge on every row."],
    ["Tick on the track marks Default's value", ""],
    ["Profile actions behind ⋯", "Delete asks first (see screen 4);", "Default has no ⋯ menu at all."],
])
s.svg("2-profiles-overrides")

# ---------- 3: new profile sheet ----------
s = S(1240, 640, "3 · New Profile sheet",
      "Opened from “+ New Profile…”, the ⋯ ▸ Duplicate item (pre-selects the copy), or the palette's “Profile: New…”.")
s.window(WX, WY, WW, WH, "Settings")
sidebar(s, WX, WY, WH, ["Default", "Presenting"], "Presenting", "Default")
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
s.text(SX + 46, SY + 270, "Make it the active profile", 13, INK)
s.button(SX + SW - 200, SY + SH - 44, 84, "Cancel")
s.button(SX + SW - 106, SY + SH - 44, 84, "Create", primary=True)
s.note(1, SX + SW + 22, SY + 92, SX + SW - 24, SY + 92)
s.note(2, SX + SW + 22, SY + 196, SX + SW - 120, SY + 196)
s.note(3, SX + SW + 22, SY + 266, SX + 240, SY + 266)
s.note(4, SX + SW + 22, SY + SH - 30, SX + SW - 22, SY + SH - 30)
s.notes(830, 120, [
    ["Name, focused, pre-filled “New Profile”", "Return = Create, Esc = Cancel.", "Duplicate name → inline error, Create disabled."],
    ["Two clear starting points", "Inherit is the default choice; “copy” names", "the profile it copies (the selected one).", "One level only: profiles never inherit", "from each other, only from Default."],
    ["Optional switch on create", ""],
    ["Only modal step in the window", "Everything else saves as you change it."],
])
s.svg("3-new-profile-sheet")

# ---------- 4: switching + delete ----------
s = S(1240, 640, "4 · Switching profiles outside Settings, and deleting one",
      "View ▸ Profile, the command palette, and the delete confirmation.")
# menu
s.rect(32, 92, 360, 26, fill=FILL, stroke=LINE, r=4)
for i, t in enumerate(["GoRex", "Shell", "Edit", "View", "Tabs", "Window"]):
    s.text(48 + i * 56, 110, t, 13, INK, 600 if t == "View" else 400)
s.rect(32, 120, 230, 146, fill=PANEL, stroke=LINE, r=8)
for i, (t, k) in enumerate([("Bigger Text", "⌘+"), ("Smaller Text", "⌘−"), ("Actual Size", "⌘0"), ("Appearance", "▸"), ("Profile", "▸")]):
    yy = 146 + i * 26
    on = t == "Profile"
    if on: s.rect(38, yy - 17, 218, 24, fill=ACC, r=4)
    s.text(50, yy, t, 13, "#fff" if on else INK)
    s.text(244, yy, k, 13, "#fff" if on else FAINT, anchor="end")
s.rect(258, 230, 210, 132, fill=PANEL, stroke=LINE, r=8)
for i, t in enumerate(["Default", "Presenting", "Late night"]):
    s.text(272, 254 + i * 24, "✓" if i == 0 else "", 13, INK); s.text(292, 254 + i * 24, t, 13, INK)
s.line(266, 322, 460, 322)
s.text(292, 346, "New Profile…", 13, INK)
s.note(1, 492, 270, 470, 270)
# palette
PX, PY = 32, 400
s.rect(PX, PY, 440, 200, fill=PANEL, stroke=LINE, r=10)
s.text(PX + 16, PY + 30, "profile", 15, INK)
s.line(PX, PY + 44, PX + 440, PY + 44)
rows = [("Profile: Default", "active"), ("Profile: Presenting", ""), ("Profile: Late night", ""), ("Profile: New…", ""), ("Settings…", "⌘,")]
for i, (t, k) in enumerate(rows):
    yy = PY + 70 + i * 26
    if i == 1: s.rect(PX + 8, yy - 17, 424, 24, fill=SEL, r=5)
    s.text(PX + 20, yy, t, 13, INK)
    if k: s.text(PX + 424, yy, k, 12, FAINT, anchor="end")
s.note(2, 492, PY + 96, 472, PY + 96)
# delete confirm
DX, DY = 560, 110
s.rect(DX, DY, 330, 180, fill=PANEL, stroke=LINE, r=12)
s.text(DX + 24, DY + 38, "Delete “Late night”?", 15, INK, 700)
s.text(DX + 24, DY + 64, "Its overrides are lost. If it is the active", 12.5, MUTED)
s.text(DX + 24, DY + 82, "profile, GoRex switches to Default.", 12.5, MUTED)
s.text(DX + 24, DY + 100, "This cannot be undone.", 12.5, MUTED)
s.button(DX + 330 - 200, DY + 130, 84, "Cancel")
s.rect(DX + 330 - 106, DY + 130, 84, 28, fill="#d70015", r=6); s.text(DX + 330 - 64, DY + 148.5, "Delete", 13, "#fff", 500, "middle")
s.note(3, DX + 350, DY + 80, DX + 330, DY + 80)
s.notes(930, 120, [
    ["View ▸ Profile submenu", "Radio list; rebuilt whenever profiles", "change (the menu is built once today)."],
    ["Palette entries", "One per profile, the active one marked,", "plus “Profile: New…” and “Settings…”."],
    ["Delete always confirms", "Auto-save means no Cancel later. If it", "was active, say where GoRex falls back.", "Shown over the Settings window."],
])
s.svg("4-switching-and-delete")
print("ok")
