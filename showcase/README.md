# Showcase

Tracked visual record of Nota. Rebuild with `make showcase` from the repo root,
then commit when the look should change.

```text
pdf/                 finished example documents
gallery/grids/       contact sheets of those PDFs
gallery/pages/       selected page stills (later)
gallery/features/    geometry-rule overlays (`make overlays`)
```

Example PDFs in `examples/latex/` stay gitignored. Only copies under `pdf/`
are committed.

## Geometry rules

`make overlays` draws four strokes on every page of the committed example
PDFs and tiles them next to the plain grids:

- grey — paper
- blue — type block
- black — `\linewidth`
- dashed — 5 mm brace gutter

Outputs: `gallery/features/a4-article-rules.png`,
`gallery/features/a3-twoside-rules.png`. Not a README hero; look at them
and keep or delete.

## Feature plates later

These are not built yet. When they exist they belong under `gallery/features/`,
not in the LaTeX package:

- `inline-vs-full/` — `\score` vs `\score[full]` vs `\score[example]`
- `input-formats/` — MEI, MusicXML, Humdrum, MuseScore
- `font-and-unit/` — Leipzig/Leland and unit size
- `geometry-rules/` — paper, margins, and text block overlaid on a page
