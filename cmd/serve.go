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

		err := export.Serve(ctx, metricsPort)
		if err != nil {
			logr.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
