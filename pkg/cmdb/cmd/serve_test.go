package cmd

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServe(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer func() {
			assert.IsType(t, http.ErrServerClosed, recover())
		}()
		RootCmd.SetArgs([]string{"serve"})
		RootCmd.Execute()
	}()

	time.Sleep(2 * time.Second)
	// Shut down the server
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("Server Shutdown Failed:%v", err)
	}
	wg.Wait() // Wait for ListenAndServe to return
}
