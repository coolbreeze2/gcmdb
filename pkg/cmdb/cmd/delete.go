package cmd

import (
	"fmt"
	"goTool/pkg/cmdb/client"
	"log"
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

func init() {
	deleteCmd.AddCommand(newDeleteCommand(client.NewProject()))
	deleteCmd.AddCommand(newDeleteCommand(client.NewApp()))
	RootCmd.AddCommand(deleteCmd)
}

func newDeleteCommand(r client.Object) *cobra.Command {
	kind := strings.ToLower(r.GetKind())
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <name>", kind),
		Short: fmt.Sprintf("Delete %s", kind),
		Args:  cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			getDeleteHandle(c, r, args)
		},
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
			log.Fatalf("Error from server: %v", err)
		} else {
			log.Printf("%v %v deleted.", client.LowerKind(r), name)
		}
	}
	if err != nil {
		panic(err)
	}
}

func addDeleteFlags(c *cobra.Command) {
	// TODO: namespace 应为全局参数
	c.Flags().StringP("namespace", "n", "", "namespace name")
}
