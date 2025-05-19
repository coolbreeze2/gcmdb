package cmd

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb/client"
	"os"
	"strings"

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
		panic(err)

	} else {
		if info.IsDir() {
			resources, err = parseResourceFromDir(filePath)
		} else {
			var resource client.Object
			resource, err = parseResourceFromFile(filePath)
			resources = append(resources, resource)
		}
		if err != nil {
			panic(err)
		}
	}
	applyResources(resources)
	// fmt.Printf("resources:%v", resources)
}

func addApplyFlags(c *cobra.Command) {
	c.Flags().StringP("filename", "f", "", "File or directory name")
}

func parseResourceFromDir(path string) ([]client.Object, error) {
	var objs []client.Object
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if obj, err := parseResourceFromFile(e.Name()); err == nil {
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
	kind := jsonObj["kind"]

	var o client.Object
	switch kind {
	case "Project":
		o = &client.Project{}
	case "App":
		o = &client.App{}
	}

	if err := yaml.Unmarshal(file, o); err != nil {
		return nil, err
	}

	return o, nil
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
		if err != nil {
			panic(err)
		}
	case client.ObjectNotFoundError:
		// 不存在，则创建
		createResource(r)
	}
	// 已存在，则更新
	updateResource(r)
}

func createResource(r client.Object) {
	var jsonObj, result map[string]any
	var err error

	metadata := r.GetMetadata()

	if err = structToMap(r, &jsonObj); err != nil {
		panic(err)
	}
	if result, err = r.Create(metadata.Name, metadata.Namespace, jsonObj); err != nil {
		panic(err)
	}
	kinds := strings.ToLower(r.GetKind()) + "s"
	if result == nil {
		fmt.Printf("%v/%v created\n", kinds, metadata.Name)
	}
}

func updateResource(r client.Object) {
	var jsonObj, result map[string]any
	var err error

	metadata := r.GetMetadata()

	if err = structToMap(r, &jsonObj); err != nil {
		panic(err)
	}
	if result, err = r.Update(metadata.Name, metadata.Namespace, jsonObj); err != nil {
		panic(err)
	}
	kinds := strings.ToLower(r.GetKind()) + "s"
	if result == nil {
		fmt.Printf("%v/%v unchanged\n", kinds, metadata.Name)
	} else {
		fmt.Printf("%v/%v configured\n", kinds, metadata.Name)
	}
}

func mapToStruct(m map[string]any, s any) error {
	// 先将 map 转为 JSON
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// 再将 JSON 解析到结构体
	return json.Unmarshal(data, s)
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
