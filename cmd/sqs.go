package cmd

import (
	"github.com/spf13/cobra"
)

// sqsCmd represents the sqs command
var sqsCmd = &cobra.Command{
	Use: "sqs",
}

func init() {
	rootCmd.AddCommand(sqsCmd)
}
