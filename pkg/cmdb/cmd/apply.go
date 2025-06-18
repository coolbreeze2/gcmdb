package cmd

import (
	"fmt"
	"gcmdb/global"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/client"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply resources",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(c *cobra.Command, args []string) {
		applyCmdHandle(c)
	},
}

func init() {
	addApplyFlags(applyCmd)
	RootCmd.AddCommand(applyCmd)
}

func applyCmdHandle(c *cobra.Command) {
	filePath, _ := c.Flags().GetString("filename")
	var resources []cmdb.Object

	info, err := os.Stat(filePath)
	CheckError(err)
	if info.IsDir() {
		resources, err = client.ParseResourceFromDir(filePath)
	} else {
		var resource cmdb.Object
		resource, err = client.ParseResourceFromFile(filePath)
		resources = append(resources, resource)
	}
	CheckError(err)
	CheckError(checkResourceTypeExist(resources))
	sortResource(resources)
	applyResources(resources)
}

func addApplyFlags(c *cobra.Command) {
	c.Flags().StringP("filename", "f", "", "File or directory name")
}

// 检查资源类型是否存在
func checkResourceTypeExist(resources []cmdb.Object) error {
	for _, v := range resources {
		kind := v.GetKind()
		_, err := cmdb.NewResourceWithKind(kind)
		CheckError(err)
	}
	return nil
}

// 根据资源优先级排序
func sortResource(resources []cmdb.Object) error {
	orders := map[string]int{}
	for i, v := range global.ResourceOrder {
		orders[v] = i
	}
	sort.Slice(resources, func(i, j int) bool {
		a := orders[resources[i].GetKind()]
		b := orders[resources[j].GetKind()]
		return a < b
	})
	return nil
}

func applyResources(resources []cmdb.Object) {
	for i := range resources {
		CheckError(applyResource(resources[i]))
	}
}

func applyResource(r cmdb.Object) error {
	meta := r.GetMeta()
	cli := client.DefaultCMDBClient
	_, err := cli.ReadResource(r, meta.Name, meta.Namespace, 0)
	switch err.(type) {
	case cmdb.ResourceNotFoundError:
		// 不存在，则创建
		return createUpdateResource(r, "CREATE")
	case nil:
		// 已存在，则更新
		return createUpdateResource(r, "UPDATE")
	}
	return err
}

func createUpdateResource(r cmdb.Object, action string) error {
	var result map[string]any
	var err error

	metadata := r.GetMeta()

	cli := client.DefaultCMDBClient

	switch action {
	case "CREATE":
		result, err = cli.CreateResource(r)
	case "UPDATE":
		result, err = cli.UpdateResource(r)
	}
	if err != nil {
		return err
	}

	lkind := client.LowerKind(r)
	if result == nil {
		fmt.Printf("%v/%v unchanged\n", lkind, metadata.Name)
	} else if action == "UPDATE" {
		fmt.Printf("%v/%v configured\n", lkind, metadata.Name)
	} else if action == "CREATE" {
		fmt.Printf("%v/%v created\n", lkind, metadata.Name)
	}
	return nil
}
