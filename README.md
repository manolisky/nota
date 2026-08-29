# Nota

**Nota** is a LaTeX package for including properly formatted musical notation in
LaTeX documents.

Nota uses a two-step build:

1. LaTeX writes score requests to `<jobname>.nota.aux`.
2. `nota-render` renders scores into `<jobname>-nota/`.
3. LaTeX includes the generated graphics on later passes.

The renderer is a Go CLI. SVG pages are the canonical render artifact; PDF
files are generated for `graphicx` inclusion in LaTeX.

Typographic and software design: [doc/spec.md](doc/spec.md).

## Usage

```tex
\usepackage[font=Leipzig,unit=6.875]{nota}

\includescore{score.mei}
\fullscore{full-score.mei}
\examplescore{example.krn}
```

Build a document with:

```sh
xelatex paper.tex
nota-render paper
xelatex paper.tex
xelatex paper.tex
```

Generated files:

```text
paper.nota.aux
paper-nota/
  nota-score-1-page-1.svg
  nota-score-1-page-1.pdf
  nota-score-1.tex
  nota-cache.json
```

## Repository layout

```text
nota-render/          score engine (Go)
latex/                LaTeX package (nota.sty, l3build tests)
scores/               shared music fixtures
examples/latex/       source-only how-to documents
showcase/             committed PDFs and gallery images
tools/nota-gallery/   demo screenshot tooling (not installed)
```

`examples/` is copy-paste source. Local `main.pdf` files there are gitignored.
`showcase/pdf` and `showcase/gallery` are the tracked visual record.

## Examples

```sh
make examples
```

Or one document:

```sh
cd examples/latex/a4-article
TEXINPUTS=.:../../../latex: xelatex main.tex
../../../nota-render/nota-render main
TEXINPUTS=.:../../../latex: xelatex main.tex
TEXINPUTS=.:../../../latex: xelatex main.tex
```

Refresh committed visuals:

```sh
make showcase
```

## Local Development

```sh
make renderer
make test
make install-local
kpsewhich nota.sty
nota-render --version
```

`make install-local` installs `latex/nota.sty` and the `nota-render` binary.
It does not install `nota-gallery`.

## Requirements

- LaTeX with `graphicx` and `kvoptions`
- Go, for building `nota-render`
- Verovio CLI on `PATH` for the current renderer backend
- `rsvg-convert` on `PATH` for SVG-to-PDF page conversion
- MuseScore CLI `mscore` on `PATH` only for `.mscz` inputs
- Poppler (`pdftoppm`) and ImageMagick 7 (`magick`) only for `make showcase`

The renderer is Python-free. `.mscz` files are supported only when MuseScore is
installed separately; Nota does not bundle MuseScore.

## Supported Score Inputs

Nota supports the score formats that Verovio accepts, including MEI, MusicXML,
Humdrum, ABC, Plaine & Easie, and related formats. MuseScore `.mscz` files are
converted through `mscore` before rendering.

## Current Packaging Status

This repo is optimized for GitHub/local install and macOS testing first. CTAN
and official TeX Live packaging are later hardening steps.
