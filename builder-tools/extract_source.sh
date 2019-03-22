#!/bin/bash
#
# Copyright 2019 Google LLC
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

# Standalone script to extract files from a Docker image.
# Does not require the image to have a shell.
# Meant for extracting a source bundle and LICENSE file.

set -o nounset
set -o pipefail
set -o errexit

# Set defaults.
image=''
output_dir="$(pwd)"
clobber='false'
files=('source.tar.xz' 'LICENSE')

# Usage
function short_usage() {
  echo "Usage: $0 -i '<docker_image>' -o '<output_dir>' [-f 'file1' -f 'dir/file2']"
}
function usage() {
cat <<EOF
Extract source bundle and LICENSE file from a Docker image.

$(short_usage)
  -h                      Print usage.

  -i <image>              The full docker image, including the tag or digest.

  [-o <dir>]              The local directory to extract files to.
                          Default value: Current directory (${PWD}).

  [-c]                    Clobber existing files.
                          Default: false.

  [-f <filepath> -f ...]  Additional files to attempt to extract from the image.
                          Paths are always treated as starting at the root of the image's file system (/).
                          Globs (*) are not understood.
                          By default only looks for: ${files[@]}
Example:
  $0 -i 'gcr.io/k8s-image-staging/gke-mpi-metadata-server:76d1aec08eeeab7cdcd4bfe8591f65a35e7437e8' -o . -f 'NOTICES.txt' -f 'third_party/COPYRIGHTS' -c
EOF
}

# Get arguments.
OPTIND=1  # Reset in case getopts has been used previously in the shell.
while getopts ":h?i:o:cf:" opt; do
  case "$opt" in
    i ) image=$OPTARG ;;
    o ) output_dir=$OPTARG ;;
    c ) clobber='true' ;;
    f ) files+=("${OPTARG}") ;;

    h ) usage; exit 0 ;;
    \?) echo "Unknown option: -$OPTARG" >&2; exit 1 ;;
    : ) echo "Missing option argument for -$OPTARG" >&2; exit 1 ;;
    * ) echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
  esac
done
shift $((OPTIND-1))

# Validate options.
if [ -z "${image}" ] || [ -z "${output_dir}" ]; then
  short_usage
  exit 1
elif ! [ -d "${output_dir}" ]; then
  echo "Output dir does not exist: ${output_dir}" >&2
  exit 1
fi

# Create the container without running it.
# Add an empty CMD just in case it wasn't specified.
# Allow it to fail and print an error message if the image is not found.
container=$(docker create "${image}" '')

# Cleanup
trap "docker rm --volumes --force '${container}' >/dev/null" EXIT

# Copy files to the destination dir.
res=0
for file in "${files[@]}"; do
  dest_file="${output_dir}/${file}"

  # Create the same directory structure in the output dir.
  # Use `dirname` instead of just ${output_dir} because ${file} could be a nested file path.
  mkdir -p "$(dirname ${dest_file})"

  if [ -f "${dest_file}" ] ; then
    if [ "${clobber}" = 'true' ]; then
      echo "Replacing existing file: ${dest_file}"
    else
      echo "File ${dest_file} already exists and clobber is off, skipping."
      res=1
      continue
    fi
  fi

  # Attempt to copy all files, even if one fails.
  set +o errexit
  docker cp "${container}:/${file}" "${dest_file}"
  err=$?
  set -o errexit

  if ((err)); then
    res=1
    continue
  fi

  echo "Extracted file /${file} to ${dest_file}"
done

exit $res
