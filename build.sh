#!/bin/bash
#
# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Build the addon-builder container. This script uses several environment
# variables (see uses of set_var() below) if they exist; otherwise, default
# values are used.

# set_var takes 2 arguments:
#
#   1) the name of the variable to create and use in this script, and
#   2) the default value to use if there is no environment variable of the same
#   name.
set_var() {
    if (( $# != 2 )); then
        return 1
    fi

    export "${1}"="${!1:-$2}"
}

# Set default values.
set_var REGISTRY "gcr.io/gke-release-staging"
set_var GO_IMAGE "golang"
set_var GO_VERSION "1.19"
set_var KO_VERSION "0.8.3"
set_var PLY_VERSION_GIT "$(git describe --always --dirty --long)"
set_var PLY_VERSION_DATE "$(date -u +%Y-%m-%dT%I:%M:%S%z)"

docker build --pull \
    --build-arg "GO_IMAGE=${GO_IMAGE}" \
    --build-arg "GO_VERSION=${GO_VERSION}" \
    --build-arg "KO_VERSION=${KO_VERSION}" \
    --build-arg "PLY_VERSION_GIT=${PLY_VERSION_GIT}" \
    --build-arg "PLY_VERSION_DATE=${PLY_VERSION_DATE}" \
    --label "GOLANG=${GO_VERSION}" \
    -t "${REGISTRY}/addon-builder:${GO_VERSION}" .
