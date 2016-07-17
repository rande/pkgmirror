.PHONY: help format test install update build release assets

GO_BINDATA_PREFIX = $(shell pwd)/gui/build
GO_BINDATA_PATHS = $(shell pwd)/gui/build
GO_BINDATA_IGNORE = "(.*)\.(go|DS_Store)"
GO_BINDATA_OUTPUT = $(shell pwd)/assets/bindata.go
GO_BINDATA_PACKAGE = assets
GO_PROJECTS_PATHS = ./ ./test ./test/mirror ./test/tools ./test/api ./api ./assets ./cli ./mirror/composer ./mirror/git ./mirror/npm ./mirror/bower ./commands
GO_PKG = ./,./mirror/composer,./mirror/git,./mirror/npm,./mirror/bower,./api
GO_FILES = $(shell find $(GO_PROJECTS_PATHS) -maxdepth 1 -type f -name "*.go")
JS_FILES = $(shell find ./gui/src -type f -name "*.js")

SHA1=$(shell git rev-parse HEAD)

help:     ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

format-frontend:  ## Format code to respect CS
	./gui/node_modules/.bin/eslint --fix -c ./gui/.eslintrc $(JS_FILES)

format-backend:  ## Format code to respect CS
	goimports -w $(GO_FILES)
	gofmt -l -w -s $(GO_FILES)
	go fix $(GO_PROJECTS_PATHS)
	go vet $(GO_PROJECTS_PATHS)

coverage-backend: ## run coverage tests
	mkdir -p build/coverage
	rm -rf build/coverage/*.cov
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/main.cov ./
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/composer.cov ./mirror/composer
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/git.cov ./mirror/git
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/npm.cov ./mirror/npm
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/npm.cov ./mirror/bower
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/functional_mirror.cov ./test/mirror
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/functional_api.cov ./test/api
	go test -v -timeout 60s -coverpkg $(GO_PKG) -covermode count -coverprofile=build/coverage/functional_tools.cov ./test/tools
	gocovmerge build/coverage/* > build/pkgmirror.coverage
	go tool cover -html=./build/pkgmirror.coverage -o build/pkgmirror.html

test-backend:     ## Run backend tests
	go test -v -race -timeout 60s $(GO_PROJECTS_PATHS)
	go vet $(GO_PROJECTS_PATHS)

test-frontend:      ## Run frontend tests
	exit 0

test: test-backend test-front

run: bin-dev      ## Run server
	go run -race cli/main.go run -file ./pkgmirror.toml -log-level=info

install: install-backend install-frontend

install-backend:  ## Install backend dependencies
	go get github.com/wadey/gocovmerge
	go get golang.org/x/tools/cmd/cover
	go get github.com/aktau/github-release
	go get golang.org/x/tools/cmd/goimports
	go get -u github.com/jteeuwen/go-bindata/...
	go get github.com/Masterminds/glide
	glide install

install-frontend:  ## Install frontend dependencies
	cd gui && npm install

update:  ## Update dependencies
	glide update

bin-dev:                 ## Generate bin dev assets file
	go-bindata -dev -o $(GO_BINDATA_OUTPUT) -prefix $(GO_BINDATA_PREFIX) -pkg $(GO_BINDATA_PACKAGE) -ignore $(GO_BINDATA_IGNORE) $(GO_BINDATA_PATHS)

bin: assets                 ## Generate bin assets file
	go-bindata -o $(GO_BINDATA_OUTPUT) -prefix $(GO_BINDATA_PREFIX) -pkg $(GO_BINDATA_PACKAGE) -ignore $(GO_BINDATA_IGNORE) $(GO_BINDATA_PATHS)

assets:  ## build assets
	rm -rf gui/build/*
	cd gui && NODE_ENV=production node_modules/.bin/webpack --config webpack-production.config.js --progress --colors

watch:  ## build assets
	rm -rf gui/build/*
	cd gui && node_modules/.bin/webpack-dev-server --config webpack-dev-server.config.js --progress --inline --colors

build: bin ## build binaries
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/darwin/amd64/pkgmirror cli/main.go
	GOOS=linux  GOARCH=amd64 go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/amd64/pkgmirror  cli/main.go
	GOOS=linux  GOARCH=386   go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/386/pkgmirror    cli/main.go
	GOOS=linux  GOARCH=arm   go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/arm/pkgmirror    cli/main.go
	GOOS=linux  GOARCH=arm64 go build -ldflags "-X main.RefLog=$(SHA1) -s -w" -o build/linux/arm64/pkgmirror  cli/main.go