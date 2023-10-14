{
  /**
   * Create a [bar gauge panel](https://grafana.com/docs/grafana/latest/panels/visualizations/bar-gauge-panel/),
   *
   * @name barGaugePanel.new
   *
   * @param title Panel title.
   * @param description (optional) Panel description.
   * @param datasource (optional) Panel datasource.
   * @param unit (optional) The unit of the data.
   * @param thresholds (optional) An array of threashold values.
   *
   * @method addTarget(target) Adds a target object.
   * @method addTargets(targets) Adds an array of targets.
   */
  new(
    title,
    description=null,
    datasource=null,
    unit=null,
    fieldConfigColorMode='thresholds',
    max=null,
    min=null,
    thresholdsMode='absolute',
    displayMode='gradient',
    orientation='horizontal',
    allValues=false,
    reducerFunction='lastNotNull',
    fields='',
    showUnfilled=false,
    titleSize=null,
  ):: {
    type: 'bargauge',
    title: title,
    [if description != null then 'description']: description,
    datasource: datasource,
    targets: [
    ],
    _nextTarget:: 0,
    fieldConfig: {
      defaults: {
        color: {
          mode: fieldConfigColorMode,
        },
        [if max != null then 'max']: max,
        [if min != null then 'min']: min,
        unit: unit,
        thresholds: {
          mode: thresholdsMode,
        },
      },
    },
    options: {
      displayMode: displayMode,
      orientation: orientation,
      reduceOptions: {
          values: allValues,
          calcs: [
            reducerFunction,
          ],
          fields: fields,
      },
      showUnfilled: showUnfilled,
      [if titleSize != null then 'text']: {
        titleSize:  titleSize,
      },
    },
    // thresholds
    addThreshold(step):: self {
        fieldConfig+: { defaults+: { thresholds+: { steps+: [step] } } },
    },
    addThresholds(steps):: std.foldl(function(p, s) p.addThreshold(s), steps, self),
    addTarget(target):: self {
      // automatically ref id in added targets.
      local nextTarget = super._nextTarget,
      _nextTarget: nextTarget + 1,
      targets+: [target { refId: std.char(std.codepoint('A') + nextTarget) }],
    },
    addTargets(targets):: std.foldl(function(p, t) p.addTarget(t), targets, self),
  },
}
