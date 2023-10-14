local dashboard =  import 'default_dashboard_cfg.jsonnet';
local gridPos = import 'virtual_classroom_grid_pos.jsonnet';
local properties =  import 'default_dashboard_properties.jsonnet';
local grafana = import 'grafonnet/grafana.libsonnet';
local graphPanel = grafana.graphPanel;
local pieChartPanel = grafana.pieChartPanel;
local prometheus = grafana.prometheus;
local row = grafana.row;

dashboard
.addPanel(
  row.new(
    title='Business statistics',
    titleSize='h6',
  ), gridPos=gridPos.gridPos[0]
)
.addPanel(
  graphPanel.new(
    title='Number of active rooms',
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
        expr='sum(backend_virtual_classroom_active_rooms_total{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom"})',
        legendFormat='active rooms',
    )
  ), gridPos=gridPos.gridPos[1]
)
.addPanel(
  pieChartPanel.new(
      title='Number of rooms with "update" room state requests',
      datasource=properties.datasource,
      valueName=null,
      pluginVersion='7.5.12',
  ).addThresholds(
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
        expr='sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="10"}[$__range]))',
        legendFormat='Number of rooms have from 0 to 10 request',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state",app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="20"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="10"}[$__range])))',
        legendFormat='(10, 20]',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="50"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="20"}[$__range])))',
        legendFormat='(20, 50]',
    )
  ).addTarget(
    prometheus.custom_target(
       expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="100"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="50"}[$__range])))',
       legendFormat='(50, 100]',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="200"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="100"}[$__range])))',
        legendFormat='(100, 200]',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="500"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="200"}[$__range])))',
        legendFormat='(200, 500]',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='sum(increase(backend_virtual_classroom_handled_total_count{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="500"}[$__range]))',
        legendFormat='> 500',
    )
  ), gridPos=gridPos.gridPos[2]
)
.addPanel(
  pieChartPanel.new(
      title='Number of rooms with "get" room state requests',
      datasource=properties.datasource,
      valueName=null,
      pluginVersion='7.5.12',
  ).addThresholds(
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
        expr='sum(increase(backend_virtual_classroom_handled_total_bucket{action="update_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="10"}[$__range]))',
        legendFormat='Number of rooms have from 0 to 10 request',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="500"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="100"}[$__range])))',
        legendFormat='(100, 500]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="1000"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="500"}[$__range])))',
        legendFormat='(500, 1000]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="3000"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="1000"}[$__range])))',
        legendFormat='(1000, 3000]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="5000"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="3000"}[$__range])))',
        legendFormat='(3000, 5000]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="7000"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="5000"}[$__range])))',
        legendFormat='(5000, 7000]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='sum(increase(backend_virtual_classroom_handled_total_count{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace"}[$__range])) - sum(increase(backend_virtual_classroom_handled_total_bucket{action="get_room_state", app_kubernetes_io_name=~"bob|virtualclassroom", namespace="$namespace", le="7000"}[$__range]))',
        legendFormat='> 7000',
    )
   ), gridPos=gridPos.gridPos[3]
)
.addPanel(
  pieChartPanel.new(
      title='Number of rooms with total "attendees" who really joined',
      datasource=properties.datasource,
      valueName=null,
      pluginVersion='7.5.12',
  ).addThresholds(
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
        expr='sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="5"}[$__range]))',
        legendFormat='Number of rooms have from 0 to 5 attendees',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="10"}[$__range])) -
              sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="5"}[$__range])))',
        legendFormat='(5, 10]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="15"}[$__range]))
              - sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="10"}[$__range])))',
        legendFormat='(10, 15]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="20"}[$__range]))
              - sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="15"}[$__range])))',
        legendFormat='(15, 20]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="40"}[$__range]))
              - sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="20"}[$__range])))',
        legendFormat='(20, 40]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='sum(increase(backend_virtual_classroom_attendees_total_count{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom",}[$__range])) -
              sum(increase(backend_virtual_classroom_attendees_total_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="40"}[$__range]))',
        legendFormat='> 40',
    )
   ), gridPos=gridPos.gridPos[4]
)
.addPanel(
  pieChartPanel.new(
      title='Number of rooms with total real "live time"',
      datasource=properties.datasource,
      valueName=null,
      pluginVersion='7.5.12',
  ).addThresholds(
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
        expr='sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="10"}[$__range]))',
        legendFormat='Number of rooms have real live time from 0 to 10 mins',
    )
  ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="20"}[$__range])) - sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{ namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="10"}[$__range])))',
        legendFormat='(10, 20]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="40"}[$__range]))
              - sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="20"}[$__range])))',
        legendFormat='(20, 40]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="60"}[$__range]))
              - sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="40"}[$__range])))',
        legendFormat='(40, 60]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='(sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="90"}[$__range]))
              - sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom", le="60"}[$__range])))',
        legendFormat='(60, 90]',
    )
   ).addTarget(
    prometheus.custom_target(
        expr='sum(increase(backend_virtual_classroom_total_real_live_time_minutes_count{namespace="$namespace", app_kubernetes_io_name=~"bob|virtualclassroom"}[$__range])) -
              sum(increase(backend_virtual_classroom_total_real_live_time_minutes_bucket{namespace="$namespace",app_kubernetes_io_name=~"bob|virtualclassroom", le="90"}[$__range]))',
        legendFormat='> 90',
    )
   ), gridPos=gridPos.gridPos[5]
)