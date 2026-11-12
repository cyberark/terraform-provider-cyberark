#!/bin/bash
# Sourced by build scripts. Exports the Go module proxy settings used by the
# docker builds/runs (go_proxy_docker_build, go_proxy_docker_run), when Jenkins
# has provided them:
#   /tmp/go-proxy-<ns>  GOPROXY URL (no credentials)
#   /tmp/go-netrc-<ns>  .netrc with the proxy credentials
# <ns> is a checksum of the workspace path (as computed in the Jenkinsfile), so
# concurrent builds on one machine don't overwrite each other's files.
# Without them (OSS contributors, local runs) the defaults are left untouched
# and Go uses proxy.golang.org.

GO_PROXY_BUILD_ARGS=()
GO_PROXY_RUN_ARGS=()

_go_proxy_ns="$(printf %s "$(git rev-parse --show-toplevel)" | cksum | cut -d' ' -f1)"
if [ -f "/tmp/go-proxy-${_go_proxy_ns}" ] && [ -f "/tmp/go-netrc-${_go_proxy_ns}" ]; then
  GOPROXY="$(cat "/tmp/go-proxy-${_go_proxy_ns}")"
  export GOPROXY
  export GO_NETRC_FILE="/tmp/go-netrc-${_go_proxy_ns}"
  export BUILDX_BAKE_ENTITLEMENTS_FS=0
  GO_PROXY_BUILD_ARGS=(--build-arg GOPROXY --secret "id=gonetrc,src=${GO_NETRC_FILE}")
  GO_PROXY_RUN_ARGS=(-e GOPROXY -e NETRC=/netrc -v "${GO_NETRC_FILE}:/netrc:ro")
  echo "Using the internal Go module proxy"
fi

go_proxy_docker_build() {
  docker build ${GO_PROXY_BUILD_ARGS[@]+"${GO_PROXY_BUILD_ARGS[@]}"} "$@"
}

go_proxy_docker_run() {
  docker run ${GO_PROXY_RUN_ARGS[@]+"${GO_PROXY_RUN_ARGS[@]}"} "$@"
}
