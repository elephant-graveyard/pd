# Copyright © 2020 The Homeport Team
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

version := $(shell git describe --tags --abbrev=0 2>/dev/null || (git rev-parse HEAD | cut -c-8))
sources := $(wildcard */*/*.go)
goos := $(shell uname | tr '[:upper:]' '[:lower:]')
goarch := $(shell uname -m | sed 's/x86_64/amd64/')

.PHONY: all
all: clean verify build

.PHONY: clean
clean:
	@GO111MODULE=on go clean -cache $(shell go list ./...)
	@rm -rf dist

.PHONY: verify
verify:
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod verify

.PHONY: install
install:
	@GO111MODULE=on CGO_ENABLED=0 GOOS=$(goos) GOARCH=$(goarch) go build \
		-tags netgo \
		-ldflags='-s -w -extldflags "-static" -X github.com/homeport/yft/internal/cmd.version=$(version)' \
		-o /usr/local/bin/pd \
		cmd/pd/main.go
