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
	rootCmd.AddCommand(getCmd)
}

func NewCommand(p cmdb.IResource) *cobra.Command {
	kind := strings.ToLower(p.GetKind())
	GetCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [name]", kind),
		Short: "Get resources",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			page, _ := cmd.Flags().GetInt64("page")
			limit, _ := cmd.Flags().GetInt64("limit")
			output, _ := cmd.Flags().GetString("output")
			selector, _ := cmd.Flags().GetString("selector")
			field_selector, _ := cmd.Flags().GetString("field_selector")
			opt := cmdb.NewListOptions(
				page,
				limit,
				cmdb.ParseSelector(selector),
				cmdb.ParseSelector(field_selector),
			)
			rsb := p.List(opt)
			rs := []cmdb.Resource{}
			if err := json.Unmarshal(rsb, &rs); err != nil {
				panic(err)
			}
			switch output {
			case "simple":
				tableHeader := []string{"NAME", "CREATED_AT"}
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader(tableHeader)
				table.SetBorder(false)
				table.SetColumnSeparator("")
				table.SetHeaderLine(false)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				for _, r := range rs {
					createTime, _ := r.Metadata.CreationTimeStamp.Local().MarshalText()
					row := []string{r.Metadata.Name, string(createTime)}
					table.Append(row)
				}
				table.Render()
			case "json":
				rsbts, _ := json.MarshalIndent(rs, "", "  ")
				fmt.Printf("%v", string(rsbts))
			case "yaml":
				var s []string
				for _, r := range rs {
					byts, _ := yaml.Marshal(r)
					s = append(s, string(byts))
				}
				result := strings.Join(s, "---\n")
				fmt.Printf("%v", result)
			}
		},
	}
	GetCmd.Flags().StringP("name", "n", "", "specify name")
	GetCmd.Flags().StringP("output", "o", "simple", "page number")
	GetCmd.Flags().Int64P("page", "p", 0, "page number")
	GetCmd.Flags().Int64P("limit", "s", 0, "limit size, 0 is no limit")
	GetCmd.Flags().StringP("selector", "l", "", "specify name")
	GetCmd.Flags().String("field-selector", "", "specify name")
	return GetCmd
}
