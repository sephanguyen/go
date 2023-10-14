{
    Rpcs: [
        {
            expr: 'sum(rate(istio_requests_total{destination_service_name=~"{{ Service }}", destination_service_namespace="$namespace"}[$__rate_interval]))',
        },
        {
            expr: 'sum(rate(istio_requests_total{destination_service_name=~"{{ Service }}", destination_service_namespace="$namespace"}[$__rate_interval] offset $__range))',
        },
    ],
    StatusCodeRate: {
        expr: 'sum(rate(istio_requests_total{destination_service_name=~"{{ Service }}", destination_service_namespace="$namespace"}[$__rate_interval])) by (response_code) / ignoring(response_code) group_left sum(rate(istio_requests_total{destination_service_name=~"{{ Service }}", destination_service_namespace="$namespace"}[$__rate_interval])) * 100 > 0',
    },
    LatencyByMethodP90: [
        {
            expr: 'histogram_quantile(0.90, sum(rate(istio_request_duration_milliseconds_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval])) by (le))',
        },
        {
            expr: 'histogram_quantile(0.90, sum(rate(istio_request_duration_milliseconds_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval] offset $__range)) by (le))',
        },
    ],
    LatencyByMethodP99: [
        {
            expr: 'histogram_quantile(0.99, sum(rate(istio_request_duration_milliseconds_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval])) by (le))',
        },
        {
            expr: 'histogram_quantile(0.99, sum(rate(istio_request_duration_milliseconds_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval] offset $__range)) by (le))',
        },
    ],
    PodCpuUsage: [
        {
            expr: 'sum(rate(container_cpu_usage_seconds_total{image!="",container!="POD",pod=~"^({{ Service }}-[^hasuramigrateproxyjprep]).*",namespace="$namespace"}[5m])) by (pod)',
        },
        {
            expr: 'sum(kube_pod_container_resource_requests{container!="POD",pod=~"^({{ Service }}-[^hasuramigrateproxyjprep]).*",namespace="$namespace",unit="core"}) / count(kube_pod_info{namespace="$namespace",pod=~"^(bob-[^hasuramigrateproxyjprep]).*"})',
        },
        {
            expr: 'sum(kube_pod_container_resource_limits{container!="POD",pod=~"^({{ Service }}-[^hasuramigrateproxyjprep]).*",namespace="$namespace",unit="core"}) / count(kube_pod_info{namespace="$namespace",pod=~"^(bob-[^hasuramigrateproxyjprep]).*"})',
        },
    ],
    PodMemoryUsage: [
        {
            expr: 'sum (container_memory_working_set_bytes{image!="",container!="POD",pod=~"^({{ Service }}-[^hasuramigrateproxyjprep]).*",namespace="$namespace"}) by (pod)',
        },
        {
            expr: 'sum(kube_pod_container_resource_requests{container!="POD",pod=~"^({{ Service }}-[^hasuramigrateproxyjprep]).*",namespace="$namespace",unit="byte"}) / count(kube_pod_info{namespace="$namespace",pod=~"^(bob-[^hasuramigrateproxyjprep]).*"})',
        },
        {
            expr: 'sum(kube_pod_container_resource_limits{container!="POD",pod=~"^({{ Service }}-[^hasuramigrateproxyjprep]).*",namespace="$namespace",unit="core"}) / count(kube_pod_info{namespace="$namespace",pod=~"^(bob-[^hasuramigrateproxyjprep]).*"})',
        },
    ],
    ReceiveBytes: [
        {
            expr: 'histogram_quantile(0.99, sum(rate(istio_request_bytes_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval])) by (le))',
        },
        {
            expr: 'histogram_quantile(0.99, sum(rate(istio_request_bytes_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval] offset $__range)) by (le))',
        },
    ],
    SentBytes: [
        {
            expr: 'histogram_quantile(0.99, sum(rate(istio_response_bytes_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval])) by (le))',
        },
        {
            expr: 'histogram_quantile(0.99, sum(rate(istio_response_bytes_bucket{destination_service_name=~"{{ Service }}",destination_service_namespace="$namespace"}[$__rate_interval] offset $__range)) by (le))',
        },
    ],
}