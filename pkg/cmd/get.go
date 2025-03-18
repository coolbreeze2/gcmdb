package cmd

import (
	"goTool/pkg/cmd/get"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	getCmd.AddCommand(get.GetProjectCmd)
	rootCmd.AddCommand(getCmd)
}
