#!/bin/bash

set -eu

current_manifest_filepath="current-$ENV-$ORG-$1-manifest.yaml"
new_manifest_filepath="new-$ENV-$ORG-$1-manifest.yaml"
elastic_namespace="$ENV-$ORG-elastic"

render_and_diff() {
    skaffold render -f skaffold.$1.yaml > $new_manifest_filepath
    git diff --no-index $current_manifest_filepath $new_manifest_filepath || true
}

case $1 in

  backbone)
    { helm -n $ENV-$ORG-nats-jetstream get manifest nats-jetstream ; helm -n $elastic_namespace get manifest elastic ; helm -n $ENV-$ORG-kafka get manifest kafka ; helm -n $ENV-$ORG-kafka get manifest cp-schema-registry ; } > $current_manifest_filepath
    render_and_diff $1
    ;;
  elastic)
    { helm -n $elastic_namespace get manifest elastic; } > "${current_manifest_filepath}"
    skaffold render -f skaffold.backbone.yaml -p elastic-only > "${new_manifest_filepath}"
    git diff --no-index "${current_manifest_filepath}" "${new_manifest_filepath}" || true
    ;;
  kafka)
    { helm -n $ENV-$ORG-kafka get manifest kafka; } > $current_manifest_filepath
    skaffold render -f skaffold.backbone.yaml -p kafka-only > $new_manifest_filepath
    git diff --no-index $current_manifest_filepath $new_manifest_filepath || true
    ;;
  kafka-connect)
    { helm -n $ENV-$ORG-kafka get manifest kafka-connect; } > $current_manifest_filepath
    skaffold render -f skaffold.backbone.yaml -p kafka-connect-only > $new_manifest_filepath
    git diff --no-index $current_manifest_filepath $new_manifest_filepath || true
    ;;
  monitoring)
    releaseNames=$( helm -n monitoring ls -a -o json | jq '.[] | .name ' )
    for name in $releaseNames
    do
        eval "helm -n monitoring get manifest $name" >> $current_manifest_filepath
    done ;
    render_and_diff $1
    ;;
  *)
    echo -n "unknown"
    exit 1
    ;;
esac
