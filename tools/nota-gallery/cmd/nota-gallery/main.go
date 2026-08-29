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
