{
    RequestsPerSeconds: [
        {
            expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}",{{ GRPCServerMethod }} namespace="$namespace"}[$__rate_interval]))',
        },
        {
            expr: 'sum(rate(grpc_io_server_completed_rpcs{ {{ GRPCServerMethod }} app_kubernetes_io_name=~"{{ AppKubernetesIOName }}"}[$__rate_interval] offset $__range))',
        },
    ],
    SuccessfulRequestRateByMethods: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}",{{ GRPCServerMethod }} grpc_server_status="OK", namespace="$namespace"}[$__rate_interval])) by ( grpc_server_method) / sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}", namespace="$namespace"}[$__rate_interval])) by ( grpc_server_method)',
    },
    GRPCResponseStatus: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}",{{ GRPCServerMethod }}namespace="$namespace"}[$__rate_interval])) by (grpc_server_status) / ignoring(grpc_server_status) group_left sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}", {{ GRPCServerMethod }}namespace="$namespace"}[$__rate_interval]))',
    },
    LatencyByMethodP90: {
        expr: 'histogram_quantile(0.90, sum(rate(grpc_io_server_server_latency_bucket{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}",{{ GRPCServerMethod }}namespace="$namespace"}[$__rate_interval])) by (le, grpc_server_method, db))',
    },
    LatencyByMethodP95: {
        expr: 'histogram_quantile(0.95, sum(rate(grpc_io_server_server_latency_bucket{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}",{{ GRPCServerMethod }}namespace="$namespace"}[$__rate_interval])) by (le, grpc_server_method, db))',
    },
    PodMemoryUsage: [
        {
            expr: 'sum(container_memory_working_set_bytes{image!="",container!="POD",pod=~"{{ GoGoroutinesPod }}",namespace="$namespace"}) by (pod)',
        },
        {
            expr: 'sum(kube_pod_container_resource_requests{container!="POD",pod=~"{{ GoGoroutinesPod }}",namespace="$namespace",unit="byte"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"{{ GoGoroutinesPod }}"})',
        },
        {
            expr: 'sum(kube_pod_container_resource_limits{container!="POD",pod=~"{{ GoGoroutinesPod }}",namespace="$namespace",unit="core"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"{{ GoGoroutinesPod }}"})',
        },
    ],
    PodCpuUsage: [
        {
            expr: 'sum(rate(container_cpu_usage_seconds_total{image!="",container!="POD",pod=~"{{ GoGoroutinesPod }}",namespace="$namespace"}[$__rate_interval])) by (pod)',
        },
        {
            expr: 'sum(kube_pod_container_resource_requests{container!="POD",pod=~"{{ GoGoroutinesPod }}",namespace="$namespace",unit="core"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"{{ GoGoroutinesPod }}"})',
        },
        {
            expr: 'sum(kube_pod_container_resource_limits{container!="POD",pod=~"{{ GoGoroutinesPod }}",namespace="$namespace",unit="core"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"{{ GoGoroutinesPod }}"})',
        },
    ],
    Goroutines: {
        expr: 'go_goroutines{pod=~"{{ GoGoroutinesPod }}", namespace="$namespace"}',
    },
    RequestsPerSecondsByMethods: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ AppKubernetesIOName }}",{{ GRPCServerMethod }} namespace="$namespace"}[$__rate_interval])) by (grpc_server_method)',
    },
}