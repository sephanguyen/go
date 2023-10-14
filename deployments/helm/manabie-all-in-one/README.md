# Service configuration

## Usage

### Generate a simple service

On the project root, run `make gen-helm-chart` and profile service name and port number.
To find the port number you can use, check [Service ports](#service-ports).

For example we're creating new 'awesomeservice' listening on port `6950`. You should find this output:

```
Your new service name: awesomeservice
Your new service port: 6950
<stdin>:243: trailing whitespace.

<stdin>:74: new blank line at EOF.
+
<stdin>:165: new blank line at EOF.
+
warning: 3 lines add whitespace errors.
```

_Please ignore the warning, I don't know why but it still work_
Check the git diff and you will see new helm chart for your `awesomeservice`. The structure look like this:

```
.
├── Chart.yaml
├── configs
│   ├── awesomeservice.common.config.yaml
│   ├── jprep
│   │   ├── prod
│   │   │   └── awesomeservice.config.yaml
│   │   ├── stag
│   │   │   └── awesomeservice.config.yaml
│   │   └── uat
│   │       └── awesomeservice.config.yaml
│   ├── manabie
│   │   ├── local
│   │   │   └── awesomeservice.config.yaml
│   │   ├── prod
│   │   │   └── awesomeservice.config.yaml
│   │   ├── stag
│   │   │   └── awesomeservice.config.yaml
│   │   └── uat
│   │       └── awesomeservice.config.yaml
│   └── tokyo
│       └── prod
│           └── awesomeservice.config.yaml
├── secrets
│   ├── jprep
│   │   ├── prod
│   │   ├── stag
│   │   └── uat
│   ├── manabie
│   │   ├── local
│   │   │   ├── awesomeservice_migrate.secrets.encrypted.yaml
│   │   │   ├── awesomeservice_migrate.secrets.example.yaml
│   │   │   ├── awesomeservice.secrets.encrypted.yaml
│   │   │   └── awesomeservice.secrets.example.yaml
│   │   ├── prod
│   │   ├── stag
│   │   └── uat
│   └── tokyo
│       └── prod
├── templates
│   └── app.yaml
└── values.yaml

49 directories, 22 files
```

After understanding this helm chart configuration, you can fill the common configuration and partner specific configuration, secret.  
Bellow are some info about the helm chart configuation

This one-line configuration is enough to config a simple GRPC service:

```yaml
grpcPort: 1234
```

By default it will create these k8s objects:

- ConfigMap
- Secret
- Pod Disruption Budget
- Service Account
- Deployment
- Service
- Virtual Service

### Enhanced service

Although above configuration is enough to config a simple GRPC service, normally a service has prerequisite things so it can run successfully, like:

- Need to run the database migration to create required tables.
- Expose metrics and need to tell Prometheus to collect its metrics.
- Enable Kubernetes readinessProbe.
- Config the Deployment replicas number.
- Config the container resource requests and limits.
- ...
- and many more.

This configuration may help do that:

```yaml
metrics:
  enabled: true # Enable the Kubernetes Pod Annotations. See the default annotations in Parameters section below.
migrationEnabled: true # Enable the database migration. An init-container will be created to perform the database migration.
waitForServices: # List of services that need to start first.
  - name: shamir
    port: 5680
grpcPort: 6950
readinessProbe: # Enable the Kubernetes ReadinessProbe. By default it uses the grpc_health_probe command, see Parameters section below.
  enabled: true
replicaCount: 2 # Config the Kubernetes Deployment replicas number. By default it is 1.
resources: # Config the Kubernetes Container resources requests and limits. By default it's not set.
  requests:
    memory: 64Mi
  limits:
    memory: 64Mi
```

### Full configuration (see [Parameters](#Parameters) below)

```yaml
metrics:
  enabled: true
  podAnnotations:
    prometheus.io/scheme: "http"
    prometheus.io/port: "8888"
    prometheus.io/scrape: "true"

migrationEnabled: true

grpcPort: 6950

hasuraEnabled: false

httpPort: 8080

pdbEnabled: true
pdbMaxUnavailable: 1

podAnnotations:
  sidecar.istio.io/proxyCPU: "10m"
  sidecar.istio.io/proxyMemory: "50Mi"

readinessProbe:
  enabled: true
  periodSeconds: 5
  initialDelaySeconds: 5
  timeoutSeconds: 5
  successThreshold: 1
  failureThreshold: 5
  command:
    exec:
      command:
        - sh
        - -c
        - /bin/grpc_health_probe -addr=localhost:5950 -connect-timeout 250ms -rpc-timeout 250ms
  # or you can override it like this:
  # command:
  #   httpGet:
  #     path: /
  #     port: 8080

replicaCount: 1

resources:
  requests:
    cpu: 10m
    memory: 128Mi
  limits:
    cpu: 20m
    memory: 256Mi

apiHttp:
  - match:
      - uri:
          prefix: /zeus.v1
        route:
          - destination:
              host: zeus
              port:
                number: 5950

webHttp:
  - match:
      - uri:
          prefix: /zeus.v1
    route:
      - destination:
          host: zeus
          port:
            number: 5950
    # corsPolicy is optional, if you don't config it, this default will be used:
    corsPolicy:
      allowOrigins:
        - regex: ".*"
      allowMethods:
        - POST
        - GET
        - OPTIONS
        - PUT
        - DELETE
      allowHeaders:
        - authorization
        - grpc-timeout
        - content-type
        - keep-alive
        - user-agent
        - cache-control
        - content-transfer-encoding
        - token
        - x-accept-content-transfer-encoding
        - x-accept-response-streaming
        - x-request-id
        - x-user-agent
        - x-graphql-mesh-authorization
        - x-grpc-web
        - if-none-match
        - pkg
        - version
      maxAge: 100s
      exposeHeaders:
        - grpc-status
        - grpc-message
        - etag
    # or you can override it like this:
    # corsPolicy:
    #   allowOrigins:
    #     - exact: 'example.com'
    #   allowMethods:
    #     - OPTIONS

adminHttp:
  - match:
      - uri:
          exact: /healthz
      - uri:
          prefix: /console
      - uri:
          prefix: /v1
      - uri:
          prefix: //v1 # bug: https://github.com/hasura/graphql-engine/issues/7196
      - uri:
          prefix: /v2
    route:
      - destination:
          host: zeus-hasura
          port:
            number: 8080
```

## Parameters

| Name                                 | Description                                                                                                                | Default                                                                                                                                                     | Required |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| `metrics.enabled`                    | Enable metrics annotations                                                                                                 | false                                                                                                                                                       | no       |
| `metrics.podAnnotations`             | Annotations to tell Prometheus to collect metrics                                                                          | <pre>prometheus.io/scheme: "http"<br>prometheus.io/port: "8888"<br>prometheus.io/scrape: "true"</pre>                                                       | no       |
| `migrationEnabled`                   | Enable database migration for service                                                                                      | `false`                                                                                                                                                     | no       |
| `waitForServices`                    | List of services need to start first. Notes: If we want to set no services starting first you should assign empty array [] | <pre>- name: shamir<br> port: 5680</pre>                                                                                                                    | no       |
| `grpcPort`                           | GRPC port                                                                                                                  | `""`                                                                                                                                                        | yes      |
| `hasuraEnabled`                      | Enable Hasura                                                                                                              | `false`                                                                                                                                                     | no       |
| `httpPort`                           | HTTP port                                                                                                                  | `""`                                                                                                                                                        | no       |
| `pdbEnabled`                         | Enable Pod Disruption Budget (PDB) for service                                                                             | `true`                                                                                                                                                      | no       |
| `pdbmaxUnavailable`                  | Set PDB max unavailable pods                                                                                               | `1`                                                                                                                                                         | no       |
| `podAnnotations`                     | Set annotations for pods                                                                                                   | null                                                                                                                                                        | no       |
| `readinessProbe.enabled`             | Enable service readiness probe                                                                                             | `false`                                                                                                                                                     | no       |
| `readinessProbe.periodSeconds`       | Set readiness probe periodSeconds                                                                                          | `5`                                                                                                                                                         | no       |
| `readinessProbe.initialDelaySeconds` | Set readiness probe initialDelaySeconds                                                                                    | `5`                                                                                                                                                         | no       |
| `readinessProbe.timeoutSeconds`      | Set readiness probe timeoutSeconds                                                                                         | `5`                                                                                                                                                         | no       |
| `readinessProbe.successThreshold`    | Set readiness probe successThreshold                                                                                       | `1`                                                                                                                                                         | no       |
| `readinessProbe.failureThreshold`    | Set readiness probe failureThreshold                                                                                       | `5`                                                                                                                                                         | no       |
| `readinessProbe.command`             | Set readiness probe command                                                                                                | <pre>exec: <br> command: <br> - sh<br> - -c<br> - /bin/grpc_health_probe -addr=localhost:{{ service port }} -connect-timeout 250ms -rpc-timeout 250ms</pre> | no       |
| `replicaCount`                       | Set deployment replicas                                                                                                    | `1`                                                                                                                                                         | no       |
| `resources.requests`                 | Set container resource requests                                                                                            | null                                                                                                                                                        | no       |
| `resources.limits`                   | Set container resource limits                                                                                              | null                                                                                                                                                        | no       |
| `apiHttp`                            | Set Istio virtual service for GRPC service                                                                                 | null                                                                                                                                                        | no       |
| `webHttp`                            | Set Istio virtual service for GRPC web service                                                                             | null                                                                                                                                                        | no       |
| `adminHttp`                          | Set Istio virtual service for Hasura service                                                                               | null                                                                                                                                                        | no       |
| `preHookUpsertStream`                | Enabled pre-hook run a job to upsert streams of nats-jetstream                                                             | `false`                                                                                                                                                     | no       |

### Service ports

This section lists the ports used by the services. Please update this list whenever
you are adding a new service.

Service ports:

- 50xx: Bob
- 51xx: Tom
- 52xx: Yasuo
- 53xx: Enigma
- 54xx: Fatima
- 55xx: Eureka
- 56xx: Shamir
- 57xx: Teacher Web
- 58xx: Gandalf mock server
- 59xx: Zeus
- 60xx: Draft
- 61xx: User Management (`usermgmt`)
- 62xx: Payment
- 63xx: Entry/Exit Management (`entryexitmgmt`)
- 64xx: Master Management (`mastermgmt`)
- 65xx: Lesson Management (`lessonmgmt`)
- 66xx: Invoice Management (`invoicemgmt`)
- 67xx: Virtual Classroom (`virtualclassroom`)
- 68xx: Timesheet (`timesheet`)
- 69xx: Notification Management (`notificationmgmt`)
- 70xx: Calendar (`calendar`)
- 71xx: Hephaestus (`hephaestus`)
- 72xx: Scheduling (`scheduling`)
- 74xx: Discount (`discount`)

Protocol ports:

- xx50: gRPC
- xx80: HTTP

For example, Shamir gRPC port is: 5650.

When setting up a new service, use the next number.
For example, when the highest port number is `5650` for shamir, use `5750` and `5780` for your service.
