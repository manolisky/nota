# Nota

**Nota** is a LaTeX package for including properly formatted musical notation in
LaTeX documents.

Nota uses a two-step build:

1. LaTeX writes score requests to `<jobname>.nota.aux`.
2. `nota-render` renders scores into `<jobname>-nota/`.
3. LaTeX includes the generated graphics on later passes.

The renderer is implemented as a Go CLI. SVG pages are the canonical render
artifact; PDF files are generated for normal `graphicx` inclusion in LaTeX.

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

## Examples

The example documents are structured as small real LaTeX projects:

```text
examples/
  shared_score_files/
  a4example/
    main.tex
  a3example/
    main.tex
```

Build both examples from the repo root:

```sh
make examples
```

Or build one example manually:

```sh
cd examples/a4example
xelatex main.tex
nota-render main
xelatex main.tex
xelatex main.tex
```

The generated files stay inside that example project:

```text
main.nota.aux
main-nota/
main.pdf
```

## Local Development

Build the renderer:

```sh
make renderer
```

Build the examples:

```sh
make examples
```

Install into your local TeX tree for testing:

```sh
make install-local
kpsewhich nota.sty
nota-render --version
```

## Requirements

- LaTeX with `graphicx` and `kvoptions`
- Go, for building `nota-render`
- Verovio CLI on `PATH` for the current renderer backend
- `rsvg-convert` on `PATH` for SVG-to-PDF page conversion
- MuseScore CLI `mscore` on `PATH` only for `.mscz` inputs

The renderer is Python-free. `.mscz` files are supported only when MuseScore is
installed separately; Nota does not bundle MuseScore.

## Supported Score Inputs

Nota supports the score formats that Verovio accepts, including MEI, MusicXML,
Humdrum, ABC, Plaine & Easie, and related formats. MuseScore `.mscz` files are
converted through `mscore` before rendering.

## Current Packaging Status

This repo is optimized for GitHub/local install and macOS testing first. CTAN
and official TeX Live packaging are later hardening steps.
