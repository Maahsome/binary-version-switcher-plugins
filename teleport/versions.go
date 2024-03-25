package main

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
)

func listVersions(verMatch string, all bool) {

	releaseTags, err := getTeleportDownloads()
	if err != nil {
		c.Logger.WithError(err).Error("failed to read github releases")
	}

	if all {
		for _, v := range releaseTags {
			if len(verMatch) > 0 {
				if strings.HasPrefix(v.ReleaseTag, verMatch) {
					fmt.Printf("%s\n", v.ReleaseTag)
				}
			} else {
				fmt.Printf("%s\n", v.ReleaseTag)
			}
		}
	} else {
		justMinors(releaseTags, verMatch)
	}
}

func justMinors(releaseTags map[string]ReleaseDownload, verMatch string) {

	minorRelease := map[string]semver.Version{}
	for _, v := range releaseTags {
		nv, err := semver.NewVersion(v.ReleaseTag)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Error parsing SemVer for %s", v))
		}
		verKey := fmt.Sprintf("%d.%d", nv.Major(), nv.Minor())
		processVer := true
		if len(verMatch) > 0 {
			if !strings.HasPrefix(v.ReleaseTag, verMatch) {
				processVer = false
			}
		}
		if processVer {
			if _, ok := minorRelease[verKey]; ok {
				if minorRelease[verKey].Patch() < nv.Patch() {
					minorRelease[verKey] = *nv
				}
			} else {
				// add
				minorRelease[verKey] = *nv
			}
		}
	}
	vs := []*semver.Version{}
	for _, v := range minorRelease {
		vs = append(vs, &v)
	}

	sort.Sort(semver.Collection(vs))

	// kubernetes/kubernetes uses a "v", so we should indicate so
	for _, v := range vs {
		fmt.Printf("%s\n", releaseTags[v.String()].ReleaseTag)
	}
}

func getTeleportDownloads() (map[string]ReleaseDownload, error) {

	targetOS := runtime.GOOS
	if targetOS != "linux" && targetOS != "darwin" {
		c.Logger.Fatal("failed to determine target os")
	}

	targetArch := runtime.GOARCH
	if targetArch == "amd64" {
		targetArch = "amd64"
	} else if targetArch == "arm64" {
		targetArch = "arm64"
	} else {
		c.Logger.Fatal("failed to determine target architecture")
	}

	client := resty.New()
	downloadURL := "https://goteleport.com/_next/data/latest/download.json"

	response, err := client.R().Get(downloadURL)
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		return map[string]ReleaseDownload{}, err
	}

	body := string(response.Body())
	// Unescape HTML entities
	unescapedContent := html.UnescapeString(body)

	// Regular expression to find the buildId value
	re := regexp.MustCompile(`"buildId":"([^"]+)"`)

	// Find the buildId value
	buildId := ""
	matches := re.FindStringSubmatch(unescapedContent)
	if len(matches) > 1 {
		buildId = matches[1]
	} else {
		return map[string]ReleaseDownload{}, fmt.Errorf("buildId not found")
	}

	url := fmt.Sprintf("https://goteleport.com/_next/data/%s/download.json", buildId)

	resp, err := client.R().Get(url)
	if err != nil {
		fmt.Printf("Error downloading download.json: %v\n", err)
		return map[string]ReleaseDownload{}, err
	}

	if resp.StatusCode() != 200 {
		return map[string]ReleaseDownload{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	var releases Downloads
	if err := json.Unmarshal(resp.Body(), &releases); err != nil {
		return map[string]ReleaseDownload{}, err
	}

	releaseInfo := map[string]ReleaseDownload{}
	for _, release := range releases.PageProps.InitialDownloads {
		for _, version := range release.Versions {
			for _, asset := range version.Assets {
				if asset.OS == targetOS && asset.Arch == targetArch {
					if strings.HasSuffix(asset.Name, ".tar.gz") {
						releaseInfo[version.Version] = ReleaseDownload{
							ReleaseTag: version.Version,
							Download:   asset.PublicURL,
						}
					}
				}
			}
		}
	}
	return releaseInfo, nil
}
