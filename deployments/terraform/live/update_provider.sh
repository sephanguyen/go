#!/bin/bash

for d in ./*/*platforms*/; do
  echo $d
  (cd $d && terragrunt init -upgrade)
done
