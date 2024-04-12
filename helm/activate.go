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

	path := filepath.Join(binPath, fmt.Sprintf("helm/%s/helm", ver))
	if !fileExists(path) {
		c.Logger.Infof("Downloading version %s of helm to path %s", ver, path)
		err := downloadArtifact(ver, path)
		if err != nil {
			c.Logger.WithError(err).Fatal("failed to download artifact")
		}
	}

	c.Logger.Infof("Activating version %s of helm", ver)
	err := changeFilePermissionsAndSymlink(path, symLinkPath)
	if err != nil {
		c.Logger.WithError(err).Error("failed to change perms and symlink")
	}

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DownloadArtifact downloads the helm binary for the specified SEMVER and saves it to the given PATH.
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

	// https://get.helm.sh/helm-v${HELM_SEMVER}-${TARGET_OS}-amd64.tar.gz
	url := fmt.Sprintf("https://get.helm.sh/helm-%s-%s-%s.tar.gz", semver, targetOS, targetArch)

	c.Logger.Debugf("downloading helm from %s", url)
	client := resty.New()
	resp, err := client.R().SetOutput(fmt.Sprintf("%s.tar.gz", path)).Get(url)
	if err != nil {
		return fmt.Errorf("error downloading helm: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error downloading helm: status code %d", resp.StatusCode())
	}

	// extract archive
	if err := extractTarGz(fmt.Sprintf("%s.tar.gz", path), filepath.Dir(path), targetOS, targetArch); err != nil {
		return fmt.Errorf("error extracting helm: %v", err)
	}

	return nil
}

// extractTarGz extracts files from the "teleport" directory inside a tar.gz file to a destination directory.
func extractTarGz(tarGzPath, destDir string, targetOS string, targetArch string) error {
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

		// Check if the file is inside the "teleport" directory
		if strings.HasPrefix(header.Name, fmt.Sprintf("%s-%s", targetOS, targetArch)) {
			// Construct the full path for the file, excluding the "teleport/" prefix
			filePath := filepath.Join(destDir, strings.TrimPrefix(header.Name, fmt.Sprintf("%s-%s/", targetOS, targetArch)))

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
