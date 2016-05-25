.PHONY: help format test install update

GO_FILES = $(shell find . -type f -name "*.go")

help:     ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

format:  ## Format code to respect CS
	goimports -w $(GO_FILES)
	gofmt -l -w -s .
	go fix ./...
	go vet ./...

test:      ## Run backend tests
	go test ./...
	go vet ./...

install:  ## Install backend dependencies
	go get github.com/boltdb/bolt/...
	go get golang.org/x/tools/cmd/goimports
	go list -f '{{range .Imports}}{{.}} {{end}}' ./... | xargs go get -v
	go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs go get -v

update:  ## Update dependencies
	go get -u all
