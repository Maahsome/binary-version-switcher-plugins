package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
)

func listVersions(verMatch string, all bool) {

	releaseTags, err := getGitHubReleases("bronze1man", "yaml2json")
	if err != nil {
		c.Logger.WithError(err).Error("failed to read github releases")
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

	// bronze1man/yaml2json uses a "v", so we should indicate so
	for _, v := range vs {
		shortV, _ := strings.CutSuffix(v.String(), ".0")
		fmt.Printf("v%s\n", shortV)
	}
}

// GetGitHubReleases fetches all releases for the given owner and repo from GitHub.
// curl -L \
// -H "Accept: application/vnd.github+json" \
// -H "X-GitHub-Api-Version: 2022-11-28" \
// https://api.github.com/repos/kubernetes/kubernetes/releases
func getGitHubReleases(owner, repo string) ([]string, error) {
	client := resty.New()
	var allReleases GithubReleaseList
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d", owner, repo, page)
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("X-GitHub-Api-Version", "2022-11-28").
			SetHeader("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUB_TOKEN"))).
			Get(url)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		var releases GithubReleaseList
		if err := json.Unmarshal(resp.Body(), &releases); err != nil {
			return nil, err
		}

		if len(releases) == 0 {
			break
		}

		allReleases = append(allReleases, releases...)
		page++
	}

	var tagNames []string
	for _, release := range allReleases {
		tagNames = append(tagNames, release.TagName)
	}
	return tagNames, nil
}
