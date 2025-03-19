package cmd

import (
	"fmt"
	"goTool/pkg/cmdb"
	"strings"

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
	getCmd.AddCommand(NewCommand(cmdb.NewProject().Resource))
	rootCmd.AddCommand(getCmd)
}

func NewCommand(p cmdb.Resource) *cobra.Command {
	kind := strings.ToLower(p.GetKind())
	GetCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [name]", kind),
		Short: "Get resources",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			page, _ := cmd.Flags().GetInt("page")
			limit, _ := cmd.Flags().GetInt("limit")
			fmt.Printf("page: %v limit: %v\n", page, limit)
			p.List()
		},
	}
	GetCmd.Flags().StringP("name", "n", "", "specify name")
	GetCmd.Flags().IntP("page", "p", 0, "page number")
	GetCmd.Flags().IntP("limit", "s", 10, "limit size, 0 is no limit")
	return GetCmd
}
