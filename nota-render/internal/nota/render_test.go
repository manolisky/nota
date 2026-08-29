package nota

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteInclude(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nota-score-1.tex")
	images := []string{
		filepath.Join(dir, "nota-score-1-page-1.pdf"),
		filepath.Join(dir, "nota-score-1-page-2.pdf"),
	}
	if err := WriteInclude(path, images, "inline"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	want := "\\makebox[\\linewidth][c]{\\includegraphics[scale=1]{nota-score-1-page-2.pdf}}"
	if !strings.Contains(text, want) {
		t.Fatalf("expected %q in %s", want, text)
	}
	if strings.Contains(text, "\\textwidth") {
		t.Fatalf("include should use \\linewidth, got %s", text)
	}
}

func TestWriteExampleInclude(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nota-score-1.tex")
	images := []string{
		filepath.Join(dir, "nota-score-1-page-1.pdf"),
		filepath.Join(dir, "nota-score-1-page-2.pdf"),
		filepath.Join(dir, "nota-score-1-page-3.pdf"),
	}
	if err := WriteInclude(path, images, "example"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "page=2") {
		t.Fatalf("example include should not write page selectors: %s", text)
	}
	if !strings.Contains(text, "\\makebox[\\linewidth][c]") {
		t.Fatalf("example include should use \\linewidth: %s", text)
	}
	if strings.Count(text, "includegraphics") != 1 {
		t.Fatalf("example include should emit only the first page: %s", text)
	}
}
