#!/bin/bash
org=${ORG:-}
env=${ENV:-}
ci=${CI:-false} # we don't need prompt. If we use this bash on workflow
namespace=${NAMESPACE:-backend}
COMMAND=${COMMAND:-}

if [[ -z "$org" ]] || [[ -z "$env" ]] ; then 
    echo "ERROR: require 'org, env, argument' is not empty, for example:"
    echo "org: manabie, jprep, aic,..."
    echo "env: local, stag,..."
    exit 1
fi

echo "=========== Current context: "
kubectl config current-context

# check prompt
if [[ $ci != true ]]; then
    read -p "Are you sure with your current context [Y/N]?  " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Restart service stopped"
        exit 0
    fi
fi

IFS=' ' read -r -a FILE <<< "$COMMAND"

# # Grant permission for file
chmod 700 $FILE
# Run script file
echo "===================================================RESULT=================================================="
eval $COMMAND
exitcode=$?
echo "===================================================RESULT=================================================="
exit $exitcode