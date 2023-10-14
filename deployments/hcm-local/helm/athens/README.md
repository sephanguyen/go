```
helm repo add gomods https://gomods.github.io/athens-charts


helm install --namespace athens --create-namespace mongodb  -f values.yaml athens-proxy-0.7.0.tgz --dry-run > test.yaml
helm upgrade --namespace athens --create-namespace mongodb  -f values.yaml athens-proxy-0.7.0.tgz
```