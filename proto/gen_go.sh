#!/bin/bash
OUT_DIR=../../pkg/manabuf/
mkdir -p $OUT_DIR
OPTION="paths=source_relative"

shopt -s globstar nullglob
for f in **/**/*.proto
do
# FAILSAFE #
# * Check if "$f" FILE exists and is a regular file and then only copy it #
  if [[ "$f" != @(*"syllabus"*|*"options"*) ]];
  then
   protoc -I=./ \
    --go_out=$OPTION:$OUT_DIR \
    --go_opt=$OPTION \
    --go-grpc_out=require_unimplemented_servers=false,$OPTION:$OUT_DIR \
    --go-grpc_opt=require_unimplemented_servers=false,$OPTION \
    --grpc-gateway_out $OPTION:$OUT_DIR \
    --grpc-gateway_opt $OPTION \
    $f;
  fi
done
  