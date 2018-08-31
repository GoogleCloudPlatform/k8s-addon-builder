The addon-builder image is based on gcr.io/cloud-builders/docker, but also
packages in a Golang compiler. To build the addon-builder with the Golang
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
