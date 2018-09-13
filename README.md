# k8s-addon-builder

## Using

This repository hosts code to build a Docker image (called "k8s-addon-builder") that
has both golang and docker installed. Depending on both Go and Docker is common
in the Kubernetes world of addon images.

It also ships with a command line utility called "ply" that can help with some common git and docker-based tasks.

## Building

The k8s-addon-builder image is based on gcr.io/cloud-builders/docker, but also
packages in a Golang compiler. To build the k8s-addon-builder with the Golang
compiler from gcr.io/cloud-builders/go:debian, run

```
./build-in-gcb.sh
```

To build with a different version of Go (instead of the one in
gcr.io/cloud-builders/go:debian), do


```
_GO_IMAGE=golang:1.9-stretch ./build-in-gcb.sh
```

The above works because gcr.io/cloud-builders/go:debian uses the official golang
Docker image as a base image.
