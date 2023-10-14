#!/bin/bash

set -eux


prefix_endpoint="https:\/\/admin.local-green.manabie.io:31600\/"
local_endpoint="http:\/\/localhost:12345\/"

function apply_hasura_metadata {
   if kubectl get deployments/$service -n backend >/dev/null 2>&1; then
      echo "==== Deployment $service found. ===="
     
      PORT_FORWARD=$(kubectl -n backend port-forward service/$service --pod-running-timeout=1m0s 12345:8080 >/dev/null &)
      ./scripts/wait-for.sh localhost:12345 -t 60

      if $PORT_FORWARD ; then
         
         # replace https://admin.local-green.manabie.io:31600 by localhost    
         sed_arg="s/$endpoint/$local_endpoint/g"     
         sed -i $sed_arg $path/config.yaml

         echo "Applying metadata for $path"

         hasura metadata apply --project $path --skip-update-check
         if [[ $? != 0 ]]; then
            echo "::error title=hasura::Hasura metadata apply failed"
         fi

         # kill process ID of port 12345
         kill -9 $(lsof -t -i:12345) 
      fi

   fi
}


hasura version --skip-update-check || curl -L https://github.com/hasura/graphql-engine/raw/stable/cli/get.sh | VERSION=v2.8.1 bash

v2=("draft")


#install python lib for parsing db name from hcl file
pip3 install python-hcl2 typing_extensions pyyaml
v1=($(python3 deployments/services-directory/hcl2hasura.py))

echo "Checking metadata for v1"
for svc in "${v1[@]}"
do
   path="deployments/helm/manabie-all-in-one/charts/$svc/files/hasura"
   endpoint="$prefix_endpoint$svc"
   if [[ $svc == "bob" ]]; then 
      endpoint="$prefix_endpoint"
   fi

   endpoint=$endpoint service=$svc-hasura path=$path apply_hasura_metadata 
done


echo "Checking metadata for v2"
for svc in "${v2[@]}" 
do
   path="deployments/helm/manabie-all-in-one/charts/$svc/files/hasurav2"

   endpoint="$prefix_endpoint$svc""v2\/"
   endpoint=$endpoint service=$svc-hasurav2 path=$path apply_hasura_metadata 
done

echo "exit code $?"
