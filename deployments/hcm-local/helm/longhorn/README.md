```
helm install longhorn longhorn/longhorn -f values.yaml --namespace longhorn-system --create-namespace --version 1.4.2
```

kubectl patch ns/longhorn-system \
    --type json \
    --patch='[ { "op": "remove", "path": "/metadata/finalizers" } ]'