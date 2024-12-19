package cmd

import (
	"github.com/spf13/cobra"
)

// s3Cmd represents the s3 command
var s3Cmd = &cobra.Command{
	Use: "s3",
}

func init() {
	rootCmd.AddCommand(s3Cmd)
}
