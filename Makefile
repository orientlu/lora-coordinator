.PHONY: build run clean server docker docker-run dev-requirements

PKG := $(shell go list ./...)
APPNAME=lora-coordinator
VERSION := $(shell git describe --always |sed -e "s/^v//")

build:
	@echo "Start build"
	@mkdir -p build
	go build $(GO_EXTRA_BUILD_ARGS) -ldflags "-s -w -X main.version=$(VERSION)" -o build/lora-coordinator cmd/lora-coordinator/main.go

run:
	@go run -ldflags "-X main.version=$(VERSION)" cmd/lora-coordinator/main.go

clean:
	@echo "Start clean"
	@rm -rf ./build

server:
	./build/lora-coordinator

docker:
	docker build -t $(APPNAME):$(VERSION) .
	docker image ls | grep $(APPNAME)

docker-run:
	docker run -it --rm --entrypoint "./$APPNAME" $(APPNAME):${VERSION}

dev-requirements:
	go install golang.org/x/lint/golint
	go install github.com/goreleaser/goreleaser
	go install github.com/goreleaser/nfpm
