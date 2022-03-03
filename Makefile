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

export GO111MODULE ?= on
export CGO_ENABLED ?= 0

PROJECT   ?= kubectl-nsenter
REPOPATH  ?= github.com/pabateman/$(PROJECT)
COMMIT    := $(shell git rev-parse HEAD)
VERSION   ?= $(shell git describe --always --tags)
GOOS      ?= $(shell go env GOOS)
GOPATH    ?= $(shell go env GOPATH)

BUILDDIR   := $(shell pwd)/out
PLATFORMS  ?= darwin/amd64 darwin/arm64 linux/amd64
DISTFILE   := $(BUILDDIR)/$(PROJECT)-$(VERSION)-source.tar.gz
ASSETS     := $(BUILDDIR)/$(PROJECT)-darwin-amd64.tar.gz \
              $(BUILDDIR)/$(PROJECT)-darwin-arm64.tar.gz \
              $(BUILDDIR)/$(PROJECT)-linux-amd64.tar.gz

CHECKSUMS  := $(patsubst %, %.sha256, $(ASSETS))
DATE_FMT = %Y-%m-%dT%H:%M:%SZ
COMPRESS := gzip --best -k -c
define doUPX
	upx -9q $@
endef

ifdef SOURCE_DATE_EPOCH
    # GNU and BSD date require different options for a fixed date
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "+$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "+$(DATE_FMT)" 2>/dev/null)
else
    BUILD_DATE ?= $(shell date "+$(DATE_FMT)")
endif


.PHONY: help
help:
	@echo 'Valid make targets:'
	@echo '  - all:      build binaries for all supported platforms'
	@echo '  - build:    build binaries for all supported platforms'
	@echo '  - clean:    clean up build directory'
	@echo '  - deploy:   build artifacts for a new deployment'
	@echo '  - dev:      build the binary for the current platform'
	@echo '  - dist:     create a tar archive of the source code'
	@echo '  - help:     print this help'
	@echo '  - lint:     run golangci-lint'

$(BUILDDIR):
	mkdir -p "$@"

.PHONY: all
all: lint build deploy


.PHONY: dev
dev: CGO_ENABLED := 1
dev:
	go build -race -o $(PROJECT) cmd/$(PROJECT)/main.go

.PHONY: build
build: $(BUILDDIR)
	cd cmd/$(PROJECT) && \
	GOFLAGS="-trimpath" gox -osarch="$(PLATFORMS)" -output="$(BUILDDIR)/$(PROJECT)-{{.OS}}-{{.Arch}}" && \
	cd ../..

.PHONY: lint
lint:
	sh hack/run-lint.sh

.PRECIOUS: %.gz
%.gz: %
	$(COMPRESS) "$<" > "$@"

%.tar: %
	tar cf "$@" -C $(BUILDDIR) $(patsubst $(BUILDDIR)/%,%,$^)

%.sha256: %
	sha256sum  $< > $@

.INTERMEDIATE: $(DISTFILE:.gz=)
$(DISTFILE:.gz=): $(BUILDDIR)
	git archive --prefix="$(PROJECT)-$(VERSION)/" --format=tar HEAD > "$@"

.PHONY: deploy
deploy: $(CHECKSUMS)

.PHONY: dist
dist: $(DISTFILE)

.PHONY: clean
clean:
	$(RM) -r $(BUILDDIR) $(PROJECT)

$(BUILDDIR)/$(PROJECT)-darwin-arm64: build
$(BUILDDIR)/$(PROJECT)-darwin-amd64: build
$(BUILDDIR)/$(PROJECT)-linux-amd64: build
	$(doUPX)
