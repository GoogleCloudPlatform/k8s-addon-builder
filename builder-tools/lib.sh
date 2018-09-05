#!/bin/bash
#
# Copyright 2018 Google LLC
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

set -ex

# Print docker images that match a regex.
find_images()
{
  local images_found
  regex="$1"
  images_found=$(docker images --format "{{.Repository}}:{{.Tag}}" | grep "${regex}")
  if [[ -z "${images_found}" ]]; then
    echo >&2 "no images found for regex ${regex}"
    return 1
  fi
  echo "${images_found}"
}

# Create a new Docker image from an existing one by adding metadata via LABELs.
mk_labeled_image()
{
  if (( $# < 4 )); then
    echo >&2 "usage: mk_labeled_image TMPDIR IMAGE_OLD IMAGE_NEW LABEL..."
    return 1
  fi
  local tmpdir="$1"
  local image_old="$2"
  local image_new="$3"
  shift 3

  # Handle spaces inside LABEL values.
  local labels=()
  for label in "$@"; do
    labels+=("--label" "${label}")
  done

  docker_build_dir="$(mktemp -d -p "${tmpdir}")"
  pushd "${docker_build_dir}"

  # Use interim tag to guarantee that we are not pulling from
  # gcr.io/google-containers. Use the $BUILD_ID that comes with GCB's
  # environment for a free UUID.
  image_local="$BUILD_ID:local"
  docker tag "${image_old}" "${image_local}"
  echo "FROM ${image_local}" > Dockerfile

  docker build "${labels[@]}" -t "${image_new}" .
}

# Label found images with metadata, then push them.
label_and_push_images()
{
  if (( $# < 4 )); then
    echo >&2 "usage: label_and_push_images TMPDIR REGEX PUSH_REGISTRY LABEL1 LABEL2 ..."
    return 1
  fi
  local regex tmpdir labels
  tmpdir="$1"
  regex="$2"
  push_registry="$3"
  if [[ -z "$regex" ]]; then
    echo >&2 "regex cannot be empty"
    return 1
  fi
  shift 3
  labels=("$@")
  images_found=$(find_images "${regex}")
  for image in $images_found; do
    registry_old="$(extract_registry "${image}")"
    if [[ -n "${registry_old}" ]]; then
      image_new="${image/${registry_old}/${push_registry}}"
    else
      # Handle the case where the image is named <image> instead of
      # <registry>/<image>.
      image_new="${push_registry}/${image}"
    fi
    mk_labeled_image "${tmpdir}" "${image}" "${image_new}" "${labels[@]}"
    docker push "${image_new}"
  done
}

label_and_push_images_gcb()
{
  if (( $# < 1 )); then
    echo >&2 "usage: label_and_push_images_gcb REGEX LABEL1 LABEL2 ..."
    return 1
  fi
  if [[ -z "${BUILD_ID}" ]]; then
    echo >&2 "\$BUILD_ID must be defined"
    return 1
  elif [[ -z "${PROJECT_ID}" ]]; then
    echo >&2 "\$PROJECT_ID must be defined"
    return 1
  fi
  local regex=""
  local push_registry="gcr.io/$PROJECT_ID"
  regex="$1"
  shift
  if [[ -n "${PUSH_REGISTRY:-}" ]]; then
    push_registry="${PUSH_REGISTRY}"
  fi
  label_and_push_images /workspace "$regex" "${push_registry}" \
    GCB_BUILD_ID="$BUILD_ID" \
    GCB_PROJECT_ID="$PROJECT_ID" \
    "$@"
}

# Extract registry name from docker image string.
#   gcr.io/google-containers/etcd:tag => gcr.io/google-containers
#   k8s.gcr.io/etcd:tag => k8s.gcr.io
extract_registry()
{
  if [[ "$1" =~ "/" ]]; then
    echo "$1" | rev | sed 's|[^/]\+\?/||' | rev
  fi
}

# For all images that match REGEX, set their registry to REGISTRY.
set_registry()
{
  if (( $# != 2 )); then
    echo >&2 "usage: set_registry REGISTRY REGEX"
    return 1
  fi
  local push_registry=$1
  local regex=$2
  images_found=$(find_images "${regex}")
  for image in $images_found; do
    registry_old="$(extract_registry "${image}")"
    if [[ -n "${registry_old}" ]]; then
      image_new="${image//${registry_old}/${push_registry}}"
    else
      image_new="${push_registry}/${image}"
    fi
    # Give the image a new name.
    docker tag "${image}" "${image_new}"
    # Remove the old image (so that only the new image name remains).
    docker rmi "${image}"
  done
}

# Get the tag of a full image, but do not treat "latest" as a tag.
get_docker_tag()
{
  tag=$(echo "$1" | awk -F: '{print $2}')
  if [[ "${tag}" != "latest" ]]; then
    echo "${tag}"
  fi
}

# Get everything except the tag (if any).
get_registry_and_name()
{
  tag=$(echo "$1" | awk -F: '{print $1}')
  echo "${tag}"
}

# For all images that match REGEX, either SET a tag or APPEND a suffix to it. If
# the TAG starts with a "-", "_", or ".", treat it as a suffix; otherwise just
# SET it. Abort if the resulting tag would exceed 128 characters.
set_docker_tag()
{
  if (( $# != 2 )); then
    echo >&2 "usage: set_docker_tag TAG REGEX"
    return 1
  fi
  local desired_tag=$1
  local regex=$2
  local append=0
  local images_found
  local append
  local existing_tag
  local new_tag
  if [[ "${desired_tag}" =~ ^[-_.] ]]; then
    append=1
  fi
  images_found=$(find_images "${regex}")
  for image in $images_found; do
    existing_tag=$(get_docker_tag image)
    if [[ -n "${existing_tag}" ]]; then
      if ((append)); then
        # Append the existing tag.
        new_tag="${existing_tag}${desired_tag}"
      else
        # Overwrite existing tag.
        new_tag="${desired_tag}"
      fi
    else # No tag exists (i.e., it is ":latest").
      if ((append)); then
        echo >&2 "cannot append tag ${desired_tag} to image ${image} (tag must start with a letter or number)"
        return 1
      else
        new_tag="${desired_tag}"
      fi
    fi

    registry_and_name="$(get_registry_and_name "${image}")"
    docker tag "${image}" "${registry_and_name}:${new_tag}"
    docker rmi "${image}"
  done
}
