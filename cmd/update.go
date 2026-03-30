package cmd

import (
	"github.com/lian-yang/trans/internal/updater"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update trans to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updater.Update(version)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
