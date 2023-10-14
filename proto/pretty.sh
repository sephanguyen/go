#!/bin/sh
#this script mean to be run inside the protoc_builder container
for x in ./*/*/*.proto; do
    clang-format -style=google -i $x
done