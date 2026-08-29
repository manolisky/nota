package nota

import (
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	data := "1|inline|score.mei|Leipzig|6.875|345.0pt|550.0pt\n"
	reqs, errs := ParseManifest(data)
	if len(errs) != 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	if reqs[0].ID != "1" || reqs[0].Type != "inline" || reqs[0].Font != "Leipzig" {
		t.Fatalf("unexpected request: %#v", reqs[0])
	}
	if reqs[0].LineWidth != "345.0pt" || reqs[0].TextHeight != "550.0pt" {
		t.Fatalf("unexpected geometry fields: %#v", reqs[0])
	}
}

func TestParseManifestMalformed(t *testing.T) {
	_, errs := ParseManifest("1|inline|missing\n")
	if len(errs) != 1 {
		t.Fatalf("expected one parse error, got %d", len(errs))
	}
}

func TestLatexDimensionToPixels(t *testing.T) {
	got, err := LatexDimensionToPixels("100.0pt")
	if err != nil {
		t.Fatal(err)
	}
	if got != 351 {
		t.Fatalf("expected 351 (TeX pt), got %d", got)
	}
}

func TestGeometryFromRequest(t *testing.T) {
	g, err := GeometryFromRequest(Request{LineWidth: "100.0pt", TextHeight: "200.0pt"})
	if err != nil {
		t.Fatal(err)
	}
	if g.LineWidthPX != 351 || g.TextHeightPX != 702 {
		t.Fatalf("unexpected geometry: %#v", g)
	}
}

func TestOptionsForInline(t *testing.T) {
	r := Request{Type: "inline", Font: "Leland", Unit: "8"}
	g := Geometry{LineWidthPX: 1000, TextHeightPX: 1400}
	opts := OptionsForRequest(r, g)
	if opts.Header != "none" || !opts.AdjustPageHeight || opts.PageWidth != 1100 {
		t.Fatalf("unexpected inline options: %#v", opts)
	}
	if opts.PageMarginLeft != notationOverhangVU || opts.PageMarginRight != notationOverhangVU {
		t.Fatalf("unexpected inline overhang: %#v", opts)
	}
	if opts.PageHeight != 100 || opts.AdjustPageWidth || opts.JustifyVertically {
		t.Fatalf("unexpected inline options: %#v", opts)
	}
}

func TestOptionsForFullscore(t *testing.T) {
	r := Request{Type: "fullscore", Font: "Leipzig", Unit: "6.875"}
	g := Geometry{LineWidthPX: 1000, TextHeightPX: 1400}
	opts := OptionsForRequest(r, g)
	if opts.PageWidth != 1100 || opts.PageHeight != 1400 {
		t.Fatalf("unexpected fullscore size: %#v", opts)
	}
	if !opts.JustifyVertically || opts.AdjustPageHeight || opts.Header != "" {
		t.Fatalf("unexpected fullscore options: %#v", opts)
	}
}

func TestOptionsForExample(t *testing.T) {
	r := Request{Type: "example", Font: "Leipzig", Unit: "6.875"}
	g := Geometry{LineWidthPX: 1000, TextHeightPX: 1400}
	opts := OptionsForRequest(r, g)
	if opts.Header != "none" || opts.Breaks != "encoded" || opts.PageWidth != 1100 {
		t.Fatalf("unexpected example options: %#v", opts)
	}
	if !opts.AdjustPageHeight || !opts.AdjustPageWidth {
		t.Fatalf("unexpected example options: %#v", opts)
	}
}

func TestVerovioArgsPassNotationOverhang(t *testing.T) {
	opts := OptionsForRequest(Request{Type: "inline", Font: "Leland", Unit: "8"}, Geometry{LineWidthPX: 1000})
	args := verovioArgs(opts)
	joined := strings.Join(args, " ")
	for _, flag := range []string{"--page-margin-left 50", "--page-margin-right 50", "--page-margin-top 0", "--page-margin-bottom 0"} {
		if !strings.Contains(joined, flag) {
			t.Fatalf("expected %q in %s", flag, joined)
		}
	}
}

func TestParseJobs(t *testing.T) {
	if _, err := ParseJobs("0"); err == nil {
		t.Fatal("expected invalid jobs error")
	}
	if got, err := ParseJobs("2"); err != nil || got != 2 {
		t.Fatalf("expected 2 jobs, got %d err %v", got, err)
	}
}
