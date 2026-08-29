# Nota spec

Typographic and software design. Canonical package: `latex/nota.sty`.
Renderer: `nota-render/`.

Scores should inherit the document’s type block and TeX’s page builder so they
read as native content, not as a second layout engine pasted in.

## Commands

- `\includescore{file}` — in-flow fragment. Cropped systems as boxes; TeX
  breaks pages. Space before and after from `\bigskipamount`.
- `\fullscore{file}` — type-block-tall pages. Verovio justifies systems
  vertically inside each page. The last page may be short.
- `\examplescore{file}` — encoded breaks (`--breaks encoded`). Usually a
  single cropped system, often inside a float.

LaTeX places boxes. Verovio engraves. Running heads, `\thepage`, and
ragged/flush bottom belong to the class.

## Option layers

Three explicit layers, plus automatic inheritance that is not an option.

- **Package** — `\usepackage[font=Leipzig,unit=6.875]{nota}`
- **Zone** — `\notasetup{font=Leland,unit=8}` using **local** assignments so
  `\begingroup` and environments revert them (same model as `\lstset` /
  `\tikzset`)
- **Command** — `\includescore[unit=7]{file.mei}`

Resolution: **command > zone > package > built-in default**.

Zone and command options are **specified, not implemented**. Today only
package `font` and `unit` exist; the three commands take a path only.

## Inherited context vs options

Measured at the command (current TeX context). Never declared as options:

- `\linewidth` — staff systems fill this. Minipage, twocolumn, quote, and
  `\newgeometry` change it. This is the type-block width, not `\textwidth`
  and not the paper.
- `\textheight` — height of a `\fullscore` page
- `\bigskipamount` / `\medskipamount` — skips around scores and between
  systems
- ragged vs flush bottom (`\@textbottom`) — leftover vertical space

Declared as options (package / zone / command):

- `font` — SMuFL music font (default `Leipzig`)
- `unit` — staff size in MEI units (default `6.875`)
- `lyricfont` — text inside the score (default `verovio`; see Fonts)

## Features

Status: **done** (current behavior), **specified** (agreed, not built),
**later**.

### 1. Fonts

Music: SMuFL via `font`. Default Leipzig. **Done** as a package option.

Lyrics, tempo, and directions stay on **Verovio’s text font by default**.
Those glyphs have special widths; not every document font works as a lyric
font.

Specified, not implemented:

- `lyricfont=verovio` — default; current behavior once the option exists
- `lyricfont=document` — later; document roman via Verovio `--svg-css`

Do not auto-bind lyrics to `\rmfamily`.

### 2. Font size

Staff size is the `unit` option (MEI unit: half the distance between staff
lines). Default `6.875`. **Done** as a package option.

`unit` does not follow `\small` / `\large` / `\f@size` unless we add that
later. Lyrics size stays Verovio (`--lyric-size`) unless `lyricfont=document`
forces a later decision.

### 3. Margins / type block

Inherit the type block, not the paper. **Done.**

Staff systems fill `\linewidth`. A 5 mm brace gutter (`notationOverhangVU` in
`nota-render/internal/nota/types.go`) hangs into the margin: the PDF is
slightly wider than `\linewidth` and is centered with
`\makebox[\linewidth][c]{...}`.

Call-site width is not a separate feature. Measuring `\the\linewidth` at
the command *is* how margins work.

### 4. Skips

**Done.**

- Around `\includescore` (and after `\fullscore`): `\vspace{\bigskipamount}`.
  Document glue; stretch is allowed.
- Between `\includescore` systems: `\parskip` is the natural width of
  `\medskipamount` (`\dimexpr\medskipamount\relax`, no plus/minus).
  Flush-bottom cannot open holes between staves.

### 5. In flow

**Done.**

- `\includescore` / `inline`: tiny Verovio page height, `--adjust-page-height`,
  one cropped system per box. TeX pours them into leftover space and
  chooses page breaks.
- `\fullscore`: `--page-height` = `\textheight`, `--justify-vertically`, no
  height crop. One plate per LaTeX page.
- `\examplescore` / `example`: encoded breaks, crop width and height.

## Non-goals

- Last-system stretch. Verovio justifies systems.
- Widow/orphan systems. Systems are tall; a leftover staff is acceptable.
- Current color following `\color`. Later, not now.
- First-system indent. Verovio’s job if we want it.
- Verovio headers, footers, or page numbers. Already suppressed:
  `--header none` on inline and example; `--footer none` always.

Ragged vs flush bottom is not a Nota switch. TeX owns leftover space. Nota
only froze inter-staff stretch so flush-bottom cannot fight the staves.

## Build protocol

Two-step build:

1. LaTeX writes requests to `<jobname>.nota.aux`.
2. `nota-render` writes SVG (canonical), PDF, and `nota-score-N.tex`.
3. Later LaTeX passes `\input` those files.

Aux line (current):

```text
ID|TYPE|PATH|FONT|UNIT|LINEWIDTH|TEXTHEIGHT
```

`TYPE` is `inline`, `fullscore`, or `example`. `LINEWIDTH` and
`TEXTHEIGHT` are TeX `pt` dimensions measured at the command.

Zone and command options, when implemented, only change what those fields
contain at `\immediate\write` time. `lyricfont` adds a field or stays
omitted until implemented.

Cache key includes renderer version, request fields, and input file
bytes (`nota-render/internal/nota/types.go`).

Placement vs engine: generated includes are `\includegraphics` in a
`\linewidth` box. Verovio does not paginate the LaTeX document.

## Open decisions

- Exact `lyricfont` value names (`verovio` / `document` vs a font name).
- Whether `unit` should later track `\f@size`.
- Twoside inner vs outer for the brace gutter (today it hangs equally).
