package testing

import (
	"goTool/pkg/cmdb/client"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUrlJoin(t *testing.T) {
	baseUrl := "http://123.com/api/v1"
	expectedUrl := "http://123.com/api/v1/apps/dev-app/"
	url, err := client.UrlJoin(baseUrl, "apps", "dev-app/")
	assert.Equal(t, expectedUrl, url)
	assert.NoError(t, err)
}

func TestReadroject(t *testing.T) {
	o := client.NewProject()
	obj, err := o.Read("devops", "", 0)
	assert.IsType(t, map[string]any{}, obj)
	assert.NoError(t, err)
}

func TestListProject(t *testing.T) {
	o := client.NewProject()
	objs, err := o.List(&client.ListOptions{})
	assert.Less(t, 0, len(objs))
	assert.NoError(t, err)
}

func TestUpdateProject(t *testing.T) {
	o := client.NewProject()
	obj, err := o.Read("devops", "", 0)
	assert.NoError(t, err)

	newName := RandomString(6)
	obj["spec"].(map[string]any)["nameInChain"] = newName
	obj, err = o.Update("devops", "", obj)
	assert.NoError(t, err)
	assert.Equal(t, newName, obj["spec"].(map[string]any)["nameInChain"])

	obj2, err := o.Update("devops", "", obj)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(obj2))
}

// 生成随机字符串
func RandomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func init() {
	os.Setenv("CMDB_API_URL", "http://127.0.0.1:8080/api/v1")
}
