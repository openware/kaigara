#!/bin/bash

set -x

platforms=("darwin/amd64" "darwin/arm64" "linux/amd64" "windows/amd64")
sources=("kaigara" "kai")

VERSION=""
if [ ! -z "${KAIGARA_VERSION}" ]; then
    VERSION="-X main.Version=${KAIGARA_VERSION}"
fi

for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_os="${GOOS}_${GOARCH}"
    if [ ${GOOS} = "windows" ]; then
        output_os+='.exe'
    fi

    for src in ${sources[@]}; do
        GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -a -ldflags "-w ${VERSION}" -o bin/${src}_${output_os} ./cmd/${src}
        if [ $? -ne 0 ]; then
            echo "An error has occurred while building ${src}! Aborting the script execution..."
            exit 1
        fi
    done
done
