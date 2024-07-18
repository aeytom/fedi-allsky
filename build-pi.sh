#!/bin/bash

set -o pipefail -o nounset -o errexit

export GOARCH="arm"
# export GOBIN=""
# export GOEXE=""
# export GOFLAGS=""
export GOHOSTARCH="arm"
# export GOHOSTOS="linux"
export GOOS="linux"
# export GOPROXY=""
# export GORACE=""
# export GOTMPDIR=""
#export GCCGO="gccgo"
export GOARM="7,hardfloat"
export CC="arm-linux-gnueabihf-gcc"
export CXX="arm-linux-gnueabihf-g++"
export CGO_ENABLED="1"
#export GOMOD=""
# export CGO_CFLAGS="-g -O2"
# export CGO_CPPFLAGS=""
# export CGO_CXXFLAGS="-g -O2"
# export CGO_FFLAGS="-g -O2"
# export CGO_LDFLAGS="-g -O2"
# export PKG_CONFIG="pkg-config"

go build -v
ls -l fedi-allsky
#scp -p fedi-motion-control scam.wg:/fedi/
