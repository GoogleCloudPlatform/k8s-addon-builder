# k8s-addon-builder

## Using

This repository hosts code to build a Docker image (called "k8s-addon-builder") that
has both golang and docker installed. Depending on both Go and Docker is common
in the Kubernetes world of addon images.

It also ships with a command line utility called "ply" that can help with some common git and docker-based tasks.

## Building

Use the included [`build.sh` script](./build.sh). This script uses environment variables to
adjust the build parameters. If the environment variables are not defined, it
uses default values. See [the script](./build.sh) for more details.
