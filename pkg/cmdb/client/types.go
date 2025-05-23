package client

var ResourceOrder = [...]string{
	"Secret",
	"Project",
	"Datacenter",
	"Zone",
	"Namespace",
	"SCM",
	"HostNode",
	"HelmRepository",
	"ContainerRegistry",
	"App",
	"ConfigCenter",
	"DeployPlatform",
	"DeployTemplate",
	"ResourceRange",
	"Orchestration",
	"AppDeployment",
	"AppInstance",
	"AppInstanceRun",
	"VirtualNetwork",
	"Subnet",
	"DatabaseService",
}

var DefaultCMDBClient = &CMDBClient{}

func NewCMDBClient(apiUrl string) *CMDBClient {
	return &CMDBClient{ApiUrl: apiUrl}
}

type CMDBClient struct {
	ApiUrl string
}

type ListOptions struct {
	Namespace     string            `json:"namespace"`
	Page          int64             `json:"page"`
	Limit         int64             `json:"limit"`
	Selector      map[string]string `json:"selector"`
	FieldSelector map[string]string `json:"field_selector"`
}

type HttpRequestArgs struct {
	Method  string
	Url     string
	Query   map[string]string
	Headers map[string]string
	Data    any
}
