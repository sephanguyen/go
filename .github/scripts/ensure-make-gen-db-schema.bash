#!/bin/bash

git diff --exit-code
if [[ $? == 1 ]]; then
    >&2 echo "Repository has uncommitted changes. Validation failed."
    exit 1
fi

make gen-db-schema

git add --intent-to-add .
git diff --exit-code
if [[ $? == 1 ]]; then
    >&2 echo "Repository has new changes after make gen-db-schema. Please ensure make gen-db-schema has been run locally."
    exit 1
fi
