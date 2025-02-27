################################################################################
#
# This file is part of purge-manager.
#
# (C) 2011 Kevin Druelle <kevin@druelle.info>
#
# This software is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
# 
# This software is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with this software.  If not, see <http://www.gnu.org/licenses/>.
# 
################################################################################


VERSION=$(shell git describe --tags)
BUILDTIME=$(shell LANG=en_US.UTF-8; date +'%d %b %Y')

TARGETS="linux/amd64,linux/386,darwin/amd64,darwin/386,windows/amd64,windows/386"

BIN = linux-amd64.bin \
	  windows-386.bin \
	  linux-386.bin \
	  darwin-amd64.bin \
	  darwin-386.bin \
	  windows-amd64.bin 

ARTIFACTS = linux-amd64.zip \
			linux-amd64.deb \
			linux-386.zip \
			linux-386.deb \
			darwin-amd64.zip \
			darwin-386.zip \
			windows-amd64.zip \
			windows-386.zip

SRC = $(wildcard *.go)

DIST_DIR = dist
BIN_DIR  = bin

BINTARGETS = $(addprefix ${BIN_DIR}/purge-manager-${VERSION}-, ${BIN})

RELEASES = $(addprefix ${DIST_DIR}/purge-manager-${VERSION}-, ${ARTIFACTS})

SHELL := /bin/bash

.PHONY: all
all: $(BINTARGETS)

ci: $(RELEASES)

$(BINTARGETS) : $(SRC)
	$(eval OS   = $(shell echo $@ | sed -E 's/.*-([a-z]+)-[a-z0-9]+\.bin(\.exe)?/\1/g'))
	$(eval ARCH = $(shell echo $@ | sed -E 's/.*-([a-z0-9]+)\.bin(\.exe)?/\1/g'))
	@mkdir -p build
	@mkdir -p $(BIN_DIR)
	@printf "building %s/%s.... " ${OS} ${ARCH}
	@export GOPATH="$(shell go env GOPATH)"; \
		xgo -targets "${OS}/${ARCH}" -ldflags '-s -w -X "main.version=${VERSION}" -X "main.buildTime=${BUILDTIME}"' -dest ./build -out purge-manager-${VERSION} ./ > /dev/null
	@find ./build -name "*-${OS}*${ARCH}*" | sed -E 'p;s/${OS}-.*/${OS}-${ARCH}.bin/g; s/build/bin/g' | xargs -n2 cp
	@printf "done.\n"

$(filter %.zip, $(RELEASES)) : dist/%.zip : bin/%.bin
	@mkdir -p dist
	@printf "creating %s.... " $(notdir $@)
	@if [[ "$<" =~ "windows" ]]; then \
		cp $< dist/purge-manager.exe ; \
		cd dist && zip $(notdir $@) purge-manager.exe > /dev/null && rm purge-manager.exe && cd ../ ; \
	else \
		cp $< dist/purge-manager ; \
		cd dist && zip $(notdir $@) purge-manager > /dev/null && rm purge-manager && cd ../ ; \
	fi
	@printf "done.\n"

$(filter %.deb, $(RELEASES)) : dist/%.deb : bin/%.bin
	$(eval OS = $(shell uname))
	@mkdir -p dist
	@printf "creating %s.... " $(notdir $@)
	@if [[ "${OS}" == "Darwin" ]]; then \
		export VERSION=${VERSION}; \
		sed -i '' 's#.*/usr/local/bin/purge-manager.*#    ./$<: "/usr/local/bin/purge-manager"#g' nfpm.yaml ; \
		sed -i '' 's#.*arch.*#arch: "$(shell echo $< | sed -E 's/.*-([a-z0-9]+).bin$$/\1/g')"#g' nfpm.yaml ; \
		nfpm pkg -t $@ > /dev/null ; \
	else \
		export VERSION=${VERSION}; \
		sed -i 's#.*/usr/local/bin/purge-manager.*#    ./$<: "/usr/local/bin/purge-manager"#g' nfpm.yaml ; \
		sed -i 's#.*arch.*#arch: "$(shell echo $< | sed -E 's/.*-([a-z0-9]+).bin$$/\1/g')"#g' nfpm.yaml ; \
		nfpm pkg -t $@ > /dev/null ; \
	fi
	@printf "done."



local:
	go build -ldflags '-X "purge-manager/command.Version=${VERSION}" -X "purge-manager/command.BuildTime=${BUILDTIME}"'

releases:
	export GOPATH="$(shell go env GOPATH)"; \
	xgo -targets linux/amd64,linux/386,darwin/amd64,darwin/386,windows/amd64,windows/386 \
	-out purge-manager \
	-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILDTIME}" \
	-out purge-manager -dest ./bin/ ./

linux:
	export GOPATH="$(shell go env GOPATH)"; \
	xgo -targets linux/amd64 -out purge-manager \
	-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILDTIME}" \
	-dest ./bin/ ./

