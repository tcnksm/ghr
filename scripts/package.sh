#!/bin/bash
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

-d pkg && rm -rf ./pkg
make crossbuild

VERSION=$(gobump show -r)

# Generate shasum
pushd ./pkg/dist/v${VERSION}
shasum -a 256 * > ./v${VERSION}_SHASUMS
popd
