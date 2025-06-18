set -ex
go test -p 1 -coverprofile=.coverage.out -coverpkg=gcmdb/pkg/cmdb/... -count=1 ./...
go tool cover -html .coverage.out -o .coverage.html
go tool cover -func .coverage.out