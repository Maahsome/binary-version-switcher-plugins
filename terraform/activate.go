package main

import (
	"archive/zip"
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

	path := filepath.Join(binPath, fmt.Sprintf("terraform/%s/terraform", ver))
	if !fileExists(path) {
		c.Logger.Infof("Downloading version %s of terraform to path %s", ver, path)
		err := downloadArtifact(ver, path)
		if err != nil {
			c.Logger.WithError(err).Error("failed to download artifact")
		}
	}

	c.Logger.Infof("Activating version %s of terraform", ver)
	err := changeFilePermissionsAndSymlink(path, symLinkPath)
	if err != nil {
		c.Logger.WithError(err).Error("failed to change perms and symlink")
	}

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DownloadArtifact downloads the terraform binary for the specified SEMVER and saves it to the given PATH.
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

	url := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", semver, semver, targetOS, targetArch)

	client := resty.New()
	resp, err := client.R().SetOutput(fmt.Sprintf("%s.zip", path)).Get(url)
	if err != nil {
		return fmt.Errorf("error downloading terraform: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error downloading terraform: status code %d", resp.StatusCode())
	}

	// extract
	dlDir := filepath.Dir(path)
	uerr := unzip(fmt.Sprintf("%s.zip", path), dlDir)
	if uerr != nil {
		return fmt.Errorf("failed to extract the terraform archive: %w", uerr)
	}
	rerr := os.Remove(fmt.Sprintf("%s.zip", path))
	if rerr != nil {
		return fmt.Errorf("failed to remove archive file: %w", rerr)
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

// unzip extracts a zip file specified by the src parameter to the dest directory.
func unzip(src, dest string) error {
	// Open the zip file specified by the src parameter.
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Iterate through the files in the archive.
	for _, f := range r.File {
		// Construct the path to the file within the destination directory.
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip (a security vulnerability that allows an attacker to write files outside the destination directory).
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		// Create the directory structure for the file.
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create the directory structure for the file if it doesn't already exist.
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Open the file within the zip archive.
		inFile, err := f.Open()
		if err != nil {
			return err
		}

		// Create the file in the destination directory.
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			inFile.Close()
			return err
		}

		// Copy the contents of the file from the archive to the destination file.
		_, err = io.Copy(outFile, inFile)

		// Close the files.
		inFile.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
