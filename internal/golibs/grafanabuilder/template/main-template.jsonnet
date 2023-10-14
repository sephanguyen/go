local dashboard =  import 'default_dashboard_cfg.jsonnet';
local properties =  import 'default_dashboard_properties.jsonnet';
local gridPos = import 'grid-pos-template.jsonnet';
local grafana = import 'grafonnet/grafana.libsonnet';
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local row = grafana.row;
// you can also import other pannels if you need, see more "internal/golibs/grafanabuilder/grafonnet"

/*
dashboard
.addPanel(
  row.new(
    title='Title Row',
    titleSize='h6',
  ), gridPos=gridPos.gridPos[0]
).
addPanel(
  graphPanel.new(
    title='title',
    datasource=properties.datasource,
    fill=1,
    legend_show=false,
    pointradius=2,
    value_type='individual',
  ).resetYaxes().
  addYaxis(
    format='none',
  ).addYaxis(
    format='short',
  ).addTarget(
    prometheus.custom_target(
        expr='expr',
        legendFormat='legendFormat',
    )
  ), gridPos=gridPos.gridPos[1]
)
*/