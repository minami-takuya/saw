package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var concurrency = 2

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "saw",
	Short: "my aws tools",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", concurrency, "the number of concurrent workers")
}
