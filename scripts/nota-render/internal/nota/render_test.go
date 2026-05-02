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
	want := "\\includegraphics[scale=1]{nota-score-1-page-2.pdf}"
	if !strings.Contains(string(data), want) {
		t.Fatalf("expected %q in %s", want, string(data))
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
	if strings.Contains(string(data), "page=2") {
		t.Fatalf("example include should not write page selectors: %s", string(data))
	}
}
