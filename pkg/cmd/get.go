package cmd

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func init() {
	getCmd.AddCommand(NewCommand(cmdb.NewProject()))
	getCmd.AddCommand(NewCommand(cmdb.NewApp()))
	RootCmd.AddCommand(getCmd)
}

func NewCommand(r cmdb.IResource) *cobra.Command {
	kind := strings.ToLower(r.GetKind())
	GetCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [name]", kind),
		Short: "Get resources",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(c *cobra.Command, args []string) {
			getCmdHandle(c, r, args)
		},
	}
	addFlags(GetCmd)
	return GetCmd
}

func getCmdHandle(c *cobra.Command, r cmdb.IResource, args []string) {
	outputFmt, _ := c.Flags().GetString("output")
	revision, _ := c.Flags().GetInt64("revision")
	opt := parseListOptionsFlags(c)
	var name string
	var resources []map[string]interface{}

	if len(args) == 1 {
		name = args[0]
		resources = append(resources, r.Read(name, opt.Namespace, revision))
	} else {
		resources = append(resources, r.List(opt)...)
	}

	switch outputFmt {
	case "simple":
		outputFmtSimple(resources)
	case "json":
		outputFmtJson(resources)
	case "yaml":
		outputFmtYaml(resources)

	}
}

func addFlags(c *cobra.Command) {
	// TODO: namespace 应为全局参数
	c.Flags().StringP("namespace", "n", "", "namespace name")
	c.Flags().StringP("output", "o", "simple", "page number")
	c.Flags().Int64P("page", "p", 0, "page number")
	c.Flags().Int64P("limit", "s", 0, "limit size, 0 is no limit")
	c.Flags().StringP("selector", "l", "", "specify name")
	c.Flags().String("field-selector", "", "specify name")
}

func parseListOptionsFlags(c *cobra.Command) *cmdb.ListOptions {
	namespace, _ := c.Flags().GetString("namespace")
	page, _ := c.Flags().GetInt64("page")
	limit, _ := c.Flags().GetInt64("limit")
	selector, _ := c.Flags().GetString("selector")
	field_selector, _ := c.Flags().GetString("field_selector")
	opt := cmdb.NewListOptions(
		namespace,
		page,
		limit,
		cmdb.ParseSelector(selector),
		cmdb.ParseSelector(field_selector),
	)
	return opt
}

func outputFmtSimple(resources []map[string]interface{}) {
	tableHeader := []string{"NAME", "CREATED_AT"}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeader)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	for _, r := range resources {
		metadataField := r["metadata"].(map[string]interface{})
		createTime := metadataField["creationTimestamp"].(string)
		name := metadataField["name"].(string)
		row := []string{name, string(createTime)}
		table.Append(row)
	}
	table.Render()
}

func outputFmtJson(resources []map[string]interface{}) {
	rsbts, _ := json.MarshalIndent(resources, "", "  ")
	fmt.Printf("%v", string(rsbts))
}

func outputFmtYaml(resources []map[string]interface{}) {
	var s []string
	for _, r := range resources {
		byts, _ := yaml.MarshalWithOptions(r, yaml.AutoInt())
		s = append(s, string(byts))
	}
	result := strings.Join(s, "---\n")
	fmt.Printf("%v", result)
}
