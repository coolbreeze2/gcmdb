package cmd

import (
	"fmt"
	apiv1 "gcmdb/pkg/cmdb/server/apis/v1"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

var server *http.Server

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start server",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(c *cobra.Command, args []string) {
		serveCmdHandle(c)
	},
}

func init() {
	addServeFlags(serveCmd)
	RootCmd.AddCommand(serveCmd)
}

func serveCmdHandle(c *cobra.Command) {
	port, _ := c.Flags().GetInt16("port")
	serveStart(port)
}

func addServeFlags(c *cobra.Command) {
	c.Flags().Int16P("port", "p", 3333, "Serve port")
}

func serveStart(port int16) {
	addr := fmt.Sprintf(":%s", strconv.Itoa(int(port)))
	fmt.Printf("serve address: %s\n", addr)
	server = &http.Server{Addr: addr, Handler: apiv1.NewRouter(nil)}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
