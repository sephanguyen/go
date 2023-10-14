#!/bin/bash

set -eu

current_dir="${BASH_SOURCE%/*}"
SRC_DIR="${SRC_DIR:-"${current_dir}"}"
DST_DIR="${DST_DIR:-"${current_dir}/../camel/libs/proto/src/main/java/"}"

mkdir -p "${DST_DIR}"

# $ protoc --version
# libprotoc 23.4

# See https://github.com/grpc/grpc-java/tree/master/compiler on how to download
# protoc-gen-grpc-java and use it as a plugin

protoc -I="${SRC_DIR}" \
  --java_out="${DST_DIR}" \
  --plugin=protoc-gen-grpc-java="${current_dir}/../protoc-gen-grpc-java-1.57.0-linux-x86_64.exe" \
  --grpc-java_out="${DST_DIR}" \
  "${SRC_DIR}/zeus/v1/log.proto"
