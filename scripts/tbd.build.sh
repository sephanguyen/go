#!/bin/bash

# Convenient script to trigger build

cat <<EOF | gh workflow run tbd.build --json
{
  "auto_deploy": "false",
  "env": "preproduction",
  "orgs": "tokyo",
  "be_release_tag": "20230705000000.bb1dc2b5d3",
  "fe_release_tag": "20230704044525.58b505ec",
  "me_release_tag": "20230703021723.07d8317d",
  "me_apps": "learner, teacher",
  "me_platforms": "web"
}
EOF
