package cmd

import (
	"github.com/spf13/cobra"
)

// sqsCmd represents the sqs command
var sqsCmd = &cobra.Command{
	Use: "sqs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sqsCmd)
}
