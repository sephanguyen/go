## Configs & Secrets

### Documentation

- Config: [Confluence](https://manabie.atlassian.net/wiki/spaces/TECH/pages/481886611/Server+configuration)
- Secret: [Confluence](https://manabie.atlassian.net/wiki/spaces/TECH/pages/471728621/Secrets+management+sops)

### Quickstart

```sh
# Working with secrets in-place
cd deployments/helm/manabie-all-in-one/charts/bob/secrets/manabie/stag/
sops bob.secrets.encrypted.yaml
```
