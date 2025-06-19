package cmd

import (
	"fmt"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/client"
	"gcmdb/pkg/cmdb/conversion"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff resources",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(c *cobra.Command, args []string) {
		diffCmdHandle(c)
	},
}

func init() {
	addDiffFlags(diffCmd)
	RootCmd.AddCommand(diffCmd)
}

func diffCmdHandle(c *cobra.Command) {
	filePath, _ := c.Flags().GetString("filename")
	var resources []cmdb.Object
	var filePaths []string

	info, err := os.Stat(filePath)
	CheckError(err)
	if info.IsDir() {
		resources, filePaths, err = client.ParseResourceFromDir(filePath)
	} else {
		var resource cmdb.Object
		resource, err = client.ParseResourceFromFile(filePath)
		resources = append(resources, resource)
		filePaths = append(filePaths, filePath)
	}
	CheckError(err)
	CheckError(checkResourceTypeExist(resources))
	diffResources(resources, filePaths)
}

func addDiffFlags(c *cobra.Command) {
	c.Flags().StringP("filename", "f", "", "File or directory name")
}

func diffResources(resources []cmdb.Object, filePaths []string) {
	for i := range resources {
		CheckError(diffResource(resources[i], filePaths[i]))
	}
}

func diffResource(o cmdb.Object, filePath string) error {
	meta := o.GetMeta()
	cli := client.DefaultCMDBClient
	serverObj, err := cli.ReadResource(o, meta.Name, meta.Namespace, 0)
	CheckError(err)
	client.RemoveResourceManageFields(serverObj)

	var oMap map[string]any
	CheckError(conversion.StructToMap(o, &oMap))
	client.RemoveResourceManageFields(oMap)

	serverObjBytes, _ := yaml.MarshalWithOptions(serverObj, yaml.AutoInt())
	ObjBytes, _ := yaml.MarshalWithOptions(oMap, yaml.AutoInt())
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(serverObjBytes)),
		B:        difflib.SplitLines(string(ObjBytes)),
		FromFile: "Server",
		ToFile:   filePath,
		Context:  3,
	}
	text, _ := difflib.GetUnifiedDiffString(diff)
	fmt.Println(text)
	return nil
}
