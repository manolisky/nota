# Nota spec

How scores should look in a LaTeX document, and how the package and
renderer cooperate to get that look.

This is the design we are aiming at, not a user manual. Where the current
code differs, the spec says so.

**Now** means the current package already does this.
**Planned** means we have agreed; it is not built yet.
**Later** means we might want it, but not now.

The LaTeX package lives in `latex/nota.sty`. The renderer is the Go
program `nota-render`. Engraving is done by [Verovio](https://www.verovio.org/),
which turns MusicXML, MEI, Humdrum, and similar files into SVG.

## Goal

A score should read as part of the document, the way a display equation
does. It should use the same text block, the same gaps the class already
uses, and TeX’s own page breaks.

It should not bring its own page size, margins, or page numbers.

## How a score gets onto the page

TeX does not draw staves. On the first run, `\score{file.mei}` writes a
request to `<jobname>.nota.aux`. Then you run `nota-render`, which asks
Verovio for SVG, converts each page to PDF, and writes a small
`nota-score-N.tex` that is just `\includegraphics` commands. Later LaTeX
runs include those files.

LaTeX decides *where* the pictures sit. Verovio decides *how the music is
engraved* (note spacing, system breaks, lyric widths).

## Command

**Now.** One command:

```tex
\score{file.mei}                 % in the text flow (default)
\score[full]{file.mei}           % one music page per LaTeX page
\score[example]{file.mei}      % a short excerpt, often in a figure
```

The optional argument is a comma-separated list. Placement and engraving
keys share it:

```tex
\score[full, unit=8]{file.mei}
```

### `\score{file}` (default)

A *system* is one line of music (one or more staves stacked). Verovio
crops each system to its ink. LaTeX treats each system as a box and
breaks pages the way it would between paragraphs. If a paragraph leaves
space at the bottom of the page, the first systems can sit in that space.

Space before and after the whole score is the document’s
`\bigskipamount` (the usual large skip, as around a display).

### `\score[full]{file}`

Each picture is as tall as the type block (see below). Verovio spreads
systems down that page. The last page of the piece may be short, like the
last page of a published score.

### `\score[example]{file}`

Verovio keeps the line breaks encoded in the file, instead of reflowing
to `\linewidth`. Use this for a short excerpt, typically inside a
`figure`.

## Options

Three places you can set engraving keys, plus layout that is **never** an
option because it is already true of the document.

1. **Package** — once for the document:

   ```tex
   \usepackage[font=Leipzig]{nota}
   \usepackage[font=Leipzig, unit=8]{nota}   % lock staff size
   ```

2. **Zone** — from here until the current TeX group ends
   (`\begingroup`, an environment, a minipage). Then the previous values
   return. **Now.**

   ```tex
   \notasetup{font=Leland, unit=8}
   ```

   A chapter of full scores can say `\notasetup{full}` and then
   `\score{act1.mei}` without repeating `[full]`.

3. **Command** — this score only. **Now.**

   ```tex
   \score[full, unit=7]{file.mei}
   ```

If the same key is set in more than one place, the nearest wins:
**command, then zone, then package, then the built-in default.**

`full` and `example` are zone and command only. They are not package
options: a mixed article should default to in-flow.

### Keys you set

- `font` — music (SMuFL) font. Default `Leipzig`. **Now** as a package
  option.
- `unit` — staff size. In Verovio, one unit is half the space between two
  staff lines. If you do not set it, Nota derives it from the current
  LaTeX font size (see Staff size). **Now.**
- `indent` — first-system indent for default `\score`. Off unless you
  set it (`indent=15pt`). Not taken from `\parindent`. **Now.**
  Verovio has no indent flag; see First-system indent.
- `lyricfont` — font for lyrics, tempo, directions. Default `verovio`
  (Verovio’s own text font). Those characters have special widths; a
  random document font often looks wrong. **Planned** as an option;
  `lyricfont=document` (use the document’s roman) is **later**.

### Layout you do not set

At the moment `\score` runs, LaTeX already knows the current line width,
page text height, and skips. The package reads those registers. You do
not pass them as keys.

The *type block* is the rectangle where body text lives: inside the
margins, not the paper. `geometry` and the document class set it.

- `\linewidth` — how wide a staff system is. In a one-column article
  this is the type-block width. In a minipage, a quote, or two-column
  mode it is narrower. Measured when the command runs, so
  `\newgeometry` is picked up. **Now.**
- `\textheight` — how tall a `\score[full]` page is. **Now.**
- `\bigskipamount` — gap around an in-flow score. **Now.**
- `\medskipamount` — gap between systems of an in-flow score. **Now.**
- Ragged vs flush bottom — where leftover space on a short page goes
  (see Skips). The class already chooses this. Nota does not add a
  switch.

## Layout

### Fonts

Music uses `font` (Leipzig by default). **Now.**

Lyrics and other text in the score use Verovio’s text font unless the
user asks otherwise. We do not silently switch them to `\rmfamily`.

### Staff size

If `unit` is not set at package, zone, or command, Nota derives it from
the current font size (`\f@size`) when `\score` runs. An 11pt article
gets a larger staff than a 10pt one. A score inside `\small` or a
footnote shrinks with the text. **Now.**

The scale is fixed so a 10pt document matches today’s default:

```text
unit = 6.875 × (current size in pt / 10)
```

Verovio only accepts `4.5`–`12`. Values outside that range are clamped.

If you set `unit` anywhere, that number is used and size tracking
stops. The renderer always receives a number in the aux file; LaTeX
does the derivation.

### First-system indent

The first system of a piece can sit in from the left, like the first
line of a paragraph. Off by default. Set `indent=` at package, zone, or
command if you want it. **Now** for default `\score`.

Verovio has no `--indent`. For default `\score`, each system is its own
picture, so Nota engraves the first system to `\linewidth` minus the
indent and shifts it right with `\hspace`. Later systems stay full
width.

`\score[full]` puts several systems on one picture. Indenting only
the first of those needs MEI `scoreDef@system.leftmar` (or a later
Verovio flag). That path is **later**; until then `[full]` is flush
left.

Examples in this repo do not set `indent`.

### Margins

Scores follow the type block, not the paper size.

Systems are drawn to `\linewidth`. Braces and the leftmost barline need a
little room past the staff. Verovio adds a 5 mm gutter on left and
right, so the PDF is slightly wider than the type block. LaTeX then
centers that PDF on `\linewidth`, and the gutter hangs into the margin
the way an engraved brace does.

### Skips

Around an in-flow score, the skip is `\bigskipamount`, including its
stretch. That glue is *outside* the music, so the class may still use it
when packing a page.

Between systems, the skip is the *natural* width of `\medskipamount` with
the stretch stripped. Two-sided classes use flush bottom: leftover space
is poured into stretchable glue so every page ends at the same depth. If
the gap between staves can stretch, you get uneven white bands inside the
score and a paragraph parked on the bottom margin. Staves need a fixed
distance; leftover space stays where the class already puts it (a hole
at the bottom if ragged, other document glue if flush). **Now.**

### In the text flow

Default `\score` is a stack of cropped system pictures. TeX turns the
page when the next system does not fit.

`\score[full]` is different: each picture is already `\textheight` tall,
so it occupies a whole type-block page. Gaps between systems on that
page are Verovio’s `--spacing-system` default, not `\medskipamount`.

`\score[example]` keeps the file’s own line breaks and shrinks the
picture to the ink.

## Verovio

Nota should not re-expose every Verovio flag. Most of them are
engraving taste (slur shape, hairpin thickness). The list below is
what matters for *this* spec: layout that should follow the document,
or a small set of score options we may add later.

**Already passed (now):** `--font`, `--unit`, `--page-width`,
`--page-height`, `--page-margin-*`, `--adjust-page-height`,
`--adjust-page-width` (example), `--header none` / `--footer none`,
`--justify-vertically` (full), `--breaks encoded` (example),
`--mm-output`, plus id/SMuFL output flags.

**Map from the document (now):**

- `--unit` from `\f@size` when `unit` is unset
- `--svg-css` when `lyricfont=document` (**later**)
- First-system indent: not a Verovio flag; only if the user sets
  `indent`; narrower first page plus `\hspace` on default `\score`

**Leave at Verovio defaults (we are not wrapping them yet):**

- `--spacing-system` — gap between systems on a `\score[full]` page.
  Trust Verovio. If it ever fights the document, it can become a user
  key; it is not derived from `\medskipamount`.
- `--no-justification`, `--min-last-justification` — last-system
  stretch. Verovio already has a policy (justify if the system is at
  least 80% full).
- `--breaks-no-widow` — packing a leftover measure. Systems are tall;
  we do not fight orphans.
- `--spacing-staff` — gap between staves *inside* a system (piano,
  etc.). Engraving, not the document skip.
- `--lyric-size` — in MEI units, so it already scales with `unit`
- `--condense` — hide empty staves. Useful later as a Nota key, not
  inherited from LaTeX
- `--header` / `--footer` other than `none` — LaTeX owns the page
  style
- `--scale`, `--landscape` — we size in millimetres to the type block

A later `\score[condense]` (or similar) can pass `--condense` without
making Nota a general Verovio frontend.

## What we are not doing

- **Justifying the last system of a fragment to full width.** Verovio
  already decides system justification (`--min-last-justification`).
- **Preventing a single system at the top or bottom of a page.** Systems
  are much taller than a text line. An “orphan” staff is acceptable.
- **Following `\color`.** A score in a blue theorem could pick up the
  text color. **Later.**
- **Page numbers and running heads from Verovio.** LaTeX already has
  those. Verovio’s header is off for in-flow and example scores; its
  footer is always off. **Now.**

## Build protocol

For the renderer and the package, not for document authors.

1. LaTeX writes `<jobname>.nota.aux`.
2. `nota-render` writes SVG (the canonical engraving), PDF for
   `graphicx`, and `nota-score-N.tex`.
3. Later LaTeX runs `\input` those files.

Each aux line is:

```text
ID|TYPE|PATH|FONT|UNIT|LINEWIDTH|TEXTHEIGHT|INDENT
```

`TYPE` is `inline` (default `\score`), `full`, or `example`. Width and
height are TeX `pt` values from `\the\linewidth` and `\the\textheight`.
`UNIT` is always a number: either the user’s `unit` or the value
derived from `\f@size`. `INDENT` is `0.0pt` unless the user set
`indent`. `\score[full]` still receives it but the renderer ignores it.

When `\score` writes the aux, zone and command keys only change the
field values at write time. `lyricfont` stays omitted until that option
exists.

The cache key is renderer version, those fields, and the bytes of the
input score (`nota-render/internal/nota/types.go`).

## Still open

- The exact `lyricfont` values: `verovio` / `document`, or a font name.
- On two-sided pages, whether the brace gutter should hang into the
  inner or outer margin. Today it hangs equally on both sides.
