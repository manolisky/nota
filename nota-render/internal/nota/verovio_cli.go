package nota

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type VerovioCLIBackend struct {
	VerovioPath string
	MScorePath  string
}

func NewVerovioCLIBackend(verovioPath string, mscorePath string) VerovioCLIBackend {
	return VerovioCLIBackend{VerovioPath: verovioPath, MScorePath: mscorePath}
}

func (b VerovioCLIBackend) RenderSVGPages(inputPath string, outputDir string, baseName string, options VerovioOptions) ([]string, error) {
	renderInput := inputPath
	cleanup := func() {}
	if strings.EqualFold(filepath.Ext(inputPath), ".mscz") {
		converted, clean, err := b.convertMuseScore(inputPath)
		if err != nil {
			return nil, err
		}
		renderInput = converted
		cleanup = clean
	}
	defer cleanup()

	var outputs []string
	for page := 1; ; page++ {
		outPath := filepath.Join(outputDir, fmt.Sprintf("%s-page-%d.svg", baseName, page))
		pageOpts := options
		if page == 1 && options.FirstPageWidth > 0 {
			pageOpts.PageWidth = options.FirstPageWidth
		}
		err := b.runVerovio(renderInput, outPath, page, pageOpts)
		if err != nil {
			if page > 1 && isPageRangeError(err) {
				break
			}
			return nil, err
		}
		outputs = append(outputs, outPath)
	}
	return outputs, nil
}

func (b VerovioCLIBackend) convertMuseScore(inputPath string) (string, func(), error) {
	command, err := lookCommand(b.MScorePath)
	if err != nil {
		return "", nil, fmt.Errorf(".mscz input requires MuseScore CLI %q on PATH", b.MScorePath)
	}
	tmp, err := os.CreateTemp("", "nota-*.mxl")
	if err != nil {
		return "", nil, err
	}
	tmp.Close()
	cmd := exec.Command(command, inputPath, "-o", tmp.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("MuseScore conversion failed: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}

func (b VerovioCLIBackend) runVerovio(inputPath string, outputPath string, page int, options VerovioOptions) error {
	command, err := lookCommand(b.VerovioPath)
	if err != nil {
		return fmt.Errorf("Verovio CLI %q not found on PATH", b.VerovioPath)
	}
	args := []string{"-t", "svg", "-p", strconv.Itoa(page), "-o", outputPath}
	args = append(args, verovioArgs(options)...)
	args = append(args, inputPath)

	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("verovio page %d failed: %v: %s", page, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func verovioArgs(options VerovioOptions) []string {
	args := []string{
		"--font", options.Font,
		"--unit", options.Unit,
		"--footer", options.Footer,
	}
	if options.Header != "" {
		args = append(args, "--header", options.Header)
	}
	if options.Breaks != "" {
		args = append(args, "--breaks", options.Breaks)
	}
	if options.MMOutput {
		args = append(args, "--mm-output")
	}
	if options.XMLIDChecksum {
		args = append(args, "--xml-id-checksum")
	}
	if options.RemoveIDs {
		args = append(args, "--remove-ids")
	}
	if options.OutputSMuFLEntities {
		args = append(args, "--output-smufl-xml-entities")
	}
	args = appendIntArg(args, "--page-width", options.PageWidth)
	args = appendIntArg(args, "--page-height", options.PageHeight)
	// Always pass margins so 0 is explicit (Verovio defaults to 50 if omitted).
	args = append(args, "--page-margin-top", strconv.Itoa(options.PageMarginTop))
	args = append(args, "--page-margin-bottom", strconv.Itoa(options.PageMarginBottom))
	args = append(args, "--page-margin-left", strconv.Itoa(options.PageMarginLeft))
	args = append(args, "--page-margin-right", strconv.Itoa(options.PageMarginRight))
	if options.AdjustPageHeight {
		args = append(args, "--adjust-page-height")
	}
	if options.AdjustPageWidth {
		args = append(args, "--adjust-page-width")
	}
	if options.JustifyVertically {
		args = append(args, "--justify-vertically")
	}
	return args
}

func appendIntArg(args []string, key string, value int) []string {
	if value == 0 {
		return args
	}
	return append(args, key, strconv.Itoa(value))
}

var pageRangeRE = regexp.MustCompile(`page requested .* is not in the page range`)

func isPageRangeError(err error) bool {
	return pageRangeRE.MatchString(err.Error())
}
