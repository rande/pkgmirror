.PHONY: help format test install update build release

GO_FILES = $(shell find . -type f -name "*.go")
SHA1=$(shell git rev-parse HEAD)

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

build: ## build binaries
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/darwin/amd64/pkgmirror cli/main.go
	GOOS=linux  GOARCH=amd64 go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/amd64/pkgmirror  cli/main.go
	GOOS=linux  GOARCH=386   go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/386/pkgmirror    cli/main.go
	GOOS=linux  GOARCH=arm   go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/arm/pkgmirror    cli/main.go
	GOOS=linux  GOARCH=arm64 go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/arm64/pkgmirror  cli/main.go

release: build ## build and release binaries on github
	github-release delete  --tag master --user rande --repo pkgmirror|| exit 0
	github-release release --tag master --user rande --repo pkgmirror --name "Beta release" --pre-release
	github-release upload  --tag master --user rande --repo pkgmirror --name "pkgmirror-osx-amd64"   --file build/darwin/amd64/pkgmirror
	github-release upload  --tag master --user rande --repo pkgmirror --name "pkgmirror-linux-amd64" --file build/linux/amd64/pkgmirror
	github-release upload  --tag master --user rande --repo pkgmirror --name "pkgmirror-linux-386"   --file build/linux/386/pkgmirror
	github-release upload  --tag master --user rande --repo pkgmirror --name "pkgmirror-linux-arm"   --file build/linux/arm/pkgmirror
	github-release upload  --tag master --user rande --repo pkgmirror --name "pkgmirror-linux-arm64" --file build/linux/arm64/pkgmirror
