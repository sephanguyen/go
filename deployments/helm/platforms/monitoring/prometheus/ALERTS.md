# Prometheus Alerts

See Also [Terraform Alerts](/deployments/terraform/modules/alert-policies/ALERTS.md)

## Kubernetes Apps

**Kubernetes Pod Crash Looping**

`"{{ $labels.pod }} Pod is CrashLooping"`

`"Pod {{ $labels.namespace }}/{{ $labels.pod }} ({{ $labels.container }}) is restarting {{ $value }} times / 15 minutes."`

**Kubernetes Pod Not Ready**

`"{{ $labels.pod }} Pod in a non-ready state"`

`"Pod {{ $labels.namespace }}/{{ $labels.pod }} has been in a non-ready state for longer than 15 minutes."`

**Kubernetes Deployment Replicas Mismatch**

`"{{ $labels.deployment }} Replica count mismatch"`

`"Deployment {{ $labels.namespace }}/{{ $labels.deployment }} has not matched the expected number of replicas for longer than 15 minutes."`

**Kubernetes StatefulSet Replicas Mismatch**

`"{{ $labels.statefulset }} StatefulSet replica mismatch"`

`"StatefulSet {{ $labels.namespace }}/{{ $labels.statefulset }} has not matched the expected number of replicas for longer than 15 minutes."`

## Kubernetes Storage

**Kubernetes Volume Out Of Disk Space**

`"{{ $labels.persistentvolumeclaim }} Is running out of storage capacity"`

`"PersistentVolume claimed by {{ $labels.persistentvolumeclaim }} in Namespace {{ $labels.namespace }} is only {{ $value | humanizePercentage }} free."`

## Kubernetes Resources

**Container Memory Usage**

`A container in the {{ $labels.namespace }} namespace uses more than 95% of the memory limit`

`"{{ $labels.container }} in {{ $labels.pod }} is using {{ $value }}% of available memory."`

## NATS

**NATS Cluster Down**

`NATS Cluster down`

`"Less than 3 nodes running in NATS cluster\n VALUE = {{ $value }}"`

**NATS Active Server Down**

`expr: avg_over_time(nss_server_info{state="FT_ACTIVE"}[5m]) < 0.9`

`for: 5m`

## Backend

**High Number Of Failed GRPC Requests**

`expr: sum by(grpc_server_method, kubernetes_namespace) (rate(grpc_io_server_completed_rpcs{grpc_server_method!~".+Health.+|.+Subscribe|.+SubscribeV2|.+StreamingEvent|.+RetrieveTopicIcon",grpc_server_status=~"UNKNOWN|INTERNAL"}[5m]) > 0) / sum by(grpc_server_method, kubernetes_namespace) (rate(grpc_io_server_completed_rpcs{grpc_server_method!~".+Health.+|.+Subscribe|.+StreamingEvent"}[5m]) > 0) > 0.1`

`for: 15m`

**High Number Of Slow GRPC Requests**

`expr: histogram_quantile(0.95, sum by(le, grpc_server_method, kubernetes_namespace) (rate(grpc_io_server_server_latency_bucket{grpc_server_method!~".+Health.+|.+Upload|.+Subscribe|.+SubscribeV2|.+StreamingEvent|.+RetrieveTopicIcon"}[5m]) > 0)) / 1000 > 2.5`

`for: 15m`
