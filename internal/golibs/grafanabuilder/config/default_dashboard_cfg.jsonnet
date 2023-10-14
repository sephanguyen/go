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
local pannelTarget = import 'panel_target.jsonnet';
local properties =  import 'default_dashboard_properties.jsonnet';

dashboard.new(
  properties.title,
  editable=true,
  refresh='5s',
  timepicker={},
  schemaVersion=27,
  uid=properties.uid,
  graphTooltip='shared_crosshair',
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
  ), gridPos=properties.gridPos[0]
)
.addPanel(
  graphPanel.new(
    title='Requests per seconds',
    datasource=properties.datasource,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_min=true,
    legend_current=true,
    legend_avg=true,
    legend_sortDesc=true,
    legend_sort='max',
    legend_values=true,
    legend_sideWidth=350,
    pointradius=2,
    unit='short',
  ).resetYaxes().
  addYaxis(
    format='short',
  ).addYaxis(
    format='short',
  ).addSeriesOverride(
    {
      alias: 'current',
      bars: true
    }
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.RequestsPerSeconds[0].expr,
        legendFormat='current',
    )
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.RequestsPerSeconds[1].expr,
        legendFormat='last',
    )
  ), gridPos=properties.gridPos[1]
)
.addPanel(
  graphPanel.new(
    title='Successful request rate by methods',
    datasource=properties.datasource,
    unit='percentunit',
    legend_alignAsTable=true,
    legend_hideEmpty=true,
    legend_hideZero=true,
    legend_min=true,
    legend_rightSide=true,
    legend_sort='min',
    legend_values=true,
    pointradius=2,
  ).resetYaxes().
  addYaxis(
    format='percentunit',
    max='1',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.SuccessfulRequestRateByMethods.expr,
        legendFormat='{{ grpc_server_method }}',
    )
  ), gridPos=properties.gridPos[2]
)
.addPanel(
  graphPanel.new(
    title='gRPC Response status',
    datasource=properties.datasource,
    aliasColors={
     FailedPrecondition: 'semi-dark-red',
     NotFound: 'dark-orange',
     OK: 'dark-green',
    },
    bars=true,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_hideEmpty=true,
    legend_hideZero=true,
    legend_values=true,
    legend_min=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    lines=false,
    percentage=true,
    pointradius=2,
    stack=true,
  ).resetYaxes().
  addYaxis(
    format='percentunit',
    max='100',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.GRPCResponseStatus.expr,
        legendFormat='{{ grpc_server_status }}',
    )
  ), gridPos=properties.gridPos[3]
)
.addPanel(
  graphPanel.new(
    title='Latency by Method (P90)',
    datasource=properties.datasource,
    unit='ms',
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_sort='max',
    legend_sortDesc=true,
    legend_values=true,
    pointradius=2,
  ).resetYaxes().
  addYaxis(
    format='ms',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.LatencyByMethodP90.expr,
        legendFormat='{{ grpc_server_method }}',
    )
  ), gridPos=properties.gridPos[4]
)
.addPanel(
  graphPanel.new(
    title='Latency by Method (P95)',
    datasource=properties.datasource,
    unit='ms',
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_sort='max',
    legend_sortDesc=true,
    legend_values=true,
    pointradius=2,
  ).resetYaxes().
  addYaxis(
    format='ms',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.LatencyByMethodP95.expr,
        legendFormat='{{ grpc_server_method }}',
    )
  ), gridPos=properties.gridPos[5]
)
.addPanel(
  graphPanel.new(
    title='Pod Memory usage',
    datasource=properties.datasource,
    decimals=1,
    legend_avg=true,
    legend_current=true,
    legend_max=true,
    legend_values=true,
    legend_sideWidth=350,
    pointradius=2,
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
        format='time_series',
        legendFormat='{{ pod }}',
        intervalFactor=1,
    )
  ).addTarget(
       prometheus.custom_target(
           expr=pannelTarget.PodMemoryUsage[1].expr,
           legendFormat='Requests',
       )
  ).addTarget(
      prometheus.custom_target(
           expr=pannelTarget.PodMemoryUsage[2].expr,
           legendFormat='Limits',
      )
  ), gridPos=properties.gridPos[6]
)
.addPanel(
  graphPanel.new(
    title='Pod CPU usage',
   datasource=properties.datasource,
    legend_avg=true,
    legend_current=true,
    legend_max=true,
    legend_values=true,
    legend_sideWidth=350,
    pointradius=2,
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
        legendFormat='{{ pod }}',
    )
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.PodCpuUsage[1].expr,
        legendFormat='Requests',
    )
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.PodCpuUsage[2].expr,
        legendFormat='Limits',
    )
  ), gridPos=properties.gridPos[7]
)
.addPanel(
  graphPanel.new(
    title='Goroutines',
    datasource=properties.datasource,
    legend_alignAsTable=true,
    legend_current=true,
    legend_rightSide=true,
    legend_values=true,
    legend_sideWidth=600,
    pointradius=2,
  ).resetYaxes().
  addYaxis(
    format='short',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.Goroutines.expr,
        legendFormat='{{ pod }}',
    )
  ), gridPos=properties.gridPos[8]
)