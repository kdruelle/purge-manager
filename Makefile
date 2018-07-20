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


VERSION="0.1.0"
BUILDTIME="$(shell date +'%Y-%m-%d')"


local:
	go build -ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILDTIME}"

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

