package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

func getFiles(searchStart string, searchPattern string) (map[string]string, error) {

	searchPattern = filepath.ToSlash(searchPattern)

	if !doublestar.ValidatePathPattern(searchPattern) {
		//GlobalLogger.Panic("Invalid search pattern", zap.String("pattern", pathn))
		return nil, fmt.Errorf("invalid search pattern %s", searchPattern)
	}

	mid, pattern := doublestar.SplitPattern(searchPattern)

	combinedSearchPath := filepath.Join(searchStart, mid)
	searchFS := os.DirFS(combinedSearchPath)

	foundPaths, err := doublestar.Glob(searchFS, pattern)

	if err != nil {
		return nil, err
	}

	files := make(map[string]string, len(foundPaths))
	for _, path := range foundPaths {
		archivePath := filepath.Clean(filepath.Join(mid, path))
		abspath, err := filepath.Abs(filepath.Join(searchStart, mid, path))
		if err != nil {
			return nil, errors.Join(errors.New("failed to get absolute path for "+abspath), err)
		}
		files[archivePath] = abspath
	}

	return files, nil
}
