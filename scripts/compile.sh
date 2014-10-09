#!/bin/bash
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

XC_ARCH=${XC_ARCH:-386 amd64}
XC_OS=${XC_OS:-darwin linux windows}

COMMIT=`git describe --always`

rm -rf pkg/
gox \
    -ldflags "-X main.GitCommit \"${COMMIT}\"" \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -output "pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"
