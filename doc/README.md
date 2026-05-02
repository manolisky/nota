# Nota

Nota is a LaTeX package for including Verovio-rendered music scores in documents.

The package writes score render requests during the first LaTeX pass. The
`nota-render` command reads those requests, writes SVG score pages as the
canonical render artifacts, generates PDF compatibility files, and writes small
TeX include files for the next LaTeX pass.
