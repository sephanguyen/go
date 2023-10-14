#!/bin/bash

set -eu

export CI=${CI:-false}
if [ "$CI" == "false"  ]; then
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./build/gendbschema  ./cmd/citools/dbschema 
fi

docker compose -f ./developments/dbschema.docker-compose.yml up --abort-on-container-exit --build; \
docker compose -f ./developments/dbschema.docker-compose.yml down --volumes
