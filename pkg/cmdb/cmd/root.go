package cmd

import (
	"fmt"
	"goTool/global"
	"goTool/pkg/cmdb"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cmctl",
	Short: "Cmctl is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
		   love by spf13 and friends in Go.
		   Complete documentation is available at https://gohugo.io`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run cmctl...")
	},
}

func Execute() {
	CheckError(RootCmd.Execute())
}

func addPersistentFlags() {
	RootCmd.PersistentFlags().StringP("namespace", "n", "", "resource namespace")
}

func parseNamespace(c *cobra.Command, o cmdb.Object) string {
	namespace, _ := c.Root().PersistentFlags().GetString("namespace")
	if o.GetMeta().HasNamespace() && namespace == "" {
		fatalErrHandler(fmt.Sprintf("error: a namespace must be specified for %s", o.GetKind()), 1)
	}
	return namespace
}

func init() {
	addPersistentFlags()

	objects := []cmdb.Object{}
	for _, kind := range global.ResourceOrder {
		o, err := cmdb.NewResourceWithKind(kind)
		if err == nil {
			objects = append(objects, o)
		}
	}
	InitMutilGetCmd(objects)
	InitMutilDeleteCmd(objects)
}
