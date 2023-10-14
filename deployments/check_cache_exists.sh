#!/bin/bash

set -eu

echo "is-cache-exist=false" >> $GITHUB_OUTPUT
echo $@

# Check any cache that exists in the list of directories
for i in $@
do
    echo $i
    if [ -d "$i" ]
    then
        if [ "$(ls -A $i)" ]; then
            echo "Take action $i is not Empty"
            echo "is-cache-exist=true" >> $GITHUB_OUTPUT
        else
            echo "$i is Empty"
        fi
    else
        echo "Directory $i not found."
    fi
done