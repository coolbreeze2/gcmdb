package cmd

import (
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
		if p.Use == "get" && len(args) != 0 {
			return options, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		}
		kind := cmd.Short
		namespace := "" // TODO:
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
	options, err := o.GetNames(namespace)
	CheckError(err)
	return options
}

// TODO: Completion namespace
func completeNamespace() []string {
	var optiosn []string
	return optiosn
}

func GetResourceKindByString(kind string) (client.Object, error) {
	kind = strings.ToLower(kind)
	switch kind {
	case "project":
		return client.NewProject(), nil
	case "app":
		return client.NewApp(), nil
	}
	return nil, client.ResourceTypeError{Kind: kind}
}
