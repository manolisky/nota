package nota

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type SVGConverter interface {
	ConvertSVGPages(svgPaths []string, outputDir string, baseName string) ([]string, error)
}

type RSVGConverter struct {
	Command string
}

func NewRSVGConverter(command string) RSVGConverter {
	return RSVGConverter{Command: command}
}

func (c RSVGConverter) ConvertSVGPages(svgPaths []string, outputDir string, baseName string) ([]string, error) {
	command, err := lookCommand(c.Command)
	if err != nil {
		return nil, fmt.Errorf("SVG-to-PDF conversion requires %q on PATH", c.Command)
	}
	var outputs []string
	for i, svgPath := range svgPaths {
		outPath := filepath.Join(outputDir, fmt.Sprintf("%s-page-%d.pdf", baseName, i+1))
		cmd := exec.Command(command, "-f", "pdf", "-o", outPath, svgPath)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("convert %s: %v: %s", svgPath, err, strings.TrimSpace(stderr.String()))
		}
		outputs = append(outputs, outPath)
	}
	return outputs, nil
}
