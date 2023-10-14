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
local pannelTarget = import 'hasura_panel_target.jsonnet';
local properties =  import 'hasura_dashboard_properties.jsonnet';

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
  ), gridPos={
    h: 1,
    w: 24,
    x: 0,
    y: 0
  }
)
.addPanel(
  graphPanel.new(
    title='rpc/s',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    fill=1,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_avg=true,
    legend_current=true,
    legend_max=true,
    legend_values=true,
    lines=true,
    pointradius=2,
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
        expr=pannelTarget.Rpcs[0].expr,
        legendFormat='current',
    )
  ).addTarget(
       prometheus.custom_target(
           expr=pannelTarget.Rpcs[1].expr,
           legendFormat='last',
       )
  ), gridPos={
    h: 10,
    w: 24,
    x: 0,
    y: 1,
  }
)
.addPanel(
  graphPanel.new(
    title='Status code Rate',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    aliasColors={
     FailedPrecondition: 'semi-dark-red',
     NotFound: 'dark-orange',
     OK: 'dark-green',
    },
    bars=true,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_values=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_sort='current',
    legend_sortDesc=true,
    percentage=true,
    pointradius=2,
    stack=true,
  ).resetYaxes().
  addYaxis(
    format='short',
    max='100',
    min='0',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.StatusCodeRate.expr,
        legendFormat='{{ grpc_server_status }}',
    )
  ), gridPos={
    h: 8,
    w: 24,
    x: 0,
    y: 11,
  }
)
.addPanel(
  graphPanel.new(
    title='P90 Latency',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    unit='ms',
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_sort='max',
    legend_sortDesc=true,
    legend_values=true,
    pointradius=2,
  ).resetYaxes().
  addYaxis(
    format='ms',
  ).addYaxis(
    format='short',
  ).addSeriesOverride(
       {
         alias: 'current',
         bars: true
       }
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.LatencyByMethodP90[0].expr,
        legendFormat='current',
    )
  ).addTarget(
    prometheus.custom_target(
       expr=pannelTarget.LatencyByMethodP90[1].expr,
       legendFormat='last',
    )
  ), gridPos={
    h: 10,
    w: 24,
    x: 0,
    y: 19,
  }
)
.addPanel(
  graphPanel.new(
    title='P99 Latency',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    unit='ms',
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_sort='max',
    legend_sortDesc=true,
    legend_values=true,
    pointradius=2,
  ).resetYaxes().
  addYaxis(
    format='ms',
  ).addYaxis(
    format='short',
  ).addSeriesOverride(
       {
         alias: 'current',
         bars: true
       }
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.LatencyByMethodP99[0].expr,
        legendFormat='current',
    )
  ).addTarget(
    prometheus.custom_target(
       expr=pannelTarget.LatencyByMethodP99[1].expr,
       legendFormat='last',
    )
  ), gridPos={
    h: 10,
    w: 24,
    x: 0,
    y: 29,
  }
)
.addPanel(
  graphPanel.new(
    title='Pod CPU usage',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_values=true,
    legend_sideWidth=430,
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
  ), gridPos={
    h: 9,
    w: 24,
    x: 0,
    y: 39,
  }
)
.addPanel(
  graphPanel.new(
    title='Pod Memory usage',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    decimals=1,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_values=true,
    legend_sideWidth=430,
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
  ), gridPos={
    h: 9,
    w: 24,
    x: 0,
    y: 48,
  }
)
.addPanel(
  graphPanel.new(
    title='Receive Bytes',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    decimals=1,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_values=true,
    legend_sideWidth=430,
    pointradius=2,
  ).addSeriesOverride(
    {
      alias: 'current',
      bars: true
    }
  ).resetYaxes().
  addYaxis(
    format='bytes',
  ).addYaxis(
    format='short',
    show=false,
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.ReceiveBytes[0].expr,
        legendFormat='current',
    )
  ).addTarget(
       prometheus.custom_target(
           expr=pannelTarget.ReceiveBytes[1].expr,
           legendFormat='last',
       )
  ), gridPos={
    h: 10,
    w: 24,
    x: 0,
    y: 57,
  }
)
.addPanel(
  graphPanel.new(
    title='Sent Bytes',
    datasource={
        type: "prometheus",
        uid: "${cluster}"
    },
    decimals=1,
    legend_alignAsTable=true,
    legend_rightSide=true,
    legend_max=true,
    legend_avg=true,
    legend_current=true,
    legend_values=true,
    legend_sideWidth=430,
    pointradius=2,
  ).addSeriesOverride(
    {
      alias: 'current',
      bars: true
    }
  ).resetYaxes().
  addYaxis(
    format='bytes',
  ).addYaxis(
    format='short',
    show=false,
  ).addTarget(
    prometheus.custom_target(
        expr=pannelTarget.SentBytes[0].expr,
        legendFormat='current',
    )
  ).addTarget(
       prometheus.custom_target(
           expr=pannelTarget.SentBytes[1].expr,
           legendFormat='last',
       )
  ), gridPos={
    h: 10,
    w: 24,
    x: 0,
    y: 67,
  }
)