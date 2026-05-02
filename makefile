# Makefile for the Nota LaTeX package and renderer.

DOCS        := a3twosidedemo a4articledemo
LATEX       := pdflatex
RENDER_BIN  := scripts/nota-render/nota-render
RENDER      := $(RENDER_BIN)
GRID_SH     := ./demo-resources/pdf_to_grid.sh
RESIZE_SH   := ./demo-resources/resize_image.sh
RESIZE_PCT  := 20
TEXMFHOME   := $(shell kpsewhich -var-value=TEXMFHOME 2>/dev/null || echo "$$HOME/texmf")
LOCAL_BIN   := $(HOME)/.local/bin

.PHONY: all essentials pdfs grids resize_grid clean clean-generated setup renderer test install-local help

essentials: pdfs
	@echo "Build complete!"

all: pdfs resize_grid
	@echo "Build complete!"

renderer:
	cd scripts/nota-render && go build -o nota-render ./cmd/nota-render

test:
	cd scripts/nota-render && go test ./...
	l3build check

pdfs: renderer $(addsuffix .pdf,$(DOCS))
grids: $(addprefix demo-resources/,$(addsuffix _grid.png,$(DOCS)))
resize_grid: $(addprefix demo-resources/,$(addsuffix _resized_$(RESIZE_PCT).png,$(DOCS))) grids

%.pdf: %.tex
	@echo "==> (1/4) Building $@: LaTeX pass 1"
	$(LATEX) -jobname=$* $<
	@echo "==> (2/4) Building $@: Nota rendering"
	$(RENDER) $*
	@echo "==> (3/4) Building $@: LaTeX pass 2"
	$(LATEX) -jobname=$* $<
	@echo "==> (4/4) Building $@: LaTeX pass 3"
	$(LATEX) -jobname=$* $<

demo-resources/%_grid.png: %.pdf
	@echo "==> Creating grid layout: demo-resources/$*"
	$(GRID_SH) $< demo-resources/$*_grid.png

demo-resources/%_resized_$(RESIZE_PCT).png: demo-resources/%_grid.png
	@echo "==> Creating resized image: demo-resources/$*_resized_$(RESIZE_PCT).png ($(RESIZE_PCT)%)"
	$(RESIZE_SH) demo-resources/$*_grid.png $(RESIZE_PCT)

install-local: renderer
	@echo "==> Installing nota.sty into $(TEXMFHOME)"
	mkdir -p "$(TEXMFHOME)/tex/latex/nota"
	cp nota.sty "$(TEXMFHOME)/tex/latex/nota/nota.sty"
	@echo "==> Installing nota-render into $(LOCAL_BIN)"
	mkdir -p "$(LOCAL_BIN)"
	cp "$(RENDER_BIN)" "$(LOCAL_BIN)/nota-render"
	chmod +x "$(LOCAL_BIN)/nota-render"
	mktexlsr "$(TEXMFHOME)" >/dev/null 2>&1 || true
	@echo "==> Installed. Check with: kpsewhich nota.sty && nota-render --version"

clean-generated:
	@echo "--- Cleaning generated Nota and LaTeX files..."
	rm -rf *-nota *notaaux
	rm -f *.log *.aux *.out *.toc *.nota.aux *.scores.aux *.pdf demo-resources/*.png temp_*.png

clean: clean-generated
	rm -f "$(RENDER_BIN)"

setup:
	@echo "No Python setup is used. Install Go to build scripts/nota-render."

help:
	@echo "Nota Makefile - Available targets:"
	@echo ""
	@echo "  all            - Build demo PDFs and preview grids"
	@echo "  essentials     - Build demo PDFs only"
	@echo "  renderer       - Build the Go nota-render CLI"
	@echo "  test           - Run Go tests and l3build tests"
	@echo "  install-local  - Install nota.sty and nota-render for local testing"
	@echo "  clean          - Remove generated files and renderer binary"
	@echo ""
	@echo "Examples:"
	@echo "  make renderer"
	@echo "  make a4articledemo.pdf"
	@echo "  make install-local"
