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

	releaseTags, err := getTeleportDownloads()
	if err != nil {
		c.Logger.WithError(err).Error("failed to read github releases")
	}

	if _, ok := releaseTags[ver]; !ok {
		c.Logger.Fatalf("version %s not found, cannot download and activate", ver)
	}

	path := filepath.Join(binPath, fmt.Sprintf("teleport/%s/teleport", ver))
	if !fileExists(path) {
		c.Logger.Infof("Downloading version %s of teleport to path %s", ver, path)
		err := downloadArtifact(path, releaseTags[ver].Download)
		if err != nil {
			c.Logger.WithError(err).Error("failed to download artifact")
		}
	}

	c.Logger.Infof("Activating version %s of teleport", ver)
	err = changeFilePermissionsAndSymlink(path, symLinkPath)
	if err != nil {
		c.Logger.WithError(err).Error("failed to change perms and symlink")
	}

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DownloadArtifact downloads the teleport binary for the specified SEMVER and saves it to the given PATH.
func downloadArtifact(path string, url string) error {
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

	// url := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/teleport", semver, targetOS, targetArch)

	client := resty.New()
	resp, err := client.R().SetOutput(fmt.Sprintf("%s.tar.gz", path)).Get(url)
	if err != nil {
		return fmt.Errorf("error downloading teleport: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error downloading teleport: status code %d", resp.StatusCode())
	}

	// extract archive
	if err := extractTarGz(fmt.Sprintf("%s.tar.gz", path), filepath.Dir(path)); err != nil {
		return fmt.Errorf("error extracting teleport: %v", err)
	}

	return nil
}

// extractTarGz extracts files from the "teleport" directory inside a tar.gz file to a destination directory.
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

		// Check if the file is inside the "teleport" directory
		if strings.HasPrefix(header.Name, "teleport/t") {
			// Construct the full path for the file, excluding the "teleport/" prefix
			filePath := filepath.Join(destDir, strings.TrimPrefix(header.Name, "teleport/"))

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

	// Loop through the file names and change the permissions
	files := []string{"tbot", "tctl", "teleport", "tsh"}
	dirPath := filepath.Dir(binPath)
	for _, file := range files {
		// Construct the path to the file
		fullPath := filepath.Join(dirPath, file)
		// Change the file permissions to 0755
		if err := os.Chmod(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to change file permissions: %w", err)
		}

		// Create a symbolic link in /usr/local/bin
		symlinkPath := filepath.Join(symPath, file)
		if fileExists(symlinkPath) {
			rerr := os.Remove(symlinkPath)
			if rerr != nil {
				return fmt.Errorf("failed to remove symbolic link: %w", rerr)
			}
		}
		c.Logger.Infof("creating symlink %s -> %s", symlinkPath, fullPath)
		if err := os.Symlink(fullPath, symlinkPath); err != nil {
			return fmt.Errorf("failed to create symbolic link: %w", err)
		}
	}
	return nil
}

// #!/bin/bash

//   # download and extract
//   DL_URL=$(echo ${TELEPORT_RELEASE_DATA} | jq -r ".pageProps.initialDownloads[] | select(.majorVersion==\"${MAJOR}\") | .versions[] | select(.version==\"${TELEPORT}\") | .assets[] | select(.name==\"teleport-v${TELEPORT}-${TARGET_OS}-${TARGET_ARCH}-bin.tar.gz\") | .publicUrl")
//   SHA=$(echo ${TELEPORT_RELEASE_DATA} | jq -r ".pageProps.initialDownloads[] | select(.majorVersion==\"${MAJOR}\") | .versions[] | select(.version==\"${TELEPORT}\") | .assets[] | select(.name==\"teleport-v${TELEPORT}-${TARGET_OS}-${TARGET_ARCH}-bin.tar.gz\") | .sha256")
//   echo "Downloading ${DL_URL}"
//   curl -Ls ${DL_URL} -o "/usr/local/teleport/${TELEPORT}/teleport_${TARGET_OS}_${TARGET_ARCH}.tar.gz"
//   tar -xzvf "/usr/local/teleport/${TELEPORT}/teleport_${TARGET_OS}_${TARGET_ARCH}.tar.gz" -C "/usr/local/teleport/${TELEPORT}/" --strip-components=1 teleport/tbot teleport/tctl teleport/teleport teleport/tsh
//   rm "/usr/local/teleport/${TELEPORT}/teleport_${TARGET_OS}_${TARGET_ARCH}.tar.gz"

//   if [ -f "/usr/local/teleport/${TELEPORT}/tbot" ]; then
//     chmod 775 "/usr/local/teleport/${TELEPORT}/tbot"
//   fi
//   if [ -f "/usr/local/teleport/${TELEPORT}/tctl" ]; then
//     chmod 775 "/usr/local/teleport/${TELEPORT}/tctl"
//   fi
//   if [ -f "/usr/local/teleport/${TELEPORT}/teleport" ]; then
//     chmod 775 "/usr/local/teleport/${TELEPORT}/teleport"
//   fi
//   if [ -f "/usr/local/teleport/${TELEPORT}/tsh" ]; then
//     chmod 775 "/usr/local/teleport/${TELEPORT}/tsh"
//   fi
// fi

// # symlink it all
// if [ -h "/usr/local/bin/tbot" ]; then
//   rm "/usr/local/bin/tbot"
// fi
// if [ -h "/usr/local/bin/tctl" ]; then
//   rm "/usr/local/bin/tctl"
// fi
// if [ -h "/usr/local/bin/teleport" ]; then
//   rm "/usr/local/bin/teleport"
// fi
// if [ -h "/usr/local/bin/tsh" ]; then
//   rm "/usr/local/bin/tsh"
// fi
// ln -s "/usr/local/teleport/${TELEPORT}/tbot" "/usr/local/bin/tbot"
// ln -s "/usr/local/teleport/${TELEPORT}/tctl" "/usr/local/bin/tctl"
// ln -s "/usr/local/teleport/${TELEPORT}/teleport" "/usr/local/bin/teleport"
// ln -s "/usr/local/teleport/${TELEPORT}/tsh" "/usr/local/bin/tsh"

// # verify versions
// tbot version
// tctl version
// teleport version
// tsh version
