package cmd

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/client"
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
	var resources []cmdb.Resource

	if info, err := os.Stat(filePath); err != nil {
		CheckError(err)
	} else {
		if info.IsDir() {
			resources, err = client.ParseResourceFromDir(filePath)
		} else {
			var resource cmdb.Resource
			resource, err = client.ParseResourceFromFile(filePath)
			resources = append(resources, resource)
		}
		CheckError(err)
	}
	CheckError(checkResourceTypeExist(resources))
	sortResource(resources)
	applyResources(resources)
}

func addApplyFlags(c *cobra.Command) {
	c.Flags().StringP("filename", "f", "", "File or directory name")
}

// 检查资源类型是否存在
func checkResourceTypeExist(resources []cmdb.Resource) error {
	for _, v := range resources {
		kind := v.GetKind()
		exist := false
		for _, k := range client.ResourceOrder {
			if k == kind {
				exist = true
			}
		}
		if !exist {
			return cmdb.ResourceTypeError{Kind: kind}
		}
	}
	return nil
}

// 根据资源优先级排序
func sortResource(resources []cmdb.Resource) error {
	orders := map[string]int{}
	for i, v := range client.ResourceOrder {
		orders[v] = i
	}
	sort.Slice(resources, func(i, j int) bool {
		a := orders[resources[i].GetKind()]
		b := orders[resources[j].GetKind()]
		return a < b
	})
	return nil
}

func applyResources(resources []cmdb.Resource) {
	for i := range resources {
		applyResource(resources[i])
	}
}

func applyResource(r cmdb.Resource) {
	metadata := r.GetMeta()
	cli := client.DefaultCMDBClient
	_, err := cli.ReadResource(r, metadata.Name, metadata.Namespace, 0)
	switch err.(type) {
	default:
		CheckError(err)
	case cmdb.ResourceNotFoundError:
		// 不存在，则创建
		createUpdateResource(r, "CREATE")
	case nil:
		// 已存在，则更新
		createUpdateResource(r, "UPDATE")
	}
}

func createUpdateResource(r cmdb.Resource, action string) {
	var jsonObj, result map[string]any
	var err error

	metadata := r.GetMeta()

	err = structToMap(r, &jsonObj)
	CheckError(err)

	cli := client.DefaultCMDBClient

	switch action {
	case "CREATE":
		result, err = cli.CreateResource(r)
	case "UPDATE":
		result, err = cli.UpdateResource(r)
	}
	CheckError(err)

	lkind := client.LowerKind(r)
	if result == nil {
		fmt.Printf("%v/%v unchanged\n", lkind, metadata.Name)
	} else if action == "UPDATE" {
		fmt.Printf("%v/%v configured\n", lkind, metadata.Name)
	} else if action == "CREATE" {
		fmt.Printf("%v/%v created\n", lkind, metadata.Name)
	}
}

func structToMap(s any, m *map[string]any) error {
	// 先将 struct 转为 JSON
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	// 再将 JSON 解析到 map
	return json.Unmarshal(data, m)
}
