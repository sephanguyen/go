#!/bin/sh
OUT_DIR=../../pkg/manabuf-dart
mkdir -p $OUT_DIR
protoc -I=./ -I=/usr/local/include \
    --dart_out=grpc:$OUT_DIR \
    /usr/local/include/google/protobuf/*.proto **/**/**/*.proto **/**/*.proto;
chmod -R 777 "${OUT_DIR}"
