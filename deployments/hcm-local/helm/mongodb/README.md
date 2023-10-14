
```
helm install my-release oci://registry-1.docker.io/bitnamicharts/mongodb
helm pull oci://registry-1.docker.io/bitnamicharts/mongodb

export MONGODB_ROOT_PASSWORD=
export MONGODB_USERNAME=
export MONGODB_PASSWORD=
export MONGODB_DATABASE=
export MONGODB_REPLICA_SET_KEY=w2bWVpuDgpE6edMk

helm install --namespace mongodb --create-namespace mongodb  -f values.yaml mongodb-13.15.3.tgz --dry-run > test.yaml
helm install --namespace mongodb --create-namespace mongodb  -f values.yaml mongodb-13.15.3.tgz
helm upgrade --namespace mongodb --create-namespace mongodb  -f values.yaml mongodb-13.15.3.tgz
```