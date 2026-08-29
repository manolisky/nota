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

func TestRunUnknownCommand(t *testing.T) {
	if err := run([]string{"overlay"}); err == nil {
		t.Fatal("expected error")
	}
}
