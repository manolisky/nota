# Makefile for MEI+LaTeX Project
# Builds a3twosidedemo.tex and a4article.tex with post-processing

# --- Variables ---
DOCS        := a3twosidedemo a4articledemo
# DOCS        := a3twosidedemo a4article
VENV_DIR    := .venv
PYTHON      := $(VENV_DIR)/bin/python
LATEX       := pdflatex
RENDER      := nota/renderScores.py
GRID_SH     := ./demo-resources/pdf_to_grid.sh
RESIZE_SH   := ./demo-resources/resize_image.sh
RESIZE_PCT  := 20

# --- Targets ---
.PHONY: all pdfs grids resize_grid clean setup help

# Default target: build all documents no post-processing
essentials: pdfs
	@echo "Build complete!"

# build all documents with post-processing
all: pdfs resize_grid
	@echo "Build complete!"

# Convenience targets
pdfs: $(addsuffix .pdf,$(DOCS))
grids: $(addprefix demo-resources/,$(addsuffix _grid.png,$(DOCS)))
resize_grid: $(addprefix demo-resources/,$(addsuffix _resized_$(RESIZE_PCT).png,$(DOCS))) grids

# Pattern rule: build PDF from TEX (3-pass system)
%.pdf: %.tex
	@echo "==> (1/3) Building $@: LaTeX pass 1"
	$(LATEX) -jobname=$* $<
	@echo "==> (2/3) Building $@: Python rendering"
	$(PYTHON) $(RENDER) $*
	@echo "==> (3/3) Building $@: LaTeX pass 2"
	$(LATEX) -jobname=$* $<

# Pattern rule: create grid PNG from PDF
demo-resources/%_grid.png: %.pdf
	@echo "==> Creating grid layout: demo-resources/$*"
	$(GRID_SH) $< demo-resources/$*_grid.png

# Pattern rule: create resized PNG from _grid pngs (configurable percentage)
demo-resources/%_resized_$(RESIZE_PCT).png: demo-resources/%_grid.png
	@echo "==> Creating resized image: demo-resources/$*_resized_$(RESIZE_PCT).png ($(RESIZE_PCT)%)"
	$(RESIZE_SH) demo-resources/$*_grid.png $(RESIZE_PCT)

# Clean up all generated files
clean:
	@echo "--- Cleaning up generated files..."
	rm -rf *notaaux
	rm -f *.log *.aux *.out *.scores.aux
	rm -f *.pdf demo-resources/*.png temp_*.png

# Setup the virtual environment
setup: $(VENV_DIR)

$(VENV_DIR): nota/requirements.txt
	@echo "--- Creating Python virtual environment at $(VENV_DIR)..."
	python3 -m venv $(VENV_DIR)
	@echo "--- Installing dependencies..."
	$(VENV_DIR)/bin/pip install -r nota/requirements.txt
	@touch $(VENV_DIR)

# Show available targets
help:
	@echo "MEI+LaTeX Makefile - Available targets:"
	@echo ""
	@echo "  all     - Build everything (PDFs + grids + resize_grid)"
	@echo "  pdfs    - Build PDF files only ($(DOCS))"
	@echo "  grids   - Generate grid PNG montages from PDFs"
	@echo "  resize_grid  - Generate 50% resized PNG previews (first page)"
	@echo "  clean   - Remove all generated files"
	@echo "  setup   - Create Python virtual environment"
	@echo "  help    - Show this help message"
	@echo ""
	@echo "Individual targets:"
	@echo "  <name>.pdf              - Build specific PDF"
	@echo "  <name>_grid.png         - Generate grid layout"
	@echo "  <name>_resized_50.png   - Generate 50% preview"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build everything"
	@echo "  make a3twosidedemo.pdf  # Build just a3twosidedemo.pdf"
	@echo "  make grids              # Create all grid layouts"
