#!/bin/sh
OUT_DIR=../pkg/manabuf_py/
mkdir -p $OUT_DIR
OPTION="paths=source_relative"
python -m grpc_tools.protoc -I=./ \
    --python_out=$OUT_DIR \
    --grpc_python_out=$OUT_DIR \
    aphelios/vision/**/*.proto \
    scheduling/v1/*.proto \
    shamir/v1/*.proto;
