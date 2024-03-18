package main

import (
	"fmt"

	"github.com/fatih/color"
	glog "github.com/maahsome/golang-logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	c = &Config{}
)

type Config struct {
	SymLinkDir string
	BinDir     string
	Logger     *logrus.Logger
}

func longDescription() string {
	longText := ""
	yellow := color.New(color.FgYellow).SprintFunc()

	longText += `EXAMPLE:
    List the versions of "terraform"`

	longText = fmt.Sprintf("%s\n\n    > %s\n\n", longText,
		yellow("binary-version-switcher terraform versions"))

	longText += `EXAMPLE:
    Activate a specific version of "terraform"`

	longText = fmt.Sprintf("%s\n\n    > %s\n\n", longText,
		yellow("binary-version-switcher terraform activate -v \"v1.4.6\""))

	return longText
}

var MainCmd = &cobra.Command{
	Use:   "terraform",
	Short: "Switch between versions of terraform",
	Long:  longDescription(),
}

func InitMainCmd(sym string, bin string, loglevel string) {
	activateCmd.Flags().StringP("version", "v", "", "Specify the version")
	versionsCmd.Flags().StringP("version", "v", "", "Prefix for versions to return")
	versionsCmd.Flags().BoolP("all", "a", false, "Return ALL versions")

	MainCmd.AddCommand(activateCmd)
	MainCmd.AddCommand(versionsCmd)

	c.SymLinkDir = sym
	c.BinDir = bin
	c.Logger = glog.CreateStandardLogger()
	c.Logger.Level = glog.LogLevelFromString(loglevel)
}

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List versions for terraform",
	Long:  longDescription(),
	Run: func(cmd *cobra.Command, args []string) {
		c.Logger.Info("Fetching a list of versions...")
		verMatch, _ := cmd.Flags().GetString("version")
		returnAll, _ := cmd.Flags().GetBool("all")
		listVersions(verMatch, returnAll)
	},
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "terraform activate",
	Long:  longDescription(),
	Run: func(cmd *cobra.Command, args []string) {

		activateVer, _ := cmd.Flags().GetString("version")
		activateVersion(activateVer, c.BinDir, c.SymLinkDir)
	},
}

// #!/bin/bash
// # lrwxr-xr-x  1 cmaahs  staff    55B Apr 13 06:43 /usr/local/bin/terraform -> /Applications/Docker.app/Contents/Resources/bin/terraform

// function usage() {
//   echo "Usage: ${0}"
//   echo "  -v 0.12.24    : download and symlink version 0.12.24"
//   echo "  -l            : list all versions"
//   echo "  -h            : display help"
//   exit 0
// }

// function versions() {
//   echo $(curl -s https://releases.hashicorp.com/terraform/ | grep 'href.*/terraform/' | cut -d'/' -f3 | tr '\n' ',')
//   exit 0
// }

// TF_VERSION=${1:-0.11.13}
// while getopts ":hlv:" o; do
//   case "${o}" in
//     v)
//       TF_VERSION=${OPTARG}
//       ;;
//     l)
//       versions
//       ;;
//     h)
//       usage
//       ;;
//     *)
//       usage
//       ;;
//   esac
// done
// shift $((OPTIND-1))

// # MacOS = 'darwin', Linux = 'linux', Windows = 'windows'
// if [[ "$OSTYPE" == "linux-gnu" ]]; then
//   TARGET_OS=linux
//   GROUP_OWN=root
// elif [[ "$OSTYPE" == "darwin"* ]]; then
//   TARGET_OS=darwin
//   GROUP_OWN=admin
// fi

// if [ ! -d "/usr/local/terraform" ]; then
//   MY_LOGON=$(whoami)
//   echo ${MY_LOGON}
//   sudo mkdir -p /usr/local/terraform
//   sudo chown ${MY_LOGON}:${GROUP_OWN} /usr/local/terraform
// fi;
// if [ ! -d "/usr/local/terraform/${TF_VERSION}" ]; then
//   mkdir -p "/usr/local/terraform/${TF_VERSION}"
//   #chown ${MY_LOGON}:${GROUP_OWN} "/usr/local/terraform/${TF_VERSION}"
//   # https://releases.hashicorp.com/terraform/0.12.24/terraform_0.12.24_darwin_amd64.zip
//   wget --quiet -O "/usr/local/terraform/${TF_VERSION}/terraform.zip" "https://releases.hashicorp.com/terraform/${TF_VERSION}/terraform_${TF_VERSION}_${TARGET_OS}_amd64.zip"
//   unzip "/usr/local/terraform/${TF_VERSION}/terraform.zip" -d "/usr/local/terraform/${TF_VERSION}/"
//   if [ -f "/usr/local/terraform/${TF_VERSION}/terraform" ]; then
//     chmod 775 "/usr/local/terraform/${TF_VERSION}/terraform"
//     # sudo chown ${MY_LOGON}:${GROUP_OWN} "/usr/local/terraform/${TF_VERSION}/terraform"
//   fi
// fi
// if [ -f "/usr/local/bin/terraform" ]; then
//   rm "/usr/local/bin/terraform"
// fi
// if [ -L "/usr/local/bin/terraform" ]; then
//   rm "/usr/local/bin/terraform"
// fi
// ln -s "/usr/local/terraform/${TF_VERSION}/terraform" "/usr/local/bin/terraform"
// # chown ${MY_LOGON}:${GROUP_OWN} "/usr/local/bin/terraform"
