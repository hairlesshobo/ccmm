package util

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"regexp"
)

func FileExists(path string) bool {
	// if an error occurred or its a directory, we throw up
	if stat, err := os.Stat(path); err != nil || stat.IsDir() {
		return false
	}

	return true
}

func requireMultipleFileOrDir(rootDir string, items []string, needsDir bool) bool {
	itemType := "file"
	if needsDir {
		itemType = "directory"
	}

	for _, checkPath := range items {
		slog.Debug(fmt.Sprintf("util.requireMultipleFileOrDir: Testing for %s '%s' in path '%s'", itemType, checkPath, rootDir))
		fullPath := path.Join(rootDir, checkPath)

		if stat, err := os.Stat(fullPath); err != nil || (needsDir && !stat.IsDir()) || (!needsDir && stat.IsDir()) {
			slog.Debug(fmt.Sprintf("util.requireMultipleFileOrDir: required %s missing: %s", itemType, checkPath))
			return false
		}
	}

	return true
}

func requireRegexFileOrDirMatch(rootDir string, namePattern string, needsDir bool) (bool, string) {
	itemType := "file"
	if needsDir {
		itemType = "directory"
	}

	entries, err := os.ReadDir(rootDir)

	if err != nil {
		slog.Error(fmt.Sprintf("util.requireRegexFileOrDirMatch: Error occurred when reading directory '%s': %s", rootDir, err))
		return false, ""
	}

	for _, entry := range entries {
		slog.Debug(fmt.Sprintf("util.requireRegexFileOrDirMatch: Testing for %s with pattern '%s' in path '%s'", itemType, namePattern, rootDir))
		match, _ := regexp.MatchString(namePattern, entry.Name())

		if match && ((needsDir && entry.IsDir()) || (!needsDir && !entry.IsDir())) {
			return true, path.Join(rootDir, entry.Name())
		}
	}

	return false, ""
}

func RequireDirs(rootDir string, dirs []string) bool {
	return requireMultipleFileOrDir(rootDir, dirs, true)
}

func RequireFiles(rootDir string, files []string) bool {
	return requireMultipleFileOrDir(rootDir, files, false)
}

func RequireRegexDirMatch(rootDir string, namePattern string) (bool, string) {
	return requireRegexFileOrDirMatch(rootDir, namePattern, true)
}

func RequireRegexFileMatch(rootDir string, namePattern string) (bool, string) {
	return requireRegexFileOrDirMatch(rootDir, namePattern, false)
}

func DirectoryExists(testDir string) bool {
	if stat, err := os.Stat(testDir); err != nil || !stat.IsDir() {
		return false
	}

	return true
}
