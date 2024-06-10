#!/usr/bin/env bash

# Imported from sigs.k8s.io/krew/hack/install-gox.sh

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail
[[ -n "${DEBUG:-}" ]] && set -x

scratch_dir="$(mktemp -d)"
cleanup() {
  rm -rf "${scratch_dir}"
}
trap cleanup EXIT

install_gox() {
  gobin="$(go env GOPATH)/bin"
  cd "${scratch_dir}"
  env GOPATH="${scratch_dir}" \
    GOBIN="${gobin}" \
    go install github.com/mitchellh/gox@9f712387e2d2c810d99040228f89ae5bb5dd21e5
}

ensure_gox() {
  command -v "gox" &>/dev/null
}

install_gox
if ! ensure_gox; then
  echo >&2 "gox not in PATH"
  exit 1
fi
