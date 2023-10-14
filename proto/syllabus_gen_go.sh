#!/bin/bash
OUT_DIR=../../pkg/manabuf/
mkdir -p $OUT_DIR
OPTION="paths=source_relative"
protoc -I=./ \
    -I=/usr/local/include \
    --go_out=$OPTION:$OUT_DIR \
    --go_opt=$OPTION \
    --go-grpc_out=require_unimplemented_servers=false,$OPTION:$OUT_DIR \
    --go-grpc_opt=require_unimplemented_servers=false,$OPTION \
    --grpc-gateway_out $OPTION:$OUT_DIR \
    --grpc-gateway_opt $OPTION \
    --validate_out=lang=go:$OUT_DIR \
    --validate_opt=$OPTION \
    --doc_out=/docs/html --doc_opt=html,index.html \
    --syllabus_out=$OPTION:$OUT_DIR \
    --syllabus_opt=$OPTION \
    ./syllabus/**/*.proto;

# chmod 777  $OUT_DIR**/**/transform
# chmod 777  $OUT_DIR**/**/transform/*

# package="package transform"
# template="package transform \n import ( sspb \"github.com\/manabie-com\/backend\/pkg\/manabuf\/syllabus\/v1\" \n \"github.com\/manabie-com\/backend\/internal\/eureka\/entities\" \n \"github.com\/manabie-com\/backend\/internal\/eureka\/golibs\/transformhelpers\")"

# shopt -s globstar nullglob
# for f in ../../pkg/manabuf/syllabus/**/transform/*.go
# do
#   if [[ "$f" != *"options"* ]];
#   then
#     sed -i "s/$package/$template/" $f
#   fi
# done

# gofumpt -l -w ../../pkg/manabuf/syllabus
