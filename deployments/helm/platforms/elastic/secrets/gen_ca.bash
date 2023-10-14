#!/bin/bash
# ORG=manabie ENV=staging ./gen_ca.bash
mkdir -p ./$ORG/$ENV

openssl genrsa -out ./$ORG/$ENV/root-ca-key.pem 2048
openssl req -new -x509 -sha256 -subj "/CN=$ORG.$ENV.search" \
    -key ./$ORG/$ENV/root-ca-key.pem -out ./$ORG/$ENV/root-ca.pem -days 3650