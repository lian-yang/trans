package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	repoOwner    = "lian-yang"
	repoName     = "trans"
	releaseAPI   = "https://api.github.com/repos/%s/%s/releases/latest"
	downloadURL  = "https://github.com/%s/%s/releases/download/%s/trans-%s-%s.tar.gz"
)

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
}

// CheckLatestVersion queries the GitHub API for the latest release tag.
func CheckLatestVersion() (string, error) {
	url := fmt.Sprintf(releaseAPI, repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}
	return rel.TagName, nil
}

// IsNewer returns true if latest semver is greater than current.
func IsNewer(current, latest string) bool {
	cur := strings.TrimPrefix(current, "v")
	lat := strings.TrimPrefix(latest, "v")

	p1 := strings.Split(cur, ".")
	p2 := strings.Split(lat, ".")

	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}

	for i := 0; i < maxLen; i++ {
		var v1, v2 int
		if i < len(p1) {
			v1, _ = strconv.Atoi(p1[i])
		}
		if i < len(p2) {
			v2, _ = strconv.Atoi(p2[i])
		}
		if v2 > v1 {
			return true
		}
		if v2 < v1 {
			return false
		}
	}
	return false
}

// Update checks for a newer release and replaces the current binary.
func Update(currentVersion string) error {
	// Detect Homebrew-managed installation.
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	if isHomebrew(exePath) {
		fmt.Println("Homebrew-managed installation detected.")
		fmt.Println("Run: brew upgrade trans")
		return nil
	}

	latest, err := CheckLatestVersion()
	if err != nil {
		return err
	}

	if !IsNewer(currentVersion, latest) {
		fmt.Printf("Already up to date: %s\n", currentVersion)
		return nil
	}

	fmt.Printf("Updating %s → %s ...\n", currentVersion, latest)

	url := fmt.Sprintf(downloadURL, repoOwner, repoName, latest, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	binary, err := extractBinary(resp.Body)
	if err != nil {
		return err
	}

	// Atomic replace: write to temp → backup old → rename new → cleanup.
	tmpPath := exePath + ".new"
	if err := os.WriteFile(tmpPath, binary, 0755); err != nil {
		return fmt.Errorf("failed to write update: %w", err)
	}

	backupPath := exePath + ".old"
	_ = os.Remove(backupPath)
	if err := os.Rename(exePath, backupPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to backup binary: %w", err)
	}

	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Rename(backupPath, exePath)
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	os.Remove(backupPath)

	fmt.Printf("Updated to %s\n", latest)
	return nil
}

func isHomebrew(exePath string) bool {
	return strings.Contains(exePath, "homebrew") ||
		strings.Contains(exePath, "Cellar")
}

func extractBinary(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read archive: %w", err)
		}
		if !hdr.FileInfo().IsDir() {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("no binary found in archive")
}
