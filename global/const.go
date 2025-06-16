package global

const StoragePathPrefix string = "/registry"

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
	// "AppInstance",
	// "AppInstanceRun",
	// "VirtualNetwork",
	// "Subnet",
	// "DatabaseService",
}
