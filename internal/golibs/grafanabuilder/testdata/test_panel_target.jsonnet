//
{
    NumberOfRequests: {
        expr: 'sum(increase(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}", namespace="$namespace"}[5m])) by (grpc_server_method)',
        legendFormat: '{{ grpc_server_method }}',
    },
    AverageRequestsPerSeconds: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__range]))',
        legendFormat: '',
    },
    TotalNumberOfRequests: {
        expr: 'sum(increase(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__range])) by(grpc_server_method)',
        legendFormat: '{{grpc_server_method}}',
    },
    RateOfOKStatus: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}", grpc_server_status="OK", namespace="$namespace"}[$__rate_interval])) by (grpc_server_method) / sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__rate_interval])) by (grpc_server_method)',
        legendFormat: '{{grpc_server_method}}',
    },
    ErrorRateByMethod: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}", grpc_server_status!~"OK", namespace="$namespace"}[5m])) by (grpc_server_status, grpc_server_method) / sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}", namespace="$namespace"}[5m])) by (grpc_server_status, grpc_server_method)',
        legendFormat: '{{ grpc_server_method }}: {{ grpc_server_status }}',
    },
    GRPCResponseStatus: {
        expr: 'sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__rate_interval])) by (grpc_server_status) / ignoring(grpc_server_status) group_left sum(rate(grpc_io_server_completed_rpcs{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}", grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__rate_interval]))',
        legendFormat: '{{grpc_server_status}}',
    },
    LatencyByMethodP95: {
        expr: 'histogram_quantile(0.95, sum(rate(grpc_io_server_server_latency_bucket{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__rate_interval])) by (le, grpc_server_method, db))',
        legendFormat: '{{grpc_server_method}}',
    },
    LatencyByMethodP99: {
        expr: 'histogram_quantile(0.99, sum(rate(grpc_io_server_server_latency_bucket{app_kubernetes_io_name=~"{{ .AppKubernetesIOName }}",grpc_server_method=~"{{ .grpcServerMethod }}",namespace="$namespace"}[$__rate_interval])) by (le, grpc_server_method, db))',
        legendFormat: '{{grpc_server_method}}',
    },
    PodMemoryUsage: [
        {
            expr: 'sum(container_memory_working_set_bytes{image!="",container!="POD",pod=~"^({{ .service }}).+",namespace="$namespace"}) by (pod)',
            legendFormat: '{{ pod }}',
            format: 'time_series',
            intervalFactor:1,
        },
        {
            expr: 'sum(kube_pod_container_resource_requests{container!="POD",pod=~"^({{ .service }}).+",namespace="$namespace",unit="byte"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"^({{ .service }}).+"})',
            legendFormat: 'Requests',
        },
        {
            expr: 'sum(kube_pod_container_resource_limits{container!="POD",pod=~"^({{ .service }}).+",namespace="$namespace",unit="core"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"^({{ .service }}).+"})',
            legendFormat: 'Limits',
        },
    ],
    PodCpuUsage: [
        {
            expr: 'sum(rate(container_cpu_usage_seconds_total{image!="",container!="POD",pod=~"^({{ .service }}).+",namespace="$namespace"}[$__rate_interval])) by (pod)',
            legendFormat: '{{ pod }}',
        },
        {
            expr: 'sum(kube_pod_container_resource_requests{container!="POD",pod=~"^({{ .service }}).+",namespace="$namespace",unit="core"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"^({{ .service }}).+"})',
            legendFormat: 'Requests',
        },
        {
            expr: 'sum(kube_pod_container_resource_limits{container!="POD",pod=~"^({{ .service }}).+",namespace="$namespace",unit="core"}) /
                   count(kube_pod_info{namespace="$namespace",pod=~"^({{ .service }}).+"})',
            legendFormat: 'Limits',
        },
    ],
    Goroutines: {
        expr: 'go_goroutines{pod=~"{{ .GoGoroutinesPod }}", namespace="$namespace"}',
        legendFormat: '{{ pod }}',
    },
}