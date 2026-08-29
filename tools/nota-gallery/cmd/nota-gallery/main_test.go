package main

import "testing"

func TestParseGridArgs(t *testing.T) {
	args, err := parseGridArgs([]string{"in.pdf", "out.png", "--dpi", "150", "--tile=3"})
	if err != nil {
		t.Fatal(err)
	}
	if args.In != "in.pdf" || args.Out != "out.png" || args.DPI != 150 || args.Tile != 3 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestParseGridArgsRequiresTwoPaths(t *testing.T) {
	if _, err := parseGridArgs([]string{"in.pdf"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseResizeArgs(t *testing.T) {
	args, err := parseResizeArgs([]string{"sheet.png", "20"})
	if err != nil {
		t.Fatal(err)
	}
	if args.In != "sheet.png" || args.Percent != 20 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestParseOverlayArgs(t *testing.T) {
	args, err := parseOverlayArgs([]string{
		"in.pdf", "out.png",
		"--paper-mm=210x297",
		"--margin-mm", "20",
		"--all",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !args.All || args.Top != 20 || args.Left != 20 || args.PaperW != 210 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestParseOverlayArgsTwoside(t *testing.T) {
	args, err := parseOverlayArgs([]string{
		"in.pdf", "out.png",
		"--paper-mm", "297x420",
		"--top-mm", "25",
		"--bottom-mm", "25",
		"--twoside",
		"--inner-mm", "25",
		"--outer-mm", "45",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !args.TwoSide || args.Inner != 25 || args.Outer != 45 || args.Top != 25 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestOverlayLayout(t *testing.T) {
	got, err := overlayLayoutFrom(2100, 2970, 210, 297, 20, 20, 20, 20, 5)
	if err != nil {
		t.Fatal(err)
	}
	if got.TypeBlock.X0 != 200 || got.TypeBlock.X1 != 1900 {
		t.Fatalf("type block %+v", got.TypeBlock)
	}
	if got.Gutter.X0 != 150 || got.Gutter.X1 != 1950 {
		t.Fatalf("gutter %+v", got.Gutter)
	}
}

func TestTwosideMargins(t *testing.T) {
	l, r := twosideMargins(1, 25, 45)
	if l != 25 || r != 45 {
		t.Fatalf("odd page: left=%v right=%v", l, r)
	}
	l, r = twosideMargins(2, 25, 45)
	if l != 45 || r != 25 {
		t.Fatalf("even page: left=%v right=%v", l, r)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if err := run([]string{"nope"}); err == nil {
		t.Fatal("expected error")
	}
}
