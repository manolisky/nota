# Makefile for the Nota LaTeX package and renderer.

EXAMPLES     := examples/latex/a4-article examples/latex/a3-twoside
LATEX        := xelatex
RENDER_BIN   := nota-render/nota-render
GALLERY_BIN  := tools/nota-gallery/nota-gallery
RENDER       := ../../../$(RENDER_BIN)
RENDER_SRCS  := $(shell find nota-render -name '*.go')
SCORE_FILES  := $(wildcard scores/*)
TEXMFHOME    := $(shell kpsewhich -var-value=TEXMFHOME 2>/dev/null || echo "$$HOME/texmf")
LOCAL_BIN    := $(HOME)/.local/bin

.PHONY: all essentials examples pdfs showcase grids clean clean-generated renderer gallery test install-local help

essentials: examples
	@echo "Build complete!"

all: examples
	@echo "Build complete!"

renderer: $(RENDER_BIN)

$(RENDER_BIN): $(RENDER_SRCS)
	cd nota-render && go build -o nota-render ./cmd/nota-render

gallery:
	cd tools/nota-gallery && go build -o nota-gallery ./cmd/nota-gallery

test:
	cd nota-render && go test ./...
	cd tools/nota-gallery && go test ./...
	l3build check

examples: renderer $(addsuffix /main.pdf,$(EXAMPLES))
pdfs: examples

examples/%/main.pdf: examples/%/main.tex latex/nota.sty $(RENDER_BIN) $(SCORE_FILES)
	@echo "==> (1/4) Building $@: LaTeX pass 1"
	cd $(dir $<) && TEXINPUTS=.:../../../latex: $(LATEX) -jobname=main main.tex
	@echo "==> (2/4) Building $@: Nota rendering"
	cd $(dir $<) && $(RENDER) --force main
	@echo "==> (3/4) Building $@: LaTeX pass 2"
	cd $(dir $<) && TEXINPUTS=.:../../../latex: $(LATEX) -jobname=main main.tex
	@echo "==> (4/4) Building $@: LaTeX pass 3"
	cd $(dir $<) && TEXINPUTS=.:../../../latex: $(LATEX) -jobname=main main.tex

showcase: examples gallery
	@echo "==> Publishing showcase PDFs and grids"
	mkdir -p showcase/pdf showcase/gallery/grids
	cp examples/latex/a4-article/main.pdf showcase/pdf/a4-article.pdf
	cp examples/latex/a3-twoside/main.pdf showcase/pdf/a3-twoside.pdf
	$(GALLERY_BIN) grid examples/latex/a4-article/main.pdf showcase/gallery/grids/a4-article.png --dpi 120
	$(GALLERY_BIN) grid examples/latex/a3-twoside/main.pdf showcase/gallery/grids/a3-twoside.png --dpi 120

grids: showcase

install-local: renderer
	@echo "==> Installing nota.sty into $(TEXMFHOME)"
	mkdir -p "$(TEXMFHOME)/tex/latex/nota"
	cp latex/nota.sty "$(TEXMFHOME)/tex/latex/nota/nota.sty"
	@echo "==> Installing nota-render into $(LOCAL_BIN)"
	mkdir -p "$(LOCAL_BIN)"
	cp "$(RENDER_BIN)" "$(LOCAL_BIN)/nota-render"
	chmod +x "$(LOCAL_BIN)/nota-render"
	mktexlsr "$(TEXMFHOME)" >/dev/null 2>&1 || true
	@echo "==> Installed. Check with: kpsewhich nota.sty && nota-render --version"

clean-generated:
	@echo "--- Cleaning generated example files..."
	rm -rf examples/latex/*/main-nota
	rm -f examples/latex/*/main.aux examples/latex/*/main.log examples/latex/*/main.out examples/latex/*/main.toc
	rm -f examples/latex/*/main.nota.aux examples/latex/*/main.pdf examples/latex/*/main.xdv
	rm -f temp_*.png

clean: clean-generated
	rm -f "$(RENDER_BIN)" "$(GALLERY_BIN)"

help:
	@echo "Nota Makefile - Available targets:"
	@echo ""
	@echo "  all            - Build example documents"
	@echo "  examples       - Build self-contained example PDFs (gitignored)"
	@echo "  showcase       - Copy example PDFs and grids into showcase/"
	@echo "  renderer       - Build the Go nota-render CLI"
	@echo "  gallery        - Build the Go nota-gallery demo tool"
	@echo "  test           - Run Go tests and l3build tests"
	@echo "  install-local  - Install nota.sty and nota-render for local testing"
	@echo "  clean          - Remove generated example files and local binaries"
	@echo ""
	@echo "Examples:"
	@echo "  make renderer"
	@echo "  make examples"
	@echo "  make showcase"
	@echo "  cd examples/latex/a4-article && TEXINPUTS=.:../../../latex: xelatex main.tex && ../../../nota-render/nota-render main && TEXINPUTS=.:../../../latex: xelatex main.tex && TEXINPUTS=.:../../../latex: xelatex main.tex"
