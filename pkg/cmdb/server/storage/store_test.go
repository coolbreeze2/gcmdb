package storage

import (
	"context"
	"goTool/global"
	"goTool/pkg/cmdb"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestGet(t *testing.T) {
	endpoint := global.ServerSetting.ETCD_SERVER_HOST + ":" + global.ServerSetting.ETCD_SERVER_PORT
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoint},
	})
	assert.NoError(t, err)
	s := New(cli, global.StoragePathPrefix)
	ctx := context.Background()
	var out cmdb.Object
	err = s.Get(ctx, "/apps/go-app", GetOptions{}, &out)
	assert.NoError(t, err)
}
