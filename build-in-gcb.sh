#!/bin/bash

set -eux

subs=()
if [[ -n "${_GO_IMAGE:-}" ]]; then
  subs+=("_GO_IMAGE=$_GO_IMAGE")
fi
# shellcheck disable=SC2178
subs="--substitutions=$(IFS=, eval 'echo "${subs[*]}"')"

gcloud container builds submit \
  "${subs:-}" \
  --config addon-builder.cloudbuild.yaml \
  .
