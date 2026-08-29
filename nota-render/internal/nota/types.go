package nota

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const Version = "nota-render 0.2.0"

// notationOverhangVU is 5mm in Verovio page units (tenths of a millimetre).
// Staff systems fill \linewidth; braces and barlines sit in this gutter.
// Centering the slightly wider PDF in a \linewidth box hangs that gutter
// into the LaTeX margin.
const notationOverhangVU = 50

type Config struct {
	Jobname   string
	Jobs      string
	Force     bool
	Verbose   bool
	Backend   RenderBackend
	Converter SVGConverter
}

type Request struct {
	ID         string
	Type       string
	Path       string
	Font       string
	Unit       string
	LineWidth  string
	TextHeight string
	Indent     string
}

type Geometry struct {
	LineWidthPX  int
	TextHeightPX int
}

type VerovioOptions struct {
	Font                string
	Unit                string
	Header              string
	Footer              string
	Breaks              string
	PageWidth           int
	PageHeight          int
	PageMarginTop       int
	PageMarginBottom    int
	PageMarginLeft      int
	PageMarginRight     int
	AdjustPageHeight    bool
	AdjustPageWidth     bool
	JustifyVertically   bool
	FirstPageWidth      int
	MMOutput            bool
	XMLIDChecksum       bool
	RemoveIDs           bool
	OutputSMuFLEntities bool
}

type ScoreResult struct {
	Request  Request
	PageSVGs []string
	PagePDFs []string
	TexPath  string
	Cached   bool
}

type RenderBackend interface {
	RenderSVGPages(inputPath string, outputDir string, baseName string, options VerovioOptions) ([]string, error)
}

func ParseJobs(value string) (int, error) {
	if value == "" || value == "auto" {
		n := runtime.NumCPU() - 1
		if n < 1 {
			return 1, nil
		}
		return n, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("invalid --jobs value %q", value)
	}
	return n, nil
}

func ParseManifest(data string) ([]Request, []error) {
	var requests []Request
	var errs []error
	for lineNo, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 8)
		if len(parts) != 8 {
			errs = append(errs, fmt.Errorf("line %d: expected 8 fields, got %d", lineNo+1, len(parts)))
			continue
		}
		requests = append(requests, Request{
			ID: parts[0], Type: parts[1], Path: parts[2], Font: parts[3], Unit: parts[4],
			LineWidth: parts[5], TextHeight: parts[6], Indent: parts[7],
		})
	}
	return requests, errs
}

func LatexDimensionToPixels(value string) (int, error) {
	value = strings.TrimSpace(strings.TrimSuffix(value, "pt"))
	if value == "" {
		return 0, errors.New("empty dimension")
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid LaTeX dimension %q", value)
	}
	// TeX pt → tenths of a millimetre (Verovio page units). 1 pt = 1/72.27 in.
	mm := f * 25.4 / 72.27
	return int(mm * 10), nil
}

func IndentPixels(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return LatexDimensionToPixels(value)
}

func GeometryFromRequest(r Request) (Geometry, error) {
	lw, err := LatexDimensionToPixels(r.LineWidth)
	if err != nil {
		return Geometry{}, fmt.Errorf("line width: %w", err)
	}
	th, err := LatexDimensionToPixels(r.TextHeight)
	if err != nil {
		return Geometry{}, fmt.Errorf("text height: %w", err)
	}
	return Geometry{LineWidthPX: lw, TextHeightPX: th}, nil
}

func OptionsForRequest(r Request, g Geometry) VerovioOptions {
	opts := VerovioOptions{
		Font: r.Font, Unit: r.Unit, Footer: "none", MMOutput: true,
		XMLIDChecksum: true, RemoveIDs: true, OutputSMuFLEntities: true,
	}
	switch r.Type {
	case "inline":
		opts.Header = "none"
		applyNotationOverhang(&opts, g.LineWidthPX)
		// Verovio requires a page height; adjust-page-height shrinks to content.
		opts.PageHeight = 100
		opts.AdjustPageHeight = true
		indentPX, err := IndentPixels(r.Indent)
		if err == nil && indentPX > 0 && indentPX < g.LineWidthPX {
			staff := g.LineWidthPX - indentPX
			opts.FirstPageWidth = staff + 2*notationOverhangVU
		}
	case "full":
		applyNotationOverhang(&opts, g.LineWidthPX)
		opts.PageHeight = g.TextHeightPX
		opts.JustifyVertically = true
	case "example":
		opts.Header = "none"
		opts.Breaks = "encoded"
		applyNotationOverhang(&opts, g.LineWidthPX)
		opts.PageHeight = 100
		opts.AdjustPageHeight = true
		opts.AdjustPageWidth = true
	}
	return opts
}

func applyNotationOverhang(opts *VerovioOptions, lineWidthPX int) {
	opts.PageMarginLeft = notationOverhangVU
	opts.PageMarginRight = notationOverhangVU
	opts.PageWidth = lineWidthPX + 2*notationOverhangVU
}

func CacheKey(projectRoot string, r Request) (string, error) {
	inputPath := filepath.Join(projectRoot, r.Path)
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(Version))
	h.Write([]byte{0})
	h.Write([]byte(r.ID + "|" + r.Type + "|" + r.Path + "|" + r.Font + "|" + r.Unit))
	h.Write([]byte{0})
	h.Write([]byte(r.LineWidth + "|" + r.TextHeight + "|" + r.Indent))
	h.Write([]byte{0})
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil)), nil
}
