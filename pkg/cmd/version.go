package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var clientVersion = "1.0.0"

// TODO: 从 Server API 获取 server version
var serverVersion = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version numnber of Cmctl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Client: %v, Server: %v", clientVersion, serverVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
