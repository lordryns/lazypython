package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const cacheFileName = "pypi_packages_cache.json"
const cacheMaxAge = 24 * time.Hour

type PackageCache struct {
	Packages  []string  `json:"packages"`
	Timestamp time.Time `json:"timestamp"`
}

func getCacheFilePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	appCacheDir := filepath.Join(cacheDir, "lazypythoncli")
	if err := os.MkdirAll(appCacheDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(appCacheDir, cacheFileName), nil
}

func loadPackagesFromCache() ([]string, bool) {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}

	var cache PackageCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}

	if time.Since(cache.Timestamp) > cacheMaxAge {
		return nil, false
	}

	return cache.Packages, true
}

func savePackagesToCache(packages []string) error {
	cachePath, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cache := PackageCache{
		Packages:  packages,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// that is the question (lmao)
func toCacheOrNotToCache() bool {
	if packages, ok := loadPackagesFromCache(); ok {
		pythonPackages = packages
		return true
	}

	return false
}
