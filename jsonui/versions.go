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

	releaseTags, err := getGitHubReleases("gulyasm", "jsonui")
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

	// kubernetes/kubernetes uses a "v", so we should indicate so
	for _, v := range vs {
		fmt.Printf("v%s\n", v)
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

// #!/bin/bash
// # gh release list --repo gulyasm/jsonui

// set -o pipefail

// VERSION=$(gh release list --repo gulyasm/jsonui --limit 1 | cut -f3)

// readonly program="$(basename "$0")"
// verbose=0

// usage() {
//   echo "
//     This script downloads the specified operator-sdk version and symlinks it into /usr/local/bin/
//     usage: $program
//     options:
//       -l, --list                    List available versions
//       -v, --version                 Specify the version to download
//       -h, --help                    Show this help.
//   " | sed -E 's/^ {4}//'
// }

// versionlist() {
//     echo "VERSIONS"
//     gh release list --repo gulyasm/jsonui --limit 20
// }

// while getopts ":v:hl" opt; do
//    case $opt in
//       v|--version)
//          VERSION="${OPTARG}"
//          ;;
//       l|--list)
//          versionlist
// 	 exit 0
//          ;;
//       h|--help)
// 	 usage
//          exit 1
//          ;;
//    esac
// done
// shift $((OPTIND-1))

// if [[ -z ${VERSION} ]]; then
// 	usage
// 	exit 1
// fi

// # MacOS = 'darwin', Linux = 'linux', Windows = 'windows'
// if [[ "$OSTYPE" == "linux-gnu" ]]; then
//   TARGET_OS=linux
// elif [[ "$OSTYPE" == "darwin"* ]]; then
//   TARGET_OS=darwin
// fi

// echo "Installing/Activating ${VERSION}..."

// if [ ! -d "/usr/local/jsonui" ]; then
//   MY_LOGON=$(whoami)
//   echo ${MY_LOGON}
//   sudo mkdir -p /usr/local/jsonui
//   sudo chown ${MY_LOGON}:admin /usr/local/jsonui
// fi;

// if [ ! -d "/usr/local/jsonui/${VERSION}" ]; then
//   mkdir -p "/usr/local/jsonui/${VERSION}"

//   # https://github.com/gulyasm/jsonui/releases/download/v1.0.1/jsonui_linux_amd64
//   # https://github.com/gulyasm/jsonui/releases/download/v1.0.1/jsonui_darwin_amd64

//   url="https://github.com/gulyasm/jsonui/releases/download/${VERSION}/jsonui_${TARGET_OS}_amd64"
//   LINK_TARGET="/usr/local/jsonui/${VERSION}/jsonui_${TARGET_OS}_amd64"
//   curl -Ls ${url} -o ${LINK_TARGET}

// fi

// if [ -f "${LINK_TARGET}" ]; then
//   chmod 775 "${LINK_TARGET}"
// fi
// if [ -L "/usr/local/bin/jsonui" ] || [ -f "/usr/local/bin/jsonui" ]; then
//   rm "/usr/local/bin/jsonui"
// fi
// ln -s "${LINK_TARGET}" "/usr/local/bin/jsonui"
