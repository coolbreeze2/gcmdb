package cmd

import (
	"encoding/json"
	"fmt"
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/client"
	"goTool/pkg/cmdb/conversion"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type customColumn struct {
	name, path string
}

// 资源额外列
var extraCustomColumn = map[string][]customColumn{
	"app":               {{"PROJECT", "spec.project"}, {"SCM", "spec.scm.name"}},
	"appdeployment":     {{"STATUS", "status"}, {"FLOW_RUN_ID", "flow_run_id"}, {"PROJECT", "spec.template.spec.project"}, {"APP", "spec.template.spec.app"}},
	"datacenter":        {{"PROVIDER", "spec.provider"}},
	"project":           {{"NAME_IN_CHAIN", "spec.nameInChain"}},
	"scm":               {{"DATACENTER", "spec.datacenter"}, {"URL", "spec.url"}, {"SERVICE", "spec.service"}},
	"containerregistry": {{"TYPE", "spec.type"}, {"DATACENTER", "spec.datacenter"}, {"REGISTRY", "spec.registry"}},
	"zone":              {{"PROVIDER", "spec.provider"}},
	"deployplatform":    {{"DATACENTER", "spec.datacenter"}},
	"helmrepository":    {{"DATACENTER", "spec.datacenter"}, {"URL", "spec.url"}},
	"hostnode":          {{"DATACENTER", "spec.datacenter"}, {"ZONE", "spec.zone"}, {"HOSTNAME", "spec.hostname"}, {"IP", "spec.ip"}, {"PHASE", "status.phase"}},
	"namespace":         {{"ENV", "spec.bizEnv"}, {"UNIT", "spec.bizUnit"}, {"DATACENTER", "spec.datacenter"}},
	"orchestration":     {{"PREFECT_DEPLOY", "spec.name"}},
	"resourcerange":     {{"DEPLOY_TEMPLATE", "deployTemplate.name"}},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
}

func InitMutilGetCmd(objs []cmdb.Object) {
	for _, o := range objs {
		getCmd.AddCommand(newGetCmd(o))
	}
	RootCmd.AddCommand(getCmd)
}

func newGetCmd(r cmdb.Object) *cobra.Command {
	kind := strings.ToLower(r.GetKind())
	GetCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [name]", kind),
		Short: kind,
		Long:  fmt.Sprintf("Get %s", kind),
		Args:  cobra.RangeArgs(0, 1),
		Run: func(c *cobra.Command, args []string) {
			getCmdHandle(c, r, args)
		},
		ValidArgsFunction: CompleteFunc,
	}
	addGetFlags(GetCmd)
	return GetCmd
}

func getCmdHandle(c *cobra.Command, r cmdb.Object, args []string) {
	outputFmt, _ := c.Flags().GetString("output")
	revision, _ := c.Flags().GetInt64("revision")
	opt := parseListOptionsFlags(c, r)
	var err error
	var name string
	var resources []map[string]any

	cli := client.DefaultCMDBClient
	if len(args) == 1 {
		name = args[0]
		var resource map[string]any
		resource, err = cli.ReadResource(r, name, opt.Namespace, revision)
		resources = append(resources, resource)
	} else {
		resources, err = cli.ListResource(r, opt)
	}

	CheckError(err)

	switch outputFmt {
	case "json":
		outputFmtJson(resources)
	case "yaml":
		outputFmtYaml(resources)
	default:
		outputFmtSimple(resources, r)
	}
}

func addGetFlags(c *cobra.Command) {
	c.Flags().StringP("output", "o", "simple", "page number")
	c.Flags().Int64P("page", "p", 0, "page number")
	c.Flags().Int64P("limit", "s", 0, "limit size, 0 is no limit")
	c.Flags().StringP("selector", "l", "", "label selector")
	c.Flags().String("field-selector", "", "field selector")
}

func parseListOptionsFlags(c *cobra.Command, o cmdb.Object) *client.ListOptions {
	namespace := parseNamespace(c, o)
	page, _ := c.Flags().GetInt64("page")
	limit, _ := c.Flags().GetInt64("limit")
	selector, _ := c.Flags().GetString("selector")
	field_selector, _ := c.Flags().GetString("field_selector")
	return &client.ListOptions{
		Namespace:     namespace,
		Page:          page,
		Limit:         limit,
		Selector:      conversion.ParseSelector(selector),
		FieldSelector: conversion.ParseSelector(field_selector),
	}
}

func outputFmtSimple(resources []map[string]any, r cmdb.Object) {
	tableHeader := []string{"NAME"}

	// 不同 Resource 支持自定义 Column
	extraColumns, hasExCol := extraCustomColumn[strings.ToLower(r.GetKind())]
	if hasExCol {
		for _, c := range extraColumns {
			tableHeader = append(tableHeader, c.name)
		}
	}
	tableHeader = append(tableHeader, "CREATED_AT")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeader)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	for _, r := range resources {
		metadataField := r["metadata"].(map[string]any)
		createTime := metadataField["creationTimestamp"].(string)
		name := metadataField["name"].(string)
		row := []string{name}
		if hasExCol {
			for _, c := range extraColumns {
				value, ok := conversion.GetMapValueByPath(r, c.path).(string)
				if !ok {
					value = ""
				}
				row = append(row, value)
			}
		}
		row = append(row, string(createTime))
		table.Append(row)
	}
	table.Render()
}

func outputFmtJson(resources []map[string]any) {
	rsbts, _ := json.MarshalIndent(resources, "", "  ")
	fmt.Printf("%v", string(rsbts))
}

func outputFmtYaml(resources []map[string]any) {
	var s []string
	for _, r := range resources {
		byts, _ := yaml.MarshalWithOptions(r, yaml.AutoInt())
		s = append(s, string(byts))
	}
	result := strings.Join(s, "---\n")
	fmt.Printf("%v", result)
}
