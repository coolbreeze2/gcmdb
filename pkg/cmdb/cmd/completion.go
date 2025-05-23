package cmd

import (
	"goTool/pkg/cmdb"
	"goTool/pkg/cmdb/client"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func CompleteFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var options []string
	p := cmd.Parent()
	completionCmd := p.Use == "get" || p.Use == "delete"
	if p != nil && completionCmd {
		namespace, _ := p.PersistentFlags().GetString("namespace")
		if p.Use == "get" && len(args) != 0 {
			return options, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		}
		kind := cmd.Short
		names := completeName(kind, namespace)
		for _, name := range names {
			if len(args) == 0 || !slices.Contains(args, name) {
				options = append(options, name)
			}
		}
	}
	return options, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

// Completion resource name
func completeName(kind, namespace string) []string {
	o, err := GetResourceKindByString(kind)
	CheckError(err)
	cli := client.DefaultCMDBClient
	options, err := cli.GetResourceNames(o, namespace)
	CheckError(err)
	return options
}

// TODO: Completion namespace
func completeNamespace() []string {
	var optiosn []string
	return optiosn
}

func GetResourceKindByString(kind string) (cmdb.Resource, error) {
	kind = strings.ToLower(kind)
	kindMap := map[string]cmdb.Resource{
		"secret":     cmdb.NewSecret(),
		"scm":        cmdb.NewSCM(),
		"datacenter": cmdb.NewDatacenter(),
		"project":    cmdb.NewProject(),
		"app":        cmdb.NewApp(),
	}
	if r, ok := kindMap[kind]; ok {
		return r, nil
	}
	return nil, client.ResourceTypeError{Kind: kind}
}
