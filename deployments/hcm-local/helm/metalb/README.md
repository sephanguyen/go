```
helm install metallb metallb/metallb -f values.yaml -n metallb-system --create-namespace
kubectl -n metallb-system -create -f l2-address-pool.yaml
```