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
	if err := WriteInclude(path, images, "inline", "0.0pt"); err != nil {
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
	if err := WriteInclude(path, images, "example", "15.0pt"); err != nil {
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

func TestWriteIncludeIndent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nota-score-1.tex")
	images := []string{
		filepath.Join(dir, "nota-score-1-page-1.pdf"),
		filepath.Join(dir, "nota-score-1-page-2.pdf"),
	}
	if err := WriteInclude(path, images, "inline", "15.0pt"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	first := "\\hspace{15.0pt}\\makebox[\\dimexpr\\linewidth-15.0pt\\relax][c]{\\includegraphics[scale=1]{nota-score-1-page-1.pdf}}"
	if !strings.Contains(text, first) {
		t.Fatalf("expected indented first system, got %s", text)
	}
	second := "\\makebox[\\linewidth][c]{\\includegraphics[scale=1]{nota-score-1-page-2.pdf}}"
	if !strings.Contains(text, second) {
		t.Fatalf("expected full-width later system, got %s", text)
	}
}

func TestWriteFullIgnoreIndent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nota-score-1.tex")
	images := []string{filepath.Join(dir, "nota-score-1-page-1.pdf")}
	if err := WriteInclude(path, images, "full", "15.0pt"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "\\hspace") {
		t.Fatalf("full should not indent: %s", data)
	}
}
