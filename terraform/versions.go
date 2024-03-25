package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
)

func listVersions(verMatch string, all bool) {

	releaseTags, err := getTerraformVersions()
	if err != nil {
		c.Logger.WithError(err).Error("failed to read terraform releases")
	}

	if all {
		for _, v := range releaseTags {
			if len(verMatch) > 0 {
				if strings.HasPrefix(v, verMatch) {
					fmt.Printf("%s\n", v)
				}
			} else {
				fmt.Printf("%s\n", v)
			}
		}
	} else {
		justMinors(&releaseTags, verMatch)
	}
}

func extractVersions(html []byte) []string {
	var versions []string
	re := regexp.MustCompile(`\/terraform\/([\w.-]+)\/`)
	matches := re.FindAllSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 1 {
			versions = append(versions, string(match[1]))
		}
	}
	return versions
}

func justMinors(releaseTags *[]string, verMatch string) {

	minorRelease := map[string]semver.Version{}
	for _, v := range *releaseTags {
		nv, err := semver.NewVersion(v)
		if err != nil {
			c.Logger.Error(fmt.Sprintf("Error parsing SemVer for %s", v))
		}
		verKey := fmt.Sprintf("%d.%d", nv.Major(), nv.Minor())
		processVer := true
		if len(verMatch) > 0 {
			if !strings.HasPrefix(v, verMatch) {
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
		fmt.Printf("%s\n", v)
	}
}

// GetGitHubReleases fetches all releases for the given owner and repo from GitHub.
// curl -L \
// -H "Accept: application/vnd.github+json" \
// -H "X-GitHub-Api-Version: 2022-11-28" \
// https://api.github.com/repos/kubernetes/kubernetes/releases
func getTerraformVersions() ([]string, error) {
	client := resty.New()
	url := "https://releases.hashicorp.com/terraform/"
	resp, err := client.R().
		Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	tagNames := extractVersions(resp.Body())
	return tagNames, nil
}
