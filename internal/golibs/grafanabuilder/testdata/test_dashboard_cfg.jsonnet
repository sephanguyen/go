local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local annotation = grafana.annotation;
local template = grafana.template;
local row = grafana.row;
local graphPanel = grafana.graphPanel;
local statPanel = grafana.statPanel;
local pieChartPanel = grafana.pieChartPanel;
local barGaugePanel = grafana.barGaugePanel;
local prometheus = grafana.prometheus;
local pannelTarget = import 'test_panel_target.jsonnet';

dashboard.new(
  'Virtual Classroom',
  editable=true,
  refresh='5s',
  time_from='now-6h',
  time_to='now',
  timepicker={},
  schemaVersion=27,
  uid="UID",
)
.addAnnotation(annotation.default)
.addTemplate(
  template.datasource(
    name='cluster',
    query='prometheus',
    current='Thanos',
    hide='',
  )
)
.addTemplate(
  template.new(
    name='namespace',
    datasource='${cluster}',
    query={
        query: 'label_values(grpc_io_server_completed_rpcs, namespace)',
        refId: 'StandardVariableQuery'
    },
    label='Namespace',
    hide='',
    refresh='load',
    definition='label_values(grpc_io_server_completed_rpcs, namespace)'
  )
)
.addPanel(
  row.new(
    title='Basic statistics',
    titleSize='h6',
  ), gridPos={
    h: 1,
    w: 24,
    x: 0,
    y: 0
  }
)
.addPanel(
  graphPanel.new(
    title='Number of requests',
    datasource='${cluster}',
    fill=1,
    legend_show=true,
    lines=true,
    linewidth=1,
    pointradius=2,
    stack=true,
    shared_tooltip=true,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='none',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.NumberOfRequests.expr,
        legendFormat=pannelTarget.NumberOfRequests.legendFormat,
    )
  ), gridPos={
    x: 0,
    y: 1,
    w: 18,
    h: 10,
  }
)
.addPanel(
  statPanel.new(
    title='Average requests per seconds',
    datasource='${cluster}',
    reducerFunction='lastNotNull',
    graphMode='none',
    unit=null,
    pluginVersion='7.5.12',
  ).addThreshold(
  [
    {
     color: "blue",
     value: null
    },
    {
      color: "orange",
      value: 50
    },
    {
      color: "red",
      value: 100
    }
  ]
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.AverageRequestsPerSeconds.expr,
        legendFormat=pannelTarget.AverageRequestsPerSeconds.legendFormat,
    )
  ), gridPos={
    x: 18,
    y: 1,
    w: 6,
    h: 10,
  }
)
.addPanel(
  pieChartPanel.new(
    title='Total number of requests',
    datasource='${cluster}',
    valueName=null,
    pluginVersion='7.5.12',
  ).addThreshold(
     [
       {
        color: "green",
        value: null
       },
       {
         color: "red",
         value: 80
       }
     ]
  ).addTarget(
       prometheus.custom_target(
         expr=pannelTarget.TotalNumberOfRequests.expr,
         legendFormat=pannelTarget.TotalNumberOfRequests.legendFormat,
       )
  ), gridPos={
   x: 0,
   y: 11,
   w: 12,
   h: 14,
  }
)
.addPanel(
 barGaugePanel.new(
  title='Rate of "OK" status',
  datasource='${cluster}',
  unit='percentunit',
  max=1,
  thresholdsMode='percentage',
  showUnfilled=true,
 ).addThreshold(
   [
     {
      color: "orange",
      value: null
     },
     {
       color: "red",
       value: 20
     }
   ]
 ).addTarget(
     prometheus.custom_target(
        expr=pannelTarget.RateOfOKStatus.expr,
        legendFormat=pannelTarget.RateOfOKStatus.legendFormat,
     )
 ), gridPos={
    x: 12,
    y: 11,
    w: 12,
    h: 14,
 }
)
.addPanel(
  graphPanel.new(
    title='Error rate by method',
    datasource='${cluster}',
    fill=1,
    unit='percentunit',
    legend_hideEmpty=true,
    legend_hideZero=true,
    legend_show=true,
    pointradius=2,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='percentunit',
    max='1',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.ErrorRateByMethod.expr,
        legendFormat=pannelTarget.ErrorRateByMethod.legendFormat,
    )
  ), gridPos={
    x: 0,
    y: 25,
    w: 24,
    h: 8,
  }
)
.addPanel(
  graphPanel.new(
    title='gRPC Response status',
    datasource='${cluster}',
    aliasColors={
     FailedPrecondition: 'semi-dark-red',
     NotFound: 'dark-orange',
     OK: 'dark-green',
    },
    bars=true,
    fill=1,
    legend_hideEmpty=true,
    legend_hideZero=true,
    legend_show=true,
    legend_values=true,
    legend_min=true,
    legend_max=true,
    lines=false,
    percentage=true,
    pointradius=2,
    stack=true,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='percentunit',
    max='100',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.GRPCResponseStatus.expr,
        legendFormat=pannelTarget.GRPCResponseStatus.legendFormat,
    )
  ), gridPos={
    x: 0,
    y: 33,
    w: 24,
    h: 8,
  }
)
.addPanel(
  graphPanel.new(
    title='Latency by Method (P95)',
    datasource='${cluster}',
    fill=1,
    unit='s',
    pointradius=2,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='s',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.LatencyByMethodP95.expr,
        legendFormat=pannelTarget.LatencyByMethodP95.legendFormat,
    )
  ), gridPos={
    x: 0,
    y: 41,
    w: 24,
    h: 8,
  }
)
.addPanel(
  graphPanel.new(
    title='Latency by Method (P99)',
    datasource='${cluster}',
    fill=1,
    unit='s',
    pointradius=2,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='s',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.LatencyByMethodP99.expr,
        legendFormat=pannelTarget.LatencyByMethodP99.legendFormat,
    )
  ), gridPos={
    x: 0,
    y: 49,
    w: 24,
    h: 8,
  }
)
.addPanel(
  graphPanel.new(
    title='Pod Memory usage',
    datasource='${cluster}',
    fill=1,
    decimals=1,
    legend_avg=true,
    legend_current=true,
    legend_max=true,
    legend_values=true,
    legend_sideWidth=350,
    pointradius=2,
    value_type='individual',
  ).addSeriesOverride(
    {
      alias: 'Requests',
      color: '#5794F2',
      legend: false,
      nullPointMode: 'connected'
    }
  ).addSeriesOverride(
    {
     alias: 'Limits',
     color: '#F2495C',
     legend: false,
     nullPointMode: 'connected'
    }
  ).resetYaxes().
  addYaxis(
    format='bytes',
  ).addYaxis(
    format='short',
    show=false,
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.PodMemoryUsage[0].expr,
        format=pannelTarget.PodMemoryUsage[0].format,
        legendFormat=pannelTarget.PodMemoryUsage[0].legendFormat,
        intervalFactor=pannelTarget.PodMemoryUsage[0].intervalFactor,
    )
  ).addTarget(
       prometheus.custom_target(
           expr=pannelTarget.PodMemoryUsage[1].expr,
           legendFormat=pannelTarget.PodMemoryUsage[1].legendFormat,
       )
  ).addTarget(
      prometheus.custom_target(
           expr=pannelTarget.PodMemoryUsage[2].expr,
           legendFormat=pannelTarget.PodMemoryUsage[2].legendFormat,
      )
  ), gridPos={
    x: 0,
    y: 64,
    w: 12,
    h: 10,
  }
)
.addPanel(
  graphPanel.new(
    title='Pod CPU usage',
    datasource='${cluster}',
    fill=1,
    legend_avg=true,
    legend_current=true,
    legend_max=true,
    legend_values=true,
    legend_sideWidth=350,
    pointradius=2,
    value_type='individual',
  ).addSeriesOverride(
    {
      alias: 'Requests',
      color: '#73BF69',
      legend: false,
      nullPointMode: 'connected'
    }
  ).addSeriesOverride(
    {
     alias: 'Limits',
     color: '#F2495C',
     legend: false,
     nullPointMode: 'connected'
    }
  ).resetYaxes().
  addYaxis(
    format='short',
    decimals=3,
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.PodCpuUsage[0].expr,
        legendFormat=pannelTarget.PodCpuUsage[0].legendFormat,
    )
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.PodCpuUsage[1].expr,
        legendFormat=pannelTarget.PodCpuUsage[1].legendFormat,
    )
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.PodCpuUsage[2].expr,
        legendFormat=pannelTarget.PodCpuUsage[2].legendFormat,
    )
  ), gridPos={
    x: 12,
    y: 64,
    w: 12,
    h: 10,
  }
)
.addPanel(
  graphPanel.new(
    title='Goroutines',
    datasource='${cluster}',
    fill=1,
    legend_alignAsTable=true,
    legend_current=true,
    legend_rightSide=true,
    legend_values=true,
    legend_sideWidth=600,
    pointradius=2,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='short',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.Goroutines.expr,
        legendFormat=pannelTarget.Goroutines.legendFormat,
    )
  ), gridPos={
    x: 0,
    y: 74,
    w: 24,
    h: 8,
  }
)