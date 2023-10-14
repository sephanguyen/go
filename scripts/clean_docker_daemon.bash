#!/bin/bash

# This scripts finds and removes all the Docker images that were built locally
# (as we continuously build new images, there would be a lot of obsolete ones).

# We use cache to speed up CI/CD on hcmo local. So we need check that the runner is running locally or on cloud.
use_hcmo_runner=false

# check using hcmo runners or not
if [[ $RUNNER_LABELS == *"arc-runner-hcm"* ]]; then
  echo "Using HCMO runner !!!"
  use_hcmo_runner=true
fi

# # echo "Stopping local registry..."
# if docker container inspect kind-registry >/dev/null 2>&1; then
#     echo "Stopping and removing kind-registry..."
#     docker container kill kind-registry >/dev/null
#     docker container rm kind-registry >/dev/null
# fi
kind delete cluster # delete cluster to run kind_with_registry.bash later

# Remove all built images on Cloud 
if [[ "$use_hcmo_runner" == "false" ]]; then
echo "Removing built Docker images..."
# docker image prune -f
read -a backend_image_list <<<"$(docker image ls --format "{{.ID}}:{{.Repository}}:{{.Tag}}" | grep asia.gcr.io/student-coach-e1e95/backend | xargs)"
read -a backend_sk_image_list <<<"$(docker image ls --format "{{.ID}}:{{.Repository}}:{{.Tag}}" | grep asia_gcr_io_student-coach-e1e95_backend | xargs)"
read -a aphelios_image_list <<<"$(docker image ls --format "{{.ID}}:{{.Repository}}:{{.Tag}}" | grep asia.gcr.io/student-coach-e1e95/aphelios | xargs)"
read -a aphelios_sk_image_list <<<"$(docker image ls --format "{{.ID}}:{{.Repository}}:{{.Tag}}" | grep asia_gcr_io_student-coach-e1e95_aphelios | xargs)"
image_list=("${backend_image_list[@]}" "${backend_sk_image_list[@]}" "${aphelios_image_list[@]}" "${aphelios_sk_image_list[@]}")
for image in "${image_list[@]}"; do
  parts=(${image//:/ })
  docker rmi -f ${parts[0]}
done
fi

# K8s HCMO runners use hostpath to cache the images such as kind, eibanam_cucumber, ... 
# to reduce internet bandwith and execution time. So that, to avoid full of disk on the nodes,
#  we need remove built images on K8s HCMO runners .
if [[ "$use_hcmo_runner" == "true" ]]; then
echo "Removing built Docker images on k8s hcmo runners"
# Remove built Docker images that older than 1 days
read -a backend_image_list <<< $(docker images --filter=reference='asia.gcr.io/student-coach-e1e95/backend:*' --format "{{.ID}}-{{.CreatedAt}}'" | cut -d " " -f 1 | sed 's/-/ /' | awk -v date="$(date --date='12 hours ago' +%Y-%m-%d)" '$NF < date' | cut -d " " -f 1 | xargs)
read -a learner_app_image_list <<< $(docker images --filter=reference='*-learner-app:*' --format "{{.ID}}-{{.CreatedAt}}'" | cut -d " " -f 1 | sed 's/-/ /' | awk -v date="$(date --date='12 hours ago' +%Y-%m-%d)" '$NF < date' | cut -d " " -f 1 | xargs)
read -a alive_image_list <<< $(docker images --filter=reference='*-alive:*' --format "{{.ID}}-{{.CreatedAt}}'" | cut -d " " -f 1 | sed 's/-/ /' | awk -v date="$(date --date='12 hours ago' +%Y-%m-%d)" '$NF < date' | cut -d " " -f 1 | xargs)
image_list=("${backend_image_list[@]}" "${learner_app_image_list[@]}" "${alive_image_list[@]}")
for image in "${image_list[@]}"; do
  parts=(${image//:/ })

  echo "Removing ${parts[0]}"
  docker rmi -f ${parts[0]}
done
fi

rm -rf ~/.skaffold/cache
echo "Done"
