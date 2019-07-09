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

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

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
    if [[ "${image}" == "${image_new}" ]]; then
      continue
    fi
    # Give the image a new name.
    docker tag "${image}" "${image_new}"
    # Remove the old image (so that only the new image name remains).
    docker rmi "${image}"
  done
}

get_docker_tags()
{
  local auth
  local token

  local domain
  local img_path_and_tag
  local img_path

  token="$1"

  auth="Authorization: Bearer ${token}"

  domain="${2%%/*}"
  img_path_and_tag="${2#*/}"
  img_path="${img_path_and_tag%:*}"

  # Get docker tags.
  curl -v -fs -H "$auth" "https://${domain}/v2/${img_path}/tags/list" | jq -r '."tags"[]'
}

mk_next_tag_suffix_version()
{
  local token
  local full_image_name
  local tag_suffix_regex

  local domain
  local img_path_and_tag
  local img_path

  token="${1}"
  full_image_name="${2}"

  if [[ ! "${full_image_name}" =~ .+:.+$ ]]; then
    echo >&2 "image name must have a tag"
    exit 1
  fi
  tag_suffix_regex="${3}"
  domain="${2%%/*}"
  img_path_and_tag="${2#*/}"
  img_path="${img_path_and_tag%:*}"
  shift
  shift

  local remote_tags
  remote_tags="$(get_docker_tags "${token}" "${full_image_name}")"

  local remote_tag_suffix_version
  local tag_suffix_versions_found=()
  local next_tag_suffix_version=0

  local remote_image
  for remote_tag in ${remote_tags}; do
    remote_image="${domain}/${img_path}:${remote_tag}"
    local_image_regex="${full_image_name}${tag_suffix_regex}"
    if echo "${remote_image}" | grep -Fq -- "${local_image_regex}"; then
      remote_tag_suffix_version="${remote_image#"$local_image_regex"}"
      # Populate tag_suffix_versions_found with actual version numbers.
      if [[ "${remote_tag_suffix_version}" =~ ^[0-9]+$ ]]; then
        tag_suffix_versions_found+=("${remote_tag_suffix_version}")
      fi
    fi
  done

  # Find the max version (if any) and bump it.
  if (( ${#tag_suffix_versions_found[@]} )); then
    local max
    max=$(printf "%s\n" "${tag_suffix_versions_found[@]}" | sort -nr | head -n1)
    next_tag_suffix_version=$((max+1))
  fi

  echo "${next_tag_suffix_version}"
}

set_docker_tag_unique()
{
  local token="${1}"
  local pushRegex="${2}"
  local tag_suffix="${3}"
  images_found=$(find_images "${pushRegex}")
  # Handle each image on a case-by-case basis.
  for image in ${images_found}; do
    ver=$(mk_next_tag_suffix_version "${token}" "${image}" "${tag_suffix}")
    docker tag "${image}" "${image}${tag_suffix}${ver}"
    docker rmi "${image}"
  done
}
