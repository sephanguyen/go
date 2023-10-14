#!/bin/bash
set -eu

OUT_DIR=${OUT_DIR:-"../../pkg/manabuf-ts"}
PROTO_DIR=${PROTO_DIR:-"options/*.proto common/**/*.proto"}
TEAM=${TEAM:-""}

mkdir -p $OUT_DIR
protoc -I=./ -I=/usr/local/include \
    --js_out=import_style=commonjs,binary:$OUT_DIR \
    --grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:$OUT_DIR \
    /usr/local/include/google/protobuf/*.proto $PROTO_DIR;
chmod -R 777 "${OUT_DIR}"


# if team is syllabus, replace proto.* to proto.syllabus_*
grep -rl "proto." ${OUT_DIR} | xargs sed -i "s/proto\./proto\.${TEAM}_/g"

