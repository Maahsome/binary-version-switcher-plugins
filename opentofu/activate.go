package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-resty/resty/v2"
)

func activateVersion(ver string, binPath string, symLinkPath string) {

	path := filepath.Join(binPath, fmt.Sprintf("opentofu/%s/tofu", ver))
	if !fileExists(path) {
		c.Logger.Infof("Downloading version %s of opentofu to path %s", ver, path)
		err := downloadArtifact(ver, path)
		if err != nil {
			c.Logger.WithError(err).Error("failed to download artifact")
		}
	}

	c.Logger.Infof("Activating version %s of opentofu", ver)
	err := changeFilePermissionsAndSymlink(path, symLinkPath)
	if err != nil {
		c.Logger.WithError(err).Error("failed to change perms and symlink")
	}

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DownloadArtifact downloads the opentofu binary for the specified SEMVER and saves it to the given PATH.
func downloadArtifact(semver, path string) error {
	targetOS := runtime.GOOS
	if targetOS != "linux" && targetOS != "darwin" {
		return fmt.Errorf("unsupported OS: %s", targetOS)
	}

	targetArch := runtime.GOARCH
	if targetArch == "amd64" {
		targetArch = "amd64"
	} else if targetArch == "arm64" {
		targetArch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", targetArch)
	}

	if targetOS == "darwin" && targetArch == "arm" {
		return fmt.Errorf("unsupported architecture: %s", targetArch)
	}

	// https://github.com/opentofu/opentofu/releases/download/v1.7.1/tofu_1.7.1_darwin_amd64.zip

	// https://github.com/opentofu/opentofu/releases/download/v1.7.1/tofu_1.7.1_darwin_amd64.tar.gz
	// https://github.com/opentofu/opentofu/releases/download/v1.7.1/tofu_1.7.1_linux_amd64.tar.gz
	// https://github.com/opentofu/opentofu/releases/download/v1.7.1/tofu_1.7.1_darwin_arm64.tar.gz
	// https://github.com/opentofu/opentofu/releases/download/v1.7.1/tofu_1.7.1_linux_arm64.tar.gz

	realsemver := strings.TrimPrefix(semver, "v")
	url := fmt.Sprintf("https://github.com/opentofu/opentofu/releases/download/%s/tofu_%s_%s_%s.tar.gz", semver, realsemver, targetOS, targetArch)
	c.Logger.Debugf("downloading %s", url)

	client := resty.New()
	resp, err := client.R().SetOutput(fmt.Sprintf("%s.tar.gz", path)).Get(url)
	if err != nil {
		return fmt.Errorf("error downloading opentofu: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error downloading opentofu: status code %d", resp.StatusCode())
	}

	// extract archive
	if err := extractTarGz(fmt.Sprintf("%s.tar.gz", path), filepath.Dir(path)); err != nil {
		return fmt.Errorf("error extracting helm: %v", err)
	}

	return nil
}

// extractTarGz extracts files from the archive
func extractTarGz(tarGzPath, destDir string) error {
	// Open the tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Iterate through the files in the tar archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		filePath := filepath.Join(destDir, header.Name)

		// extract all the files
		// Check the type of the file
		switch header.Typeflag {
		case tar.TypeDir:
			// Create the directory
			if err := os.MkdirAll(filePath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create the file
			outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Copy the file data from the tar archive
			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported file type: %v", header.Typeflag)
		}
	}
	derr := os.Remove(tarGzPath)
	if derr != nil {
		c.Logger.Error("failed to remove archive file")
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
