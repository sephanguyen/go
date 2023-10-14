#!/bin/bash
OUT_DIR=../../pkg/manabuf-ts
mkdir -p $OUT_DIR

protoc -I=./ -I=/usr/local/include \
        --js_out=import_style=commonjs,binary:$OUT_DIR \
        --grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:$OUT_DIR \
        /usr/local/include/google/protobuf/*.proto;

shopt -s globstar nullglob
for f in **/**/*.proto
do
# FAILSAFE #
# * Check if "$f" FILE exists and is a regular file and then only copy it #
  if [[ "$f" != @(*"syllabus"*|*"options"*) ]];
  then
    protoc -I=./ -I=/usr/local/include \
        --js_out=import_style=commonjs,binary:$OUT_DIR \
        --grpc-web_out=import_style=commonjs+dts,mode=grpcwebtext:$OUT_DIR \
     $f;    
  fi
done
  