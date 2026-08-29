package nota

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Cache struct {
	Version string            `json:"version"`
	Scores  map[string]string `json:"scores"`
}

func LoadCache(path string) Cache {
	data, err := os.ReadFile(path)
	if err != nil {
		return Cache{Version: Version, Scores: map[string]string{}}
	}
	var cache Cache
	if json.Unmarshal(data, &cache) != nil || cache.Scores == nil {
		return Cache{Version: Version, Scores: map[string]string{}}
	}
	return cache
}

func SaveCache(path string, cache Cache) error {
	cache.Version = Version
	if cache.Scores == nil {
		cache.Scores = map[string]string{}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
