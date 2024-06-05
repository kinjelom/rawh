#!/bin/bash

VERSION_LABEL="0.0.1"

DIST_DIR=".dist"

declare -a os_array=("linux" "darwin" "windows")
declare -a arch_array=("amd64" "arm64")

mkdir -p "$DIST_DIR"

for OS in "${os_array[@]}"; do
    for ARCH in "${arch_array[@]}"; do
        DIST_PATH="$DIST_DIR/rawh-v$VERSION_LABEL-$OS-$ARCH"
        echo "build $DIST_PATH"
        if GOOS=$OS GOARCH=$ARCH go build -ldflags "-X 'main.version=$VERSION_LABEL'" -o "$DIST_PATH" >> "$DIST_PATH.log"; then
            sha256sum "$DIST_PATH"  | awk '{print $1}' > "$DIST_PATH.sum"
            if [ ! -s "$DIST_PATH.log" ]; then
                rm "$DIST_PATH.log"
            fi
        fi
    done
done

