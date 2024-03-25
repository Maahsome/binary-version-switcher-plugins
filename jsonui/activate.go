package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-resty/resty/v2"
)

func activateVersion(ver string, binPath string, symLinkPath string) {

	path := filepath.Join(binPath, fmt.Sprintf("jsonui/%s/jsonui", ver))
	if !fileExists(path) {
		c.Logger.Infof("Downloading version %s of jsonui to path %s", ver, path)
		err := downloadArtifact(ver, path)
		if err != nil {
			c.Logger.WithError(err).Error("failed to download artifact")
		}
	}

	c.Logger.Infof("Activating version %s of jsonui", ver)
	err := changeFilePermissionsAndSymlink(path, symLinkPath)
	if err != nil {
		c.Logger.WithError(err).Error("failed to change perms and symlink")
	}

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DownloadArtifact downloads the jsonui binary for the specified SEMVER and saves it to the given PATH.
func downloadArtifact(semver, path string) error {
	targetOS := runtime.GOOS
	if targetOS != "linux" && targetOS != "darwin" {
		return fmt.Errorf("unsupported OS: %s", targetOS)
	}

	targetArch := runtime.GOARCH
	if targetArch == "amd64" {
		targetArch = "amd64"
	} else if targetArch == "arm64" {
		targetArch = "arm"
	} else {
		return fmt.Errorf("unsupported architecture: %s", targetArch)
	}

	if targetOS == "darwin" && targetArch == "arm" {
		return fmt.Errorf("unsupported architecture: %s", targetArch)
	}

	url := fmt.Sprintf("https://github.com/gulyasm/jsonui/releases/download/%s/jsonui_%s_%s", semver, targetOS, targetArch)

	client := resty.New()
	resp, err := client.R().SetOutput(path).Get(url)
	if err != nil {
		return fmt.Errorf("error downloading jsonui: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error downloading jsonui: status code %d", resp.StatusCode())
	}

	return nil
}

// changeFilePermissionsAndSymlink changes the permissions of the file at the given path to 0755
// and creates a symbolic link to it in /usr/local/bin.
func changeFilePermissionsAndSymlink(binPath string, symPath string) error {
	// Change the file permissions to 0755
	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("failed to change file permissions: %w", err)
	}

	// Get the filename from the path
	filename := filepath.Base(binPath)

	// Create a symbolic link in /usr/local/bin
	symlinkPath := filepath.Join(symPath, filename)
	if fileExists(symlinkPath) {
		rerr := os.Remove(symlinkPath)
		if rerr != nil {
			return fmt.Errorf("failed to remove symbolic link: %w", rerr)
		}
	}
	c.Logger.Infof("creating symlink %s -> %s", symlinkPath, binPath)
	if err := os.Symlink(binPath, symlinkPath); err != nil {
		return fmt.Errorf("failed to create symbolic link: %w", err)
	}

	return nil
}
