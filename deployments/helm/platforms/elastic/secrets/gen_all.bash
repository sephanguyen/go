#!/bin/bash
# ORG=manabie ENV=staging ./gen_all.bash
ORG=$ORG ENV=$ENV ./gen_ca.bash
ORG=$ORG ENV=$ENV NAME=admin ./gen_cert.bash
ORG=$ORG ENV=$ENV NAME=node ./gen_cert.bash