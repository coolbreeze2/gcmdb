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

func init() {
	addPersistentFlags()

	objects := []cmdb.Resource{}
	for _, kind := range global.ResourceOrder {
		o, err := cmdb.NewResourceWithKind(kind)
		if err == nil {
			objects = append(objects, o)
		}
	}
	InitMutilGetCmd(objects)
	InitMutilDeleteCmd(objects)
}
