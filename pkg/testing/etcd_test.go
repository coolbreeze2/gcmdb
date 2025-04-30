package testing

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestEtcdStore(t *testing.T) {
	// 创建嵌入式 etcd 服务器
	inCfg := NewTestConfig(t)
	server, err := embed.StartEtcd(inCfg)
	if err != nil {
		t.Fatalf("failed to start etcd: %v", err)
	}
	defer server.Close()

	// 创建 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{server.Clients[0].Addr().String()},
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer cli.Close()

	// TODO: 测试用例操作
	// ...
}

// NewTestConfig returns a configuration for an embedded etcd server.
// The configuration is based on embed.NewConfig(), with the following adjustments:
//   - sets UnsafeNoFsync = true to improve test performance (only reasonable in a test-only
//     single-member server we never intend to restart or keep data from)
//   - uses free ports for client and peer listeners
//   - cleans up the data directory on test termination
//   - silences server logs other than errors
func NewTestConfig(t *testing.T) *embed.Config {
	cfg := embed.NewConfig()

	cfg.UnsafeNoFsync = true

	ports, err := getAvailablePorts(2)
	if err != nil {
		t.Fatal(err)
	}
	clientURL := url.URL{Scheme: "http", Host: net.JoinHostPort("localhost", strconv.Itoa(ports[0]))}
	peerURL := url.URL{Scheme: "http", Host: net.JoinHostPort("localhost", strconv.Itoa(ports[1]))}

	cfg.ListenPeerUrls = []url.URL{peerURL}
	cfg.AdvertisePeerUrls = []url.URL{peerURL}
	cfg.ListenClientUrls = []url.URL{clientURL}
	cfg.AdvertiseClientUrls = []url.URL{clientURL}
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)

	cfg.ZapLoggerBuilder = embed.NewZapLoggerBuilder(zaptest.NewLogger(t, zaptest.Level(zapcore.ErrorLevel)).Named("etcd-server"))
	cfg.Dir = t.TempDir()
	os.Chmod(cfg.Dir, 0700)
	return cfg
}

// getAvailablePort returns a TCP port that is available for binding.
func getAvailablePorts(count int) ([]int, error) {
	ports := []int{}
	for i := 0; i < count; i++ {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			return nil, fmt.Errorf("could not bind to a port: %v", err)
		}
		// It is possible but unlikely that someone else will bind this port before we get a chance to use it.
		defer l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
}
