package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var pageIndexRE = regexp.MustCompile(`(\d+)`)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nota-gallery: %v\n", err)
		os.Exit(1)
	}
}

func run(raw []string) error {
	if len(raw) == 0 {
		usage()
		return fmt.Errorf("missing command")
	}
	switch raw[0] {
	case "grid":
		args, err := parseGridArgs(raw[1:])
		if err != nil {
			return err
		}
		return grid(args)
	case "resize":
		args, err := parseResizeArgs(raw[1:])
		if err != nil {
			return err
		}
		return resize(args)
	case "overlay":
		args, err := parseOverlayArgs(raw[1:])
		if err != nil {
			return err
		}
		return overlay(args)
	case "-h", "--help", "help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %s", raw[0])
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  nota-gallery grid <in.pdf> <out.png> [--dpi 300] [--tile 4]")
	fmt.Fprintln(os.Stderr, "  nota-gallery resize <in.png> <percent>")
	fmt.Fprintln(os.Stderr, "  nota-gallery overlay <in.pdf> <out.png> --paper-mm WxH [options]")
	fmt.Fprintln(os.Stderr, "    --page N | --all   --dpi 120   --tile 4   --gutter-mm 5")
	fmt.Fprintln(os.Stderr, "    --margin-mm N  or  --top-mm --bottom-mm --left-mm --right-mm")
	fmt.Fprintln(os.Stderr, "    --twoside --inner-mm N --outer-mm N")
}

type gridArgs struct {
	In   string
	Out  string
	DPI  int
	Tile int
}

