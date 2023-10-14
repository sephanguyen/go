#!/bin/bash

set -e

for dir in ./manabie-all-in-one/charts/*; do
    echo "pwd: $dir"
    cd $dir
    helm dependency update
    cd ../../..
done

for dir in ./platforms/nats-jetstream; do
    echo "pwd: $dir"
    cd $dir
    helm dependency update
    cd ../..
done
