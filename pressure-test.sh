set -e
cmdbAPI="http://localhost:3333/api/v1"
# cmdbAPI="http://localhost:8080/api/v1"
pressureTest() {
    bombardier -c 125 -n 10000 "$cmdbAPI/${1}s/"
}

pressureTestNamespaced() {
    bombardier -c 125 -n 10000 "$cmdbAPI/${1}s/${2}"
}

kinds=(
	"secret"
	"project"
	"datacenter"
	"zone"
	"namespace"
	"scm"
	"hostnode"
	"helmrepository"
	"containerregistry"
	"app"
	"configcenter"
	"deployplatform"
	"orchestration"
)

namespacedKinds=(
	"deploytemplate"
	"resourcerange"
	"appdeployment"
	"appinstance"
)

for kind in ${kinds[@]};
do
    pressureTest $kind
done

for kind in ${namespacedKinds[@]};
do
    pressureTestNamespaced $kind test
done