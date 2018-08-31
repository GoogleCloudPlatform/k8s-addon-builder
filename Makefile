GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOPATH?=$(HOME)/go
BINARY_PATH_COLA=$(GOPATH)/src/github.com/GoogleCloudPlatform/addon-builder/ply
REGISTRY?=gcr.io/gke-release-staging
VERSION_GIT=$(shell git describe --always --dirty --long)
VERSION_DATE=$(shell date -u +%Y-%m-%dT%I:%M:%S%z)
LDFLAGS=-X github.com/GoogleCloudPlatform/addon-builder/cmd.VersionDate=${VERSION_DATE}
LDFLAGS+=-X github.com/GoogleCloudPlatform/addon-builder/cmd.VersionGit=${VERSION_GIT}
LDFLAGS+=-s

all: test build
build:
	$(GOBUILD) \
	-ldflags "$(LDFLAGS)" \
	-o $(BINARY_PATH_COLA) -v main.go
build-static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a \
		-ldflags "$(LDFLAGS)" \
		-o $(BINARY_PATH_COLA) \
		-v \
		main.go
docker-image:
	docker build -t $(REGISTRY)/addon-builder:latest .
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH_COLA)
deps:
	$(GOGET) github.com/golang/dep/cmd/dep
	dep ensure
