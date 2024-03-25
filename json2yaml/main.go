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
    List the versions of "json2yaml"`

	longText = fmt.Sprintf("%s\n\n    > %s\n\n", longText,
		yellow("binary-version-switcher json2yaml versions"))

	longText += `EXAMPLE:
    Activate a specific version of "json2yaml"`

	longText = fmt.Sprintf("%s\n\n    > %s\n\n", longText,
		yellow("binary-version-switcher json2yaml activate -v \"v1.28.0\""))

	return longText
}

var MainCmd = &cobra.Command{
	Use:   "json2yaml",
	Short: "Switch between versions of json2yaml",
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
	Short: "json2yaml list versions",
	Long:  fmt.Sprintf("%s\n\n%s", longDescription(), `Unless "-a" is specified, only the highest PATCH for each MAJOR.MINOR will be returned`),
	Run: func(cmd *cobra.Command, args []string) {
		c.Logger.Info("Fetching a list of versions...")
		verMatch, _ := cmd.Flags().GetString("version")
		returnAll, _ := cmd.Flags().GetBool("all")
		listVersions(verMatch, returnAll)
	},
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "json2yaml activate",
	Long:  longDescription(),
	Run: func(cmd *cobra.Command, args []string) {

		activateVer, _ := cmd.Flags().GetString("version")
		activateVersion(activateVer, c.BinDir, c.SymLinkDir)
	},
}
