
all: k8s-node-watcher

fmt:
	go fmt -x ./...
	go mod tidy

k8s-node-watcher: fmt
	go build -v ./...
	go build -v

clean:
	go clean -v

snapshot: clean
	goreleaser build --rm-dist --snapshot
