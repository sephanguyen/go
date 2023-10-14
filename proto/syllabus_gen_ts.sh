#!/bin/sh
OUT_DIR=../../pkg/manabuf-ts
mkdir -p $OUT_DIR
protoc -I=./ -I=/usr/local/include \
    --js_out=import_style=commonjs,binary:$OUT_DIR \
    --grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:$OUT_DIR \
    /usr/local/include/google/protobuf/*.proto options/*.proto syllabus/**/*.proto;
