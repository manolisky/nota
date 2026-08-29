package main

import (
	"fmt"
	"strconv"
	"strings"
)

const staffInsetPx = 4.0

type rect struct {
	X0, Y0, X1, Y1 float64
}

func (r rect) draw() string {
	return fmt.Sprintf("rectangle %.1f,%.1f %.1f,%.1f", r.X0, r.Y0, r.X1, r.Y1)
}

type overlayLayout struct {
	Paper     rect
	TypeBlock rect
	Staff     rect
	Gutter    rect
}

func overlayLayoutFrom(pageW, pageH int, paperW, paperH, top, bottom, left, right, gutter float64) (overlayLayout, error) {
	if pageW < 2 || pageH < 2 {
		return overlayLayout{}, fmt.Errorf("page is too small")
	}
	if paperW <= 0 || paperH <= 0 {
		return overlayLayout{}, fmt.Errorf("paper size must be positive")
	}
	if left < 0 || right < 0 || top < 0 || bottom < 0 || gutter < 0 {
		return overlayLayout{}, fmt.Errorf("margins and gutter must be non-negative")
	}
	if left+right >= paperW || top+bottom >= paperH {
		return overlayLayout{}, fmt.Errorf("margins exceed paper size")
	}

	mmX := func(mm float64) float64 { return mm / paperW * float64(pageW) }
	mmY := func(mm float64) float64 { return mm / paperH * float64(pageH) }

	typeBlock := rect{
		X0: mmX(left),
		Y0: mmY(top),
		X1: float64(pageW) - mmX(right),
		Y1: float64(pageH) - mmY(bottom),
	}
	g := mmX(gutter)
	gutterRect := rect{
		X0: max(1, typeBlock.X0-g),
		Y0: typeBlock.Y0,
		X1: min(float64(pageW)-2, typeBlock.X1+g),
		Y1: typeBlock.Y1,
	}
	return overlayLayout{
		Paper:     rect{X0: 1, Y0: 1, X1: float64(pageW) - 2, Y1: float64(pageH) - 2},
		TypeBlock: typeBlock,
		Staff: rect{
			X0: typeBlock.X0,
			Y0: typeBlock.Y0 + staffInsetPx,
			X1: typeBlock.X1,
			Y1: typeBlock.Y1 - staffInsetPx,
		},
		Gutter: gutterRect,
	}, nil
}

func parsePaperMM(value string) (float64, float64, error) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(value)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid --paper-mm %s (want WxH)", value)
	}
	w, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || w <= 0 {
		return 0, 0, fmt.Errorf("invalid --paper-mm width %s", value)
	}
	h, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || h <= 0 {
		return 0, 0, fmt.Errorf("invalid --paper-mm height %s", value)
	}
	return w, h, nil
}

func parseMM(name, value string) (float64, error) {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("invalid %s %s", name, value)
	}
	return v, nil
}

func twosideMargins(page int, inner, outer float64) (left, right float64) {
	// 1-based PDF page: odd = recto (outer on the right).
	if page%2 == 1 {
		return inner, outer
	}
	return outer, inner
}
