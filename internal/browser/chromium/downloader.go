package chromium

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	versionsURL     = "https://googlechromelabs.github.io/chrome-for-testing/known-good-versions-with-downloads.json"
	DefaultRevision = "1321438"
)

// DefaultRevisionForTest exposes DefaultRevision for tests.
func DefaultRevisionForTest() string { return DefaultRevision }

type versionEntry struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
	Downloads struct {
		Chrome []struct {
			Platform string `json:"platform"`
			URL      string `json:"url"`
		} `json:"chrome"`
	} `json:"downloads"`
}

type versionsResp struct {
	Versions []versionEntry `json:"versions"`
}

func ResolveDownloadURL(revision string) (string, error) {
	resp, err := http.Get(versionsURL)
	if err != nil {
		return "", fmt.Errorf("fetch versions: %w", err)
	}
	defer resp.Body.Close()
	var data versionsResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	platform := Platform()
	for _, v := range data.Versions {
		if v.Revision == revision || v.Version == revision {
			for _, dl := range v.Downloads.Chrome {
				if dl.Platform == platform {
					return dl.URL, nil
				}
			}
		}
	}
	return "", fmt.Errorf("revision %q not found for platform %q", revision, platform)
}

func Download(revision, destDir string) (string, error) {
	url, err := ResolveDownloadURL(revision)
	if err != nil {
		return "", err
	}
	revDir := filepath.Join(destDir, revision)
	if err := os.MkdirAll(revDir, 0755); err != nil {
		return "", err
	}
	fmt.Printf("Downloading Chromium %s for %s...\n", revision, Platform())
	tmpFile := filepath.Join(revDir, "chromium.zip")
	if err := downloadFile(url, tmpFile); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)
	fmt.Println("Extracting...")
	if err := extractZip(tmpFile, revDir); err != nil {
		return "", err
	}
	execPath := findExecutable(revDir)
	fmt.Printf("Chromium installed to %s\n", execPath)
	return execPath, nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	clean := filepath.Clean(dest) + string(os.PathSeparator)
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(path, clean) {
			continue // zip slip protection
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
			continue
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}
		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func findExecutable(dir string) string {
	name := ExecutableName()
	var found string
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
