#!/bin/bash

set -eu

cmd=${1:-"nocommand"}
case "$cmd" in
"version")
  echo "2.0.0"
  ;;
"wrongversion")
  echo "abcd.xyz"
  ;;
"die")
  kill -11 $$
  ;;
"nocommand")
  echo "Hello world"
  ;;
*)
  echo >&2 "Unrecognized command: $cmd"
  exit 1
  ;;
esac
