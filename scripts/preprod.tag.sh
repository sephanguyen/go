#!/bin/bash

set -eux

cur_date=$(date '+%Y%m%d000000')
cur_sha=$(git rev-parse --short HEAD)
cur_branch=$(git rev-parse --abbrev-ref HEAD)
new_tag="${cur_date}.${cur_sha}"

gh release create "${new_tag}" --prerelease --target="${cur_branch}" --generate-notes