func parseGridArgs(raw []string) (gridArgs, error) {
	args := gridArgs{DPI: 300, Tile: 4}
	positionals := make([]string, 0, 2)
	for i := 0; i < len(raw); i++ {
		arg := raw[i]
		switch {
		case arg == "--dpi":
			i++
			if i >= len(raw) {
				return args, fmt.Errorf("--dpi requires a value")
			}
			dpi, err := strconv.Atoi(raw[i])
			if err != nil || dpi <= 0 {
				return args, fmt.Errorf("invalid --dpi %s", raw[i])
			}
			args.DPI = dpi
		case strings.HasPrefix(arg, "--dpi="):
			dpi, err := strconv.Atoi(strings.TrimPrefix(arg, "--dpi="))
			if err != nil || dpi <= 0 {
				return args, fmt.Errorf("invalid --dpi %s", arg)
			}
			args.DPI = dpi
		case arg == "--tile":
			i++
			if i >= len(raw) {
				return args, fmt.Errorf("--tile requires a value")
			}
			tile, err := strconv.Atoi(raw[i])
			if err != nil || tile <= 0 {
				return args, fmt.Errorf("invalid --tile %s", raw[i])
			}
			args.Tile = tile
		case strings.HasPrefix(arg, "--tile="):
			tile, err := strconv.Atoi(strings.TrimPrefix(arg, "--tile="))
			if err != nil || tile <= 0 {
				return args, fmt.Errorf("invalid --tile %s", arg)
			}
			args.Tile = tile
		case strings.HasPrefix(arg, "-"):
			return args, fmt.Errorf("unknown option %s", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 2 {
		return args, fmt.Errorf("grid requires <in.pdf> <out.png>")
	}
	args.In = positionals[0]
	args.Out = positionals[1]
	return args, nil
}

type resizeArgs struct {
	In      string
	Percent float64
}

func parseResizeArgs(raw []string) (resizeArgs, error) {
	var args resizeArgs
	if len(raw) != 2 {
		return args, fmt.Errorf("resize requires <in.png> <percent>")
	}
	args.In = raw[0]
	percent, err := strconv.ParseFloat(raw[1], 64)
	if err != nil || percent <= 0 {
		return args, fmt.Errorf("invalid percent %s", raw[1])
	}
	args.Percent = percent
	return args, nil
}

func grid(args gridArgs) error {
	if err := requireTools("pdftoppm", "magick"); err != nil {
		return err
	}
	if _, err := os.Stat(args.In); err != nil {
		return fmt.Errorf("input pdf: %w", err)
	}
	tmp, err := os.MkdirTemp("", "nota-gallery-grid-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	prefix := filepath.Join(tmp, "page")
	if err := runCmd("pdftoppm", "-png", "-rx", strconv.Itoa(args.DPI), "-ry", strconv.Itoa(args.DPI), args.In, prefix); err != nil {
		return err
	}
	pages, err := filepath.Glob(filepath.Join(tmp, "page-*.png"))
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("pdftoppm produced no pages from %s", args.In)
	}
	sort.Slice(pages, func(i, j int) bool {
		return pageIndex(pages[i]) < pageIndex(pages[j])
	})
	if err := os.MkdirAll(filepath.Dir(args.Out), 0o755); err != nil {
		return err
	}

	if rem := len(pages) % args.Tile; rem != 0 {
		w, h, err := imageSize(pages[0])
		if err != nil {
			return err
		}
		blank := filepath.Join(tmp, "blank.png")
		if err := runCmd("magick", "-size", fmt.Sprintf("%dx%d", w, h), "canvas:white", blank); err != nil {
			return err
		}
		for i := rem; i < args.Tile; i++ {
			pages = append(pages, blank)
		}
	}

	var rows []string
	for i := 0; i < len(pages); i += args.Tile {
		row := filepath.Join(tmp, fmt.Sprintf("row-%d.png", i/args.Tile))
		cmd := append([]string{}, pages[i:i+args.Tile]...)
		cmd = append(cmd, "+append", row)
		if err := runCmd("magick", cmd...); err != nil {
			return err
		}
		rows = append(rows, row)
	}
	stack := append([]string{}, rows...)
	stack = append(stack, "-append", args.Out)
	return runCmd("magick", stack...)
}

func resize(args resizeArgs) error {
	if err := requireTools("magick"); err != nil {
		return err
	}
	if _, err := os.Stat(args.In); err != nil {
		return fmt.Errorf("input image: %w", err)
	}
	dir := filepath.Dir(args.In)
	base := filepath.Base(args.In)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	clean := strconv.FormatFloat(args.Percent, 'f', -1, 64)
	out := filepath.Join(dir, fmt.Sprintf("%s_resized_%s%s", name, clean, ext))
	pct := strconv.FormatFloat(args.Percent, 'f', -1, 64) + "%"
	return runCmd("magick", args.In, "-resize", pct, out)
}

type overlayArgs struct {
	In       string
	Out      string
	Page     int
	All      bool
	DPI      int
	Tile     int
	PaperW   float64
	PaperH   float64
	Top      float64
	Bottom   float64
	Left     float64
	Right    float64
	Gutter   float64
	TwoSide  bool
	Inner    float64
	Outer    float64
	hasPaper bool
}

func parseOverlayArgs(raw []string) (overlayArgs, error) {
	args := overlayArgs{Page: 1, DPI: 120, Tile: 4, Gutter: 5}
	positionals := make([]string, 0, 2)
	var margin, top, bottom, left, right, inner, outer *float64
	for i := 0; i < len(raw); i++ {
		arg := raw[i]
		next := func(name string) (string, error) {
			if strings.HasPrefix(arg, name+"=") {
				return strings.TrimPrefix(arg, name+"="), nil
			}
			if i+1 >= len(raw) {
				return "", fmt.Errorf("%s requires a value", name)
			}
			i++
			return raw[i], nil
		}
		switch {
		case arg == "--all":
			args.All = true
		case arg == "--twoside":
			args.TwoSide = true
		case arg == "--page" || strings.HasPrefix(arg, "--page="):
			v, err := next("--page")
			if err != nil {
				return args, err
			}
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				return args, fmt.Errorf("invalid --page %s", v)
			}
			args.Page = n
		case arg == "--dpi" || strings.HasPrefix(arg, "--dpi="):
			v, err := next("--dpi")
			if err != nil {
				return args, err
			}
			dpi, err := strconv.Atoi(v)
			if err != nil || dpi <= 0 {
				return args, fmt.Errorf("invalid --dpi %s", v)
			}
			args.DPI = dpi
		case arg == "--tile" || strings.HasPrefix(arg, "--tile="):
			v, err := next("--tile")
			if err != nil {
				return args, err
			}
			tile, err := strconv.Atoi(v)
			if err != nil || tile <= 0 {
				return args, fmt.Errorf("invalid --tile %s", v)
			}
			args.Tile = tile
		case arg == "--paper-mm" || strings.HasPrefix(arg, "--paper-mm="):
			v, err := next("--paper-mm")
			if err != nil {
				return args, err
			}
			w, h, err := parsePaperMM(v)
			if err != nil {
				return args, err
			}
			args.PaperW, args.PaperH = w, h
			args.hasPaper = true
		case arg == "--margin-mm" || strings.HasPrefix(arg, "--margin-mm="):
			v, err := next("--margin-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--margin-mm", v)
			if err != nil {
				return args, err
			}
			margin = &m
		case arg == "--top-mm" || strings.HasPrefix(arg, "--top-mm="):
			v, err := next("--top-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--top-mm", v)
			if err != nil {
				return args, err
			}
			top = &m
		case arg == "--bottom-mm" || strings.HasPrefix(arg, "--bottom-mm="):
			v, err := next("--bottom-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--bottom-mm", v)
			if err != nil {
				return args, err
			}
			bottom = &m
		case arg == "--left-mm" || strings.HasPrefix(arg, "--left-mm="):
			v, err := next("--left-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--left-mm", v)
			if err != nil {
				return args, err
			}
			left = &m
		case arg == "--right-mm" || strings.HasPrefix(arg, "--right-mm="):
			v, err := next("--right-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--right-mm", v)
			if err != nil {
				return args, err
			}
			right = &m
		case arg == "--inner-mm" || strings.HasPrefix(arg, "--inner-mm="):
			v, err := next("--inner-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--inner-mm", v)
			if err != nil {
				return args, err
			}
			inner = &m
		case arg == "--outer-mm" || strings.HasPrefix(arg, "--outer-mm="):
			v, err := next("--outer-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--outer-mm", v)
			if err != nil {
				return args, err
			}
			outer = &m
		case arg == "--gutter-mm" || strings.HasPrefix(arg, "--gutter-mm="):
			v, err := next("--gutter-mm")
			if err != nil {
				return args, err
			}
			m, err := parseMM("--gutter-mm", v)
			if err != nil {
				return args, err
			}
			args.Gutter = m
		case strings.HasPrefix(arg, "-"):
			return args, fmt.Errorf("unknown option %s", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) != 2 {
		return args, fmt.Errorf("overlay requires <in.pdf> <out.png>")
	}
	if !args.hasPaper {
		return args, fmt.Errorf("overlay requires --paper-mm WxH")
	}
	pick := func(side *float64, name string) (float64, error) {
		if side != nil {
			return *side, nil
		}
		if margin != nil {
			return *margin, nil
		}
		return 0, fmt.Errorf("overlay requires --margin-mm or %s", name)
	}
	var err error
	if args.Top, err = pick(top, "--top-mm"); err != nil {
		return args, err
	}
	if args.Bottom, err = pick(bottom, "--bottom-mm"); err != nil {
		return args, err
	}
	if args.TwoSide {
		if inner == nil || outer == nil {
			return args, fmt.Errorf("--twoside requires --inner-mm and --outer-mm")
		}
		args.Inner, args.Outer = *inner, *outer
	} else {
		if args.Left, err = pick(left, "--left-mm"); err != nil {
			return args, err
		}
		if args.Right, err = pick(right, "--right-mm"); err != nil {
			return args, err
		}
	}
	args.In = positionals[0]
	args.Out = positionals[1]
	return args, nil
}

func overlay(args overlayArgs) error {
	if err := requireTools("pdftoppm", "magick"); err != nil {
		return err
	}
	if _, err := os.Stat(args.In); err != nil {
		return fmt.Errorf("input pdf: %w", err)
	}
	tmp, err := os.MkdirTemp("", "nota-gallery-overlay-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	prefix := filepath.Join(tmp, "page")
	pdftoppm := []string{"-png", "-rx", strconv.Itoa(args.DPI), "-ry", strconv.Itoa(args.DPI)}
	if !args.All {
		pdftoppm = append(pdftoppm, "-f", strconv.Itoa(args.Page), "-l", strconv.Itoa(args.Page), "-singlefile")
	}
	pdftoppm = append(pdftoppm, args.In, prefix)
	if err := runCmd("pdftoppm", pdftoppm...); err != nil {
		return err
	}

	var pages []string
	if args.All {
		matches, err := filepath.Glob(filepath.Join(tmp, "page-*.png"))
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("pdftoppm produced no pages from %s", args.In)
		}
		sort.Slice(matches, func(i, j int) bool {
			return pageIndex(matches[i]) < pageIndex(matches[j])
		})
		pages = matches
	} else {
		src := prefix + ".png"
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("pdftoppm produced no page %d from %s", args.Page, args.In)
		}
		pages = []string{src}
	}

	overlaid := make([]string, 0, len(pages))
	for _, src := range pages {
		w, h, err := imageSize(src)
		if err != nil {
			return err
		}
		left, right := args.Left, args.Right
		if args.TwoSide {
			n := args.Page
			if args.All {
				n = pageIndex(src)
			}
			left, right = twosideMargins(n, args.Inner, args.Outer)
		}
		layout, err := overlayLayoutFrom(w, h, args.PaperW, args.PaperH, args.Top, args.Bottom, left, right, args.Gutter)
		if err != nil {
			return err
		}
		out := src
		if args.All {
			out = filepath.Join(tmp, fmt.Sprintf("over-%d.png", pageIndex(src)))
		}
		if err := drawOverlay(src, out, layout); err != nil {
			return err
		}
		overlaid = append(overlaid, out)
	}

	if err := os.MkdirAll(filepath.Dir(args.Out), 0o755); err != nil {
		return err
	}
	if len(overlaid) == 1 {
		return runCmd("magick", overlaid[0], args.Out)
	}
	return tilePNGs(overlaid, args.Out, args.Tile, tmp)
}

func drawOverlay(in, out string, layout overlayLayout) error {
	return runCmd("magick", in,
		"-fill", "none",
		"-strokewidth", "1",
		"-stroke", "#C4C4C4",
		"-draw", layout.Paper.draw(),
		"-stroke", "#7A7A7A",
		"-draw", fmt.Sprintf("stroke-dasharray 8 5 %s", layout.Gutter.draw()),
		"-strokewidth", "2",
		"-stroke", "#2B5EA7",
		"-draw", layout.TypeBlock.draw(),
		"-strokewidth", "1",
		"-stroke", "#1A1A1A",
		"-draw", layout.Staff.draw(),
		out,
	)
}

func tilePNGs(pages []string, out string, tile int, tmp string) error {
	padded := append([]string{}, pages...)
	if rem := len(padded) % tile; rem != 0 {
		w, h, err := imageSize(padded[0])
		if err != nil {
			return err
		}
		blank := filepath.Join(tmp, "blank.png")
		if err := runCmd("magick", "-size", fmt.Sprintf("%dx%d", w, h), "canvas:white", blank); err != nil {
			return err
		}
		for i := rem; i < tile; i++ {
			padded = append(padded, blank)
		}
	}
	var rows []string
	for i := 0; i < len(padded); i += tile {
		row := filepath.Join(tmp, fmt.Sprintf("row-%d.png", i/tile))
		cmd := append([]string{}, padded[i:i+tile]...)
		cmd = append(cmd, "+append", row)
		if err := runCmd("magick", cmd...); err != nil {
			return err
		}
		rows = append(rows, row)
	}
	stack := append([]string{}, rows...)
	stack = append(stack, "-append", out)
	return runCmd("magick", stack...)
}

func imageSize(path string) (int, int, error) {
	cmd := exec.Command("magick", "identify", "-ping", "-format", "%w %h", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("identify %s: %w", path, err)
	}
	var w, h int
	if _, err := fmt.Sscanf(string(out), "%d %d", &w, &h); err != nil {
		return 0, 0, fmt.Errorf("parse image size %q: %w", out, err)
	}
	return w, h, nil
}

func pageIndex(path string) int {
	base := filepath.Base(path)
	m := pageIndexRE.FindString(strings.TrimSuffix(base, filepath.Ext(base)))
	n, _ := strconv.Atoi(m)
	return n
}

func requireTools(names ...string) error {
	for _, name := range names {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("%s not found on PATH", name)
		}
	}
	return nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
