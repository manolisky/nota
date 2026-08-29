package nota

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func Run(cfg Config) error {
	if cfg.Backend == nil {
		return errors.New("missing render backend")
	}
	if cfg.Converter == nil {
		return errors.New("missing SVG converter")
	}
	jobs, err := ParseJobs(cfg.Jobs)
	if err != nil {
		return err
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(projectRoot, cfg.Jobname+".nota.aux")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read manifest %s: %w", filepath.Base(manifestPath), err)
	}
	requests, parseErrs := ParseManifest(string(data))
	if len(parseErrs) > 0 {
		var b strings.Builder
		for _, parseErr := range parseErrs {
			b.WriteString(parseErr.Error())
			b.WriteByte('\n')
		}
		return fmt.Errorf("malformed manifest:\n%s", strings.TrimSpace(b.String()))
	}
	if len(requests) == 0 {
		fmt.Printf("No Nota score requests in %s\n", filepath.Base(manifestPath))
		return nil
	}

	outputDir := filepath.Join(projectRoot, cfg.Jobname+"-nota")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	cachePath := filepath.Join(outputDir, "nota-cache.json")
	cache := LoadCache(cachePath)
	var cacheMu sync.Mutex

	type job struct {
		index int
		req   Request
	}
	type result struct {
		index int
		res   ScoreResult
		err   error
	}

	jobsCh := make(chan job)
	resultsCh := make(chan result, len(requests))
	var wg sync.WaitGroup

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobsCh {
				res, err := renderOne(projectRoot, outputDir, item.req, cfg, &cache, &cacheMu)
				resultsCh <- result{index: item.index, res: res, err: err}
			}
		}()
	}

	go func() {
		for i, req := range requests {
			jobsCh <- job{index: i, req: req}
		}
		close(jobsCh)
		wg.Wait()
		close(resultsCh)
	}()

	var results []result
	for res := range resultsCh {
		results = append(results, res)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].index < results[j].index })

	var failed []string
	for _, res := range results {
		if res.err != nil {
			failed = append(failed, fmt.Sprintf("score %s: %v", res.res.Request.ID, res.err))
			continue
		}
		if res.res.Cached {
			fmt.Printf("- Score %s cached\n", res.res.Request.ID)
		} else {
			fmt.Printf("- Score %s rendered (%d page SVGs)\n", res.res.Request.ID, len(res.res.PageSVGs))
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d score(s) failed:\n%s", len(failed), strings.Join(failed, "\n"))
	}
	return SaveCache(cachePath, cache)
}

func renderOne(projectRoot string, outputDir string, r Request, cfg Config, cache *Cache, cacheMu *sync.Mutex) (ScoreResult, error) {
	result := ScoreResult{Request: r}
	inputPath := filepath.Join(projectRoot, r.Path)
	if _, err := os.Stat(inputPath); err != nil {
		return result, fmt.Errorf("input file %s: %w", r.Path, err)
	}
	key, err := CacheKey(projectRoot, r)
	if err != nil {
		return result, err
	}
	baseName := "nota-score-" + r.ID
	texPath := filepath.Join(outputDir, baseName+".tex")
	result.TexPath = texPath

	cacheMu.Lock()
	cached := !cfg.Force && cache.Scores[r.ID] == key && fileExists(texPath)
	cacheMu.Unlock()
	if cached {
		result.Cached = true
		return result, nil
	}

	geometry, err := GeometryFromRequest(r)
	if err != nil {
		return result, err
	}
	options := OptionsForRequest(r, geometry)
	svgs, err := cfg.Backend.RenderSVGPages(inputPath, outputDir, baseName, options)
	if err != nil {
		return result, err
	}
	if len(svgs) == 0 {
		return result, errors.New("Verovio produced no SVG pages")
	}
	pdfs, err := cfg.Converter.ConvertSVGPages(svgs, outputDir, baseName)
	if err != nil {
		return result, err
	}
	if err := WriteInclude(texPath, pdfs, r.Type, r.Indent); err != nil {
		return result, err
	}

	cacheMu.Lock()
	cache.Scores[r.ID] = key
	cacheMu.Unlock()

	result.PageSVGs = svgs
	result.PagePDFs = pdfs
	return result, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func WriteInclude(path string, imagePaths []string, scoreType string, indent string) error {
	var b strings.Builder
	if scoreType == "example" {
		if len(imagePaths) > 0 {
			b.WriteString(fmt.Sprintf("\\makebox[\\linewidth][c]{\\includegraphics[scale=1]{%s}}\n", filepath.Base(imagePaths[0])))
		}
	} else {
		indentPX, _ := IndentPixels(indent)
		useIndent := scoreType == "inline" && indentPX > 0
		for i, imagePath := range imagePaths {
			base := filepath.Base(imagePath)
			if useIndent && i == 0 {
				b.WriteString(fmt.Sprintf("\\hspace{%s}\\makebox[\\dimexpr\\linewidth-%s\\relax][c]{\\includegraphics[scale=1]{%s}}\\par\n", indent, indent, base))
			} else {
				b.WriteString(fmt.Sprintf("\\makebox[\\linewidth][c]{\\includegraphics[scale=1]{%s}}\\par\n", base))
			}
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
