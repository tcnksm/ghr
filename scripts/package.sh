#!/bin/bash
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

make bootstrap

VERSION=$(grep "const Version " version.go | sed -E 's/.*"(.+)"$/\1/')
COMMIT=$(git describe --always)

-d pkg && rm -rf ./pkg

goxz -pv=${VERSION} -build-ldflags="-X main.GitCommit=${COMMIT}" \
    -arch=386,amd64 -d=./pkg/dist/${VERSION}

# Generate shasum
pushd ./pkg/dist/${VERSION}
shasum -a 256 * > ./${VERSION}_SHASUMS
popd
