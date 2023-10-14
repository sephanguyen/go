#!/bin/bash

# This script contains helper functions for pretty loggings.
# To use this in your script, add `source ./scripts/log.sh`.

RED='\033[0;31m'
GRE='\033[0;32m'
YEL='\033[1;33m'
BLU='\033[0;34m'
NC='\033[0m'

DEBUG=${DEBUG:-false}  # toggle logdebug()

CI=${CI:-false} # whether running on Github Action
if [[ "$CI" == "true" ]]; then
  DEBUG=true
fi

logfatal() {
  echo >&2 -e "${RED}fatal: $*${NC}"
  exit 1
}

logerror() {
  echo >&2 -e "${RED}error: $*${NC}"
}

logwarn() {
  echo >&2 -e "${YEL}warn: $*${NC}"
}

logsuccess() {
  echo -e "${GRE}success: $*${NC}"
}

loginfo() {
  echo -e "${BLU}$*${NC}"
}

logdebug() {
  if [[ "$DEBUG" == "true" ]]; then
    echo "$*"
  fi
}
