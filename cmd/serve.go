/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	exitWhenSynced bool
	metricsPort    int
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a metrics server and poll the configured clients.",
	Run: func(cmd *cobra.Command, args []string) {
		initCommon()

		err := ethClient.Serve(ctx, metricsPort)
		if err != nil {
			logr.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().BoolVarP(&exitWhenSynced, "exit-when-synced", "", false, "Exit the program when both clients are synced")
	serveCmd.Flags().IntVarP(&metricsPort, "metrics-port", "", 9090, "Port to serve Prometheus metrics on")
}
