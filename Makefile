GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOPATH?=$(HOME)/go
BINARY_PATH_COLA=$(GOPATH)/src/github.com/GoogleCloudPlatform/addon-builder/cmd/cola
REGISTRY?=gcr.io/gke-release-staging

all: test build
cola:
	cd cmd && $(GOBUILD) -o $(BINARY_PATH_COLA) -v cola.go
cola-static:
	cd cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a \
		-ldflags "-X main.VersionDate=`date -u +%Y-%m-%dT%I:%M:%S%z` -X main.VersionGit=`git describe --always --dirty --long` -s" \
		-o $(BINARY_PATH_COLA) \
		-v \
		cola.go
build: cola
build-static: cola-static
docker-image:
	docker build -t $(REGISTRY)/addon-builder-tools:latest .
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH_COLA)
deps:
	$(GOGET) github.com/golang/dep/cmd/dep
	dep ensure
