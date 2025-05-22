package cmd

import (
	"fmt"
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

func InitMutilDeleteCmd(objs []client.Object) {
	for _, o := range objs {
		getCmd.AddCommand(newDeleteCmd(o))
	}
	RootCmd.AddCommand(deleteCmd)
}

func newDeleteCmd(r client.Object) *cobra.Command {
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
	addDeleteFlags(cmd)
	return cmd
}

func getDeleteHandle(c *cobra.Command, r client.Object, args []string) {
	namespace, _ := c.Flags().GetString("namespace")
	var err error
	var name string

	for index := range args {
		name = args[index]
		if _, err = r.Delete(name, namespace); err != nil {
			CheckError(err)
		} else {
			fmt.Printf("%v %v deleted.\n", client.LowerKind(r), name)
		}
	}
}

func addDeleteFlags(c *cobra.Command) {
	// TODO: namespace 应为全局参数
	c.Flags().StringP("namespace", "n", "", "namespace name")
}
