#!/bin/bash

# Convenient script to trigger preproduction deployment

cat <<EOF | gh workflow run deployment.preproduction --json
{
  "be-tag": "20230705000000.bb1dc2b5d3",
  "fe-tag": "20230704044525.58b505ec",
  "me-tag": "20230703021723.07d8317d",
  "organization": "tokyo",
  "sync-database": "false",
  "install-gateway": "false",
  "install-backbone": "true",
  "install-services": "true"
}
EOF
