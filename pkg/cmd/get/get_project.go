package get

import (
	"fmt"
	"goTool/pkg/cmdb"

	"github.com/spf13/cobra"
)

var GetProjectCmd = &cobra.Command{
	Use:       "project [name]",
	Short:     "Get resources",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"project"},
	Run: func(cmd *cobra.Command, args []string) {
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")
		fmt.Printf("page: %v limit: %v\n", page, limit)
		p := cmdb.NewProject()
		p.List()
	},
}

func init() {
	GetProjectCmd.Flags().IntP("page", "p", 0, "page number")
	GetProjectCmd.Flags().IntP("limit", "s", 10, "limit size, 0 is no limit")
}
