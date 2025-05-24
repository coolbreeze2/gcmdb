package cmd

import (
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/client"
	"strings"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func InitMutilDeleteCmd(objs []cmdb.Resource) {
	for _, o := range objs {
		deleteCmd.AddCommand(newDeleteCmd(o))
	}
	RootCmd.AddCommand(deleteCmd)
}

func newDeleteCmd(r cmdb.Resource) *cobra.Command {
	kind := strings.ToLower(r.GetKind())
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <name>...", kind),
		Short: kind,
		Long:  fmt.Sprintf("Delete %s", kind),
		Args:  cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			getDeleteHandle(c, r, args)
		},
		ValidArgsFunction: CompleteFunc,
	}
	return cmd
}

func getDeleteHandle(c *cobra.Command, r cmdb.Resource, args []string) {
	namespace, _ := c.Root().PersistentFlags().GetString("namespace")
	var err error
	var name string

	cli := client.DefaultCMDBClient
	for index := range args {
		name = args[index]
		if err = cli.DeleteResource(r, name, namespace); err != nil {
			CheckError(err)
		} else {
			fmt.Printf("%v %v deleted.\n", client.LowerKind(r), name)
		}
	}
}
