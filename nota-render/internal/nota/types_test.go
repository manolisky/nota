package nota

import "testing"

func TestParseManifest(t *testing.T) {
	data := "1|inline|score.mei|Leipzig|6.875|597.5pt|845pt|10pt|20pt|30pt|40pt\n"
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
	if got != 352 {
		t.Fatalf("expected 352, got %d", got)
	}
}

func TestOptionsForInline(t *testing.T) {
	r := Request{Type: "inline", Font: "Leland", Unit: "8"}
	g := Geometry{TextWidthPX: 1000}
	opts := OptionsForRequest(r, g)
	if opts.Header != "none" || !opts.AdjustPageHeight || opts.PageWidth != 1100 {
		t.Fatalf("unexpected inline options: %#v", opts)
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
