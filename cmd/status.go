/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Outputs the sync status of the ethereum nodes",
	Run: func(cmd *cobra.Command, args []string) {
		initCommon()

		status, err := export.GetSyncStatus(ctx)
		if err != nil {
			logrus.Fatal(err)
		}

		logr.Info(status)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
