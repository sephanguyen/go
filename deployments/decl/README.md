## Service definitions

Service definitions are located in:

- `stag-defs.yaml`
- `uat-defs.yaml`
- `prod-defs.yaml`

for `stag`, `uat`, and `prod` environment, respectively.

They are used to create and define properties of a service and define its properties.
Some regular use-cases:

- Create a new service with databases, hasura, kafka
- Update postgresql permissions of a service
- Update Google Cloud Project IAM permissions for a service

### Example usage

Define a service

```yaml
- name: bob
  postgresql:
    createdb: true
    grants:
      - dbname: eureka
      - dbname: bob
        grant_delete: true
  hasura:
    enabled: true
    v2_enabled: true
  kafka:
    enabled: true
  j4:
    allow_db_access: true
  identity_namespaces:
    - services
  iam_roles:
    - roles/firebase.admin
    - roles/iam.serviceAccountTokenCreator
    - roles/cloudsql.client
    - roles/cloudsql.instanceUser
  run_on_project_iam_roles:
    - roles/cloudsql.client
    - roles/cloudsql.instanceUser
    - roles/cloudprofiler.agent
  bucket_roles:
    manabie:
      stag-manabie-backend:
        - roles/storage.legacyBucketWriter
```

Separately, setting owner of a service is done in `owners.yaml`

```yaml
# owners.yaml
owners:
  bob: usermgmt
  zeus: platform
```

### Reference

The following attributes are supported for a service:

```yaml
  name: (string, required) name of the service
  postgresql: (map)
    createdb: |
      (boolean) whether to create database for this service.
      Defaults to false.
      If created, the database has the same name as the service, with a prefix
      specified by `db_prefix` variable in `config.hcl` of each project.
    grants: (array, optional) list of database permission grants for this service
      - dbname: (string, required) name of the database to grant access to
      - grant_delete: (boolean) whether to grant delete permission. Defaults to false.
    bypassrls: (boolean) whether to grant BYPASSRLS for this database user.
  hasura: (map)
    enabled: |
      (boolean) whether to enable hasura for this service. Defaults to false.
      When true, hasura will be granted permissions to this service's database
    v2_enabled: |
      (boolean) whether to enable hasura v2 for this service. Defaults to false.
      Requires `createdb` to be true to have any effects.
      When true:
        - a new hasura user will be created (name format: <env>-<org>-<service>-hasura@<project-id>.iam)
        - a new metadata database for hasura v2 will be created (name format: <dbprefix>_<service>_hasura_metadata)
        - permissions to the application database (read-only) + metadata database (read-write) will be granted to the new hasura user
  kafka: (map)
    enabled: (boolean) whether to enable kafka for this service. Defaults to false.
    grant_delete: |
      (boolean) whether to grant delete permission for kafka to this service's database.
      Defaults to false.
  j4: (map)
    allow_db_access: |
      (boolean) whether to grant read permission for j4 to this service's db 
      it does not apply to service that has db name different than svc name though, like yasuo having bob db
  identity_namespaces: (list) list of identity namespaces that the service belongs to. All options can be `serivces`, `machine-learning`, `nats-jetstream`, `elastic` and `kafka`
  iam_roles: (list) list of roles in `project_id`
  run_on_project_iam_roles: (list) list of roles in `runs_on_project_id`
  disable_iam: |
    (boolean) whether to disable IAM for this service. Defaults to false.
    When disabled, the service:
    - cannot to login to Cloud SQL using IAM
    - cannot encrypt/decrypt secrets using custom KMS key
    - cannot let the owner squad to encrypt/decrypt its secret
  bucket_roles: (map)
    <org>:
      <bucket_name>: (list) list of roles in <bucket_name>
```