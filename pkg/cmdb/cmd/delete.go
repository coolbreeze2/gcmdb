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
}

func InitMutilDeleteCmd(objs []cmdb.Object) {
	for _, o := range objs {
		deleteCmd.AddCommand(newDeleteCmd(o))
	}
	RootCmd.AddCommand(deleteCmd)
}

func newDeleteCmd(r cmdb.Object) *cobra.Command {
	// TODO: 支持 -f 从文件删除，类似 apply
	kind := strings.ToLower(r.GetKind())
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <name>...", kind),
		Short: kind,
		Long:  fmt.Sprintf("Delete %s", kind),
		Args:  cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			deleteCmdHandle(c, r, args)
		},
		ValidArgsFunction: CompleteFunc,
	}
	return cmd
}

func deleteCmdHandle(c *cobra.Command, r cmdb.Object, args []string) {
	namespace := parseNamespace(c, r)
	var name string

	cli := client.DefaultCMDBClient
	for index := range args {
		name = args[index]
		CheckError(cli.DeleteResource(r, name, namespace))
		fmt.Printf("%v %v deleted.\n", client.LowerKind(r), name)
	}
}
