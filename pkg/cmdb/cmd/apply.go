package cmd

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb/client"
	"os"
	"path"
	"sort"

	"github.com/goccy/go-yaml"
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
	var resources []client.Object

	if info, err := os.Stat(filePath); err != nil {
		CheckError(err)
	} else {
		if info.IsDir() {
			resources, err = parseResourceFromDir(filePath)
		} else {
			var resource client.Object
			resource, err = parseResourceFromFile(filePath)
			resources = append(resources, resource)
		}
		CheckError(err)
	}
	checkResourceTypeExist(resources)
	sortResource(resources)
	applyResources(resources)
}

func addApplyFlags(c *cobra.Command) {
	c.Flags().StringP("filename", "f", "", "File or directory name")
}

func parseResourceFromDir(dirPath string) ([]client.Object, error) {
	var objs []client.Object
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		filePath := path.Join(dirPath, e.Name())
		if obj, err := parseResourceFromFile(filePath); err == nil {
			objs = append(objs, obj)
		} else {
			return nil, err
		}
	}
	return objs, nil
}

func parseResourceFromFile(path string) (client.Object, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var jsonObj map[string]any
	if err = yaml.Unmarshal(file, &jsonObj); err != nil {
		return nil, err
	}
	kind := jsonObj["kind"].(string)

	o, err := GetResourceKindByString(kind)
	CheckError(err)

	if err := yaml.Unmarshal(file, o); err != nil {
		return nil, err
	}

	return o, nil
}

// 检查资源类型是否存在
func checkResourceTypeExist(resources []client.Object) error {
	for _, v := range resources {
		kind := v.GetKind()
		exist := false
		for _, k := range client.ResourceOrder {
			if k == kind {
				exist = true
			}
		}
		if !exist {
			return client.ResourceTypeError{Kind: kind}
		}
	}
	return nil
}

// 根据资源优先级排序
func sortResource(resources []client.Object) error {
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

func applyResources(resources []client.Object) {
	for i := range resources {
		applyResource(resources[i])
	}
}

func applyResource(r client.Object) {
	metadata := r.GetMetadata()
	_, err := r.Read(metadata.Name, metadata.Namespace, 0)
	switch err.(type) {
	default:
		CheckError(err)
	case client.ResourceNotFoundError:
		// 不存在，则创建
		createUpdateResource(r, "CREATE")
	case nil:
		// 已存在，则更新
		createUpdateResource(r, "UPDATE")
	}
}

func createUpdateResource(r client.Object, action string) {
	var jsonObj, result map[string]any
	var err error

	metadata := r.GetMetadata()

	err = structToMap(r, &jsonObj)
	CheckError(err)

	switch action {
	case "CREATE":
		result, err = r.Create(metadata.Name, metadata.Namespace, jsonObj)
	case "UPDATE":
		result, err = r.Update(metadata.Name, metadata.Namespace, jsonObj)
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
