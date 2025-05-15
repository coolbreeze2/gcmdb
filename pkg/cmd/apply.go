package cmd

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb/client"
	"os"

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
	fmt.Printf("resources:%v", resources)
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

	if into, ok := client.KindMap[kind.(string)]; ok {
		if err := mapToStruct(jsonObj, &into); err != nil {
			return nil, err
		} else {
			return into, nil
		}
	}
	return nil, nil
}

func applyResources(resources []client.Object) {
	//
}

func applyResource(resource client.Object) {
	//
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
