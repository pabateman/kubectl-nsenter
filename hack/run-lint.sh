#!/usr/bin/env bash

# Copyright 2020 Cornelius Weig
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

HACK=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
GOPATH="$(go env GOPATH)"
GOLANGCI_LINT_VERSION="v1.54.2"

if ! [[ -x "$GOPATH/bin/golangci-lint" ]]
then
   echo 'Installing golangci-lint'
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $GOPATH/bin $GOLANGCI_LINT_VERSION
fi

# TODO: add security check
"$GOPATH/bin/golangci-lint" run \
		--timeout 2m \
		--no-config \
		-D errcheck \
		-E goconst \
		-E gocritic \
		-E goimports \
		-E gosimple \
		-E misspell \
		-E unconvert \
		-E unparam \
		-E stylecheck \
		-E staticcheck \
		-E prealloc \
		--skip-dirs hack
