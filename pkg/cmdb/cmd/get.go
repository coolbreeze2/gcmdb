package cmd

import (
	"encoding/json"
	"fmt"
	"gcmdb/pkg/cmdb"
	"gcmdb/pkg/cmdb/client"
	"gcmdb/pkg/cmdb/conversion"
	"os"
	"strings"
	"time"

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
		outputFmtSimple(resources, r, opt.All)
	}
}

func addGetFlags(c *cobra.Command) {
	c.Flags().BoolP("all-namespaces", "A", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	c.Flags().StringP("output", "o", "simple", "output format")
	c.Flags().Int64P("page", "p", 1, "page number")
	c.Flags().Int64P("limit", "s", 0, "limit size, 0 is no limit")
	c.Flags().StringP("selector", "l", "", "label selector")
	c.Flags().String("field-selector", "", "field selector")
}

func parseListOptionsFlags(c *cobra.Command, o cmdb.Object) *client.ListOptions {
	all, _ := c.Flags().GetBool("all-namespaces")
	namespace, _ := c.Root().PersistentFlags().GetString("namespace")
	if o.GetMeta().HasNamespace() && namespace == "" && all == false {
		CheckError(fmt.Errorf("error: a namespace must be specified for %s", o.GetKind()))
	}
	page, _ := c.Flags().GetInt64("page")
	limit, _ := c.Flags().GetInt64("limit")
	selector, _ := c.Flags().GetString("selector")
	field_selector, _ := c.Flags().GetString("field_selector")
	return &client.ListOptions{
		All:           all,
		Namespace:     namespace,
		Page:          page,
		Limit:         limit,
		Selector:      conversion.ParseSelector(selector),
		FieldSelector: conversion.ParseSelector(field_selector),
	}
}

func outputFmtSimple(resources []map[string]any, r cmdb.Object, all bool) {
	tableHeader := []string{"NAME"}

	// 不同 Resource 支持自定义 Column
	extraColumns, hasExCol := extraCustomColumn[strings.ToLower(r.GetKind())]
	addNamespaceCol := r.GetMeta().HasNamespace() && all
	if addNamespaceCol {
		tableHeader = append(tableHeader, "NAMESPACE")
	}
	if hasExCol {
		for _, c := range extraColumns {
			tableHeader = append(tableHeader, c.name)
		}
	}
	tableHeader = append(tableHeader, "AGE")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(tableHeader)
	table.SetBorder(false)
	table.SetColumnSeparator("")
	table.SetHeaderLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	for _, r := range resources {
		metadataField := r["metadata"].(map[string]any)
		createTime := formatAge(metadataField["creationTimestamp"].(string))
		name := metadataField["name"].(string)
		row := []string{name}
		if addNamespaceCol {
			row = append(row, metadataField["namespace"].(string))
		}
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

func formatAge(t string) string {
	layout := "2006-01-02T15:04:05.9999999-07:00"
	time_, _ := time.Parse(layout, t)
	now := time.Now()
	duration := now.Sub(time_)
	return HumanDuration(duration)
}

// HumanDuration returns a succint representation of the provided duration
// with limited precision for consumption by humans. It provides ~2-3 significant
// figures of duration.
func HumanDuration(d time.Duration) string {
	// Allow deviation no more than 2 seconds(excluded) to tolerate machine time
	// inconsistence, it can be considered as almost now.
	if seconds := int(d.Seconds()); seconds < -1 {
		return "<invalid>"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60*2 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := int(d / time.Minute)
	if minutes < 10 {
		s := int(d/time.Second) % 60
		if s == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm%ds", minutes, s)
	} else if minutes < 60*3 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := int(d / time.Hour)
	if hours < 8 {
		m := int(d/time.Minute) % 60
		if m == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh%dm", hours, m)
	} else if hours < 48 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*8 {
		h := hours % 24
		if h == 0 {
			return fmt.Sprintf("%dd", hours/24)
		}
		return fmt.Sprintf("%dd%dh", hours/24, h)
	} else if hours < 24*365*2 {
		return fmt.Sprintf("%dd", hours/24)
	} else if hours < 24*365*8 {
		dy := int(hours/24) % 365
		if dy == 0 {
			return fmt.Sprintf("%dy", hours/24/365)
		}
		return fmt.Sprintf("%dy%dd", hours/24/365, dy)
	}
	return fmt.Sprintf("%dy", int(hours/24/365))
}
