# Makefile for the Nota LaTeX package and renderer.

EXAMPLES    := examples/a3example examples/a4example
LATEX       := xelatex
RENDER_BIN  := scripts/nota-render/nota-render
RENDER      := ../../$(RENDER_BIN)
GRID_SH     := ./demo-resources/pdf_to_grid.sh
RESIZE_SH   := ./demo-resources/resize_image.sh
RESIZE_PCT  := 20
TEXMFHOME   := $(shell kpsewhich -var-value=TEXMFHOME 2>/dev/null || echo "$$HOME/texmf")
LOCAL_BIN   := $(HOME)/.local/bin

.PHONY: all essentials examples pdfs grids resize_grid clean clean-generated setup renderer test install-local help

essentials: examples
	@echo "Build complete!"

all: examples resize_grid
	@echo "Build complete!"

renderer:
	cd scripts/nota-render && go build -o nota-render ./cmd/nota-render

test:
	cd scripts/nota-render && go test ./...
	l3build check

examples: renderer $(addsuffix /main.pdf,$(EXAMPLES))
pdfs: examples

examples/%/main.pdf: examples/%/main.tex
	@echo "==> (1/4) Building $@: LaTeX pass 1"
	cd $(dir $<) && TEXINPUTS=.:../..: $(LATEX) -jobname=main main.tex
	@echo "==> (2/4) Building $@: Nota rendering"
	cd $(dir $<) && $(RENDER) main
	@echo "==> (3/4) Building $@: LaTeX pass 2"
	cd $(dir $<) && TEXINPUTS=.:../..: $(LATEX) -jobname=main main.tex
	@echo "==> (4/4) Building $@: LaTeX pass 3"
	cd $(dir $<) && TEXINPUTS=.:../..: $(LATEX) -jobname=main main.tex

grids: $(addsuffix /main_grid.png,$(EXAMPLES))
resize_grid: $(addsuffix /main_grid_resized_$(RESIZE_PCT).png,$(EXAMPLES)) grids

examples/%/main_grid.png: examples/%/main.pdf
	@echo "==> Creating grid layout: $@"
	$(GRID_SH) $< $@

examples/%/main_grid_resized_$(RESIZE_PCT).png: examples/%/main_grid.png
	@echo "==> Creating resized image: $@ ($(RESIZE_PCT)%)"
	$(RESIZE_SH) $< $(RESIZE_PCT)

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
	@echo "--- Cleaning generated Nota and LaTeX example files..."
	rm -rf examples/*/main-nota
	rm -f examples/*/main.aux examples/*/main.log examples/*/main.out examples/*/main.toc
	rm -f examples/*/main.nota.aux examples/*/main.pdf examples/*/main_grid*.png
	rm -f demo-resources/*.png temp_*.png

clean: clean-generated
	rm -f "$(RENDER_BIN)"

setup:
	@echo "No Python setup is used. Install Go to build scripts/nota-render."

help:
	@echo "Nota Makefile - Available targets:"
	@echo ""
	@echo "  all            - Build examples and preview grids"
	@echo "  essentials     - Build examples only"
	@echo "  examples       - Build self-contained example PDFs"
	@echo "  renderer       - Build the Go nota-render CLI"
	@echo "  test           - Run Go tests and l3build tests"
	@echo "  install-local  - Install nota.sty and nota-render for local testing"
	@echo "  clean          - Remove generated files and renderer binary"
	@echo ""
	@echo "Examples:"
	@echo "  make renderer"
	@echo "  make examples"
	@echo "  cd examples/a4example && xelatex main.tex && nota-render main && xelatex main.tex && xelatex main.tex"
