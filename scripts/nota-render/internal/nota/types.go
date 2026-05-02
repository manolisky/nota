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

const Version = "nota-render 0.1.0"

type Config struct {
	Jobname   string
	Jobs      string
	Force     bool
	Verbose   bool
	Backend   RenderBackend
	Converter SVGConverter
}

type Request struct {
	ID          string
	Type        string
	Path        string
	Font        string
	Unit        string
	PaperWidth  string
	PaperHeight string
	Top         string
	Bottom      string
	Right       string
	Left        string
}

type Geometry struct {
	PaperWidthPX  int
	PaperHeightPX int
	TopPX         int
	BottomPX      int
	RightPX       int
	LeftPX        int
	TextWidthPX   int
	TextHeightPX  int
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
		parts := strings.SplitN(line, "|", 11)
		if len(parts) != 11 {
			errs = append(errs, fmt.Errorf("line %d: expected 11 fields, got %d", lineNo+1, len(parts)))
			continue
		}
		requests = append(requests, Request{
			ID: parts[0], Type: parts[1], Path: parts[2], Font: parts[3], Unit: parts[4],
			PaperWidth: parts[5], PaperHeight: parts[6], Top: parts[7], Bottom: parts[8],
			Right: parts[9], Left: parts[10],
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
	mm := f * 0.3527
	return int(mm * 10), nil
}

func GeometryFromRequest(r Request) (Geometry, error) {
	pw, err := LatexDimensionToPixels(r.PaperWidth)
	if err != nil {
		return Geometry{}, fmt.Errorf("paper width: %w", err)
	}
	ph, err := LatexDimensionToPixels(r.PaperHeight)
	if err != nil {
		return Geometry{}, fmt.Errorf("paper height: %w", err)
	}
	top, err := LatexDimensionToPixels(r.Top)
	if err != nil {
		return Geometry{}, fmt.Errorf("top margin: %w", err)
	}
	bottom, err := LatexDimensionToPixels(r.Bottom)
	if err != nil {
		return Geometry{}, fmt.Errorf("bottom margin: %w", err)
	}
	right, err := LatexDimensionToPixels(r.Right)
	if err != nil {
		return Geometry{}, fmt.Errorf("right margin: %w", err)
	}
	left, err := LatexDimensionToPixels(r.Left)
	if err != nil {
		return Geometry{}, fmt.Errorf("left margin: %w", err)
	}
	return Geometry{
		PaperWidthPX: pw, PaperHeightPX: ph, TopPX: top, BottomPX: bottom,
		RightPX: right, LeftPX: left, TextWidthPX: pw - left - right, TextHeightPX: ph - top - bottom,
	}, nil
}

func OptionsForRequest(r Request, g Geometry) VerovioOptions {
	opts := VerovioOptions{
		Font: r.Font, Unit: r.Unit, Footer: "none", MMOutput: true,
		XMLIDChecksum: true, RemoveIDs: true, OutputSMuFLEntities: true,
	}
	switch r.Type {
	case "inline":
		opts.Header = "none"
		opts.PageHeight = 100
		opts.PageWidth = g.TextWidthPX + 100
		opts.AdjustPageHeight = true
	case "fullscore":
		opts.PageWidth = g.TextWidthPX + 100
		opts.PageHeight = g.TextHeightPX + 100
		opts.PageMarginBottom = 120
		opts.JustifyVertically = true
		opts.AdjustPageHeight = true
	case "example":
		opts.Header = "none"
		opts.Breaks = "encoded"
		opts.PageWidth = g.PaperHeightPX
		opts.PageMarginLeft = g.RightPX
		opts.PageMarginRight = g.LeftPX
		opts.AdjustPageHeight = true
		opts.AdjustPageWidth = true
	}
	return opts
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
	h.Write([]byte(r.PaperWidth + "|" + r.PaperHeight + "|" + r.Top + "|" + r.Bottom + "|" + r.Right + "|" + r.Left))
	h.Write([]byte{0})
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil)), nil
}
