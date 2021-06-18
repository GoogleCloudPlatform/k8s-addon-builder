GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOPATH?=$(HOME)/go
BINARY_PATH_PLY=$(GOPATH)/src/github.com/GoogleCloudPlatform/k8s-addon-builder/ply
REGISTRY?=gcr.io/gke-release-staging
PLY_VERSION_GIT?=$(shell git describe --always --dirty --long)
PLY_VERSION_DATE?=$(shell date -u +%Y-%m-%dT%I:%M:%S%z)
LDFLAGS=-X github.com/GoogleCloudPlatform/k8s-addon-builder/cmd.VersionDate=$(PLY_VERSION_DATE)
LDFLAGS+=-X github.com/GoogleCloudPlatform/k8s-addon-builder/cmd.VersionGit=$(PLY_VERSION_GIT)
LDFLAGS+=-s

all: test build
build:
	$(GOBUILD) \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY_PATH_PLY) -v main.go
build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a \
		-ldflags "$(LDFLAGS)" \
		-o $(BINARY_PATH_PLY) \
		-v \
		main.go
docker-image:
	docker build -t $(REGISTRY)/k8s-addon-builder:latest .
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH_PLY)
deps:
	$(GOGET) github.com/golang/dep/cmd/dep
	dep ensure
