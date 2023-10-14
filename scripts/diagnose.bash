#!/bin/bash

set -u

YEL='\033[1;33m'
NC='\033[0m'
WARN="${YEL}[WARNING]${NC}"

function exit_func() {
  echo "============================================================"
  kubectl get pods -A
  echo "============================================================"
  kubectl get events -A --sort-by='.lastTimestamp'
  echo "============================================================"
}
if [[ "${CI:-false}" == "true" ]]; then
  trap exit_func EXIT
fi

# Get all namespaces
nsList=$(kubectl get namespaces --no-headers -o custom-columns=":metadata.name")
# Iterate each namespace and check status
for ns in ${nsList}; do
  deploymentList=$(kubectl get deployments.app -n "${ns}" --no-headers -o custom-columns=":metadata.name")
  for deployment in ${deploymentList}; do
    data=$(kubectl -n "$ns" get deployments.app "$deployment" -o json)
    readyReplicas=$(echo "$data" | jq -r '.status.readyReplicas // 0')
    replicas=$(echo "$data" | jq -r '.spec.replicas')
    if [[ "$readyReplicas" != "$replicas" ]]; then
      echo -e "$WARN Deployment ${ns}.${deployment} is not ready ($readyReplicas/$replicas)"
    fi
  done

  statefulsetList=$(kubectl get statefulset.app -n "${ns}" --no-headers -o custom-columns=":metadata.name")
  for sts in ${statefulsetList}; do
    data=$(kubectl -n "$ns" get statefulset.app "$sts" -o json)
    readyReplicas=$(echo "$data" | jq -r '.status.readyReplicas // 0')
    replicas=$(echo "$data" | jq -r '.spec.replicas')
    if [[ "$readyReplicas" != "$replicas" ]]; then
      echo -e "$WARN Statefulset ${ns}.${sts} is not ready ($readyReplicas/$replicas)"
    fi
  done

  jobList=$(kubectl get jobs.batch -n "${ns}" --no-headers -o custom-columns=":metadata.name")
  for job in ${jobList}; do
    data=$(kubectl -n "$ns" get jobs.batch "$job" -o json)
    successCount=$(echo "$data" | jq -r '.status.succeeded // 0')
    if [[ "$successCount" == "0" ]]; then
      finalStatus=$(echo "$data" | jq -r '.status')
      echo -e "$WARN Job ${ns}.${job} did not succeed; final status: $finalStatus"
    fi
  done

  podList=$(kubectl get pods -n "${ns}" --no-headers -o custom-columns=":metadata.name")
  for pod in ${podList}; do
    # Skip pods that belongs to cronjob, since they may disappear when we get to them
    if [[ "$pod" == "yasuo-send-scheduled-notification"* ]]; then
      continue
    fi

    # Check for pod status
    podStatus=$(kubectl -n "$ns" get pod "$pod" -o jsonpath='{.status.phase}')
    if [[ "$podStatus" != "Running" && "$podStatus" != "Succeeded" ]]; then
      echo -e "$WARN Pod ${ns}.${pod} is not running ($podStatus)"
    fi

    if [[ "$podStatus" == "Pending" ]]; then
      echo "Printing out events related to this \"Pending\" pod (kubectl -n $ns get events | grep $pod)"
      echo "============================================================"
      kubectl -n "$ns" get events | grep "$pod"
      echo "============================================================"
      echo "End of events"
      echo
    fi

    # Check for each container inside pod. Note that there can be no container when pod is pending.
    kubectl -n "$ns" get pod "$pod" -o jsonpath='{.status}' | jq -c '.initContainerStatuses + .containerStatuses // []' | jq -c '.[]' | while read -r data; do
      containerName=$(echo "$data" | jq -r '.name')
      currentState=$(echo "$data" | jq '.state' | jq -r 'keys[]')
      stateData=$(echo "$data" | jq '.state')
      if [[ "$currentState" == "running" ]]; then
        continue
      fi

      if [[ "$currentState" == "terminated" ]]; then
        reason=$(echo "$stateData" | jq -r '.terminated.reason')
        exitCode=$(echo "$stateData" | jq -r '.terminated.exitCode')
        # if [[ "$reason" == "Completed" && "$exitCode" == "0" && "$helmStatus" == "deployed" ]]; then
        if [[ "$reason" == "Completed" && "$exitCode" == "0" ]]; then
          continue
        fi
      fi

      echo -e "$WARN Container ${ns}.${pod}.${containerName} is in \"${currentState}\" state: "
      echo "$stateData" | jq

      # If .status.waiting.reason is "PodInitializing", there is no logs to get
      if [[ "$(echo "$stateData" | jq -r '.waiting.reason')" == "PodInitializing" ]]; then
        continue
      fi

      echo "Printing out container ${ns}.${pod}.${containerName} logs (kubectl -n $ns logs $pod -c $containerName):"
      echo "============================================================"
      kubectl -n "$ns" logs "$pod" -c "$containerName"
      echo "============================================================"
      echo "End of logs"
      echo ""

      echo "Printing out events related to this pod (kubectl -n $ns get events | grep $pod)"
      echo "============================================================"
      kubectl -n "$ns" get events | grep "$pod"
      echo "============================================================"
      echo "End of events"
      echo
    done
  done
done

echo "Diagnosis completed."
