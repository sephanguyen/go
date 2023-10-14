#!/bin/bash

# This script renders the helm manifest (similar to `helm template`).
# Usage:
# - Render manifest (default: local.manabie)
#       ./scripts/render.bash
# - Render manifest and output to file "manifest.yaml"
#      ./scripts/render.bash manifest.yaml
# - Render manifest for other env/org
#       ENV=prod ORG=jprep ./scripts/render.bash
set -eu

export ENV=${ENV:-local}
export ORG=${ORG:-manabie}

skaffold_file="skaffold2.local.yaml"
skaffold_profile=
output_file=
build_extra_args=""
render_extra_args="--offline=true --digest-source=tag"

# shellcheck source=../deployments/env.bash
. ./deployments/env.bash
CI=false setup_env_variables >/dev/null

cmdname=${0##*/}
usage() {
    cat <<EOF
Usage:
    $cmdname [output-file] [-f skaffold-file]
        -h | --help         Print help
        -f | --file         Skaffold file to use (default: $skaffold_file)
        -p | --profile      -p/--profile argument passed to skaffold
EOF
}

# process arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h | --help)
            usage
            exit 0
            ;;
        -f | --file)
            skaffold_file=$2
            shift 2
            ;;
        -p | --profile)
            skaffold_profile=$2
            shift 2
            ;;
        *)
            output_file=$1
            shift 1
            ;;
    esac
done

if [ -n "${output_file}" ]; then
    render_extra_args="$render_extra_args --output=${output_file}"
fi
bin_name="skaffold"
if [[ "${skaffold_file}" == "skaffold2"* ]]; then
  bin_name="skaffoldv2"
fi
if [ -n "${skaffold_profile}" ]; then
    build_extra_args="$build_extra_args --profile=${skaffold_profile}"
    render_extra_args="$render_extra_args --profile=${skaffold_profile}"
fi
$bin_name build -q --dry-run -f "$skaffold_file" $build_extra_args |
    $bin_name render -f "$skaffold_file" --build-artifacts=- $render_extra_args
