{{/*
Template for Keda ScaledObject.
It's using Keda cron trigger to scale the deployment on
on-demand node based on the cron schedule.
*/}}
{{- define "util.keda.scaledObject" -}}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: {{ include "util.fullname" . }}-on-demand-node
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  minReplicaCount: 0
  cooldownPeriod: 300
  pollingInterval: 30
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "util.fullname" . }}-on-demand-node
  triggers:
{{- range (default .Values.global.onDemandNodeDeployment.cronScheduledScaling .Values.onDemandNodeDeployment.cronScheduledScaling) }}
  - type: cron
    metadata:
      timezone: {{ .timezone }}
      start: {{ .start }}
      end: {{ .end }}
      desiredReplicas: {{ .desiredReplicas | quote }}
{{- end }}
{{- end -}}

{{/*
Template for Keda ScaledObject.
It's using Keda cron trigger to scale the Hasura deployment on
on-demand node based on the cron schedule.
*/}}
{{- define "util.keda.hasuraScaledObject" -}}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: {{ include "util.fullname" . }}-hasura-on-demand-node
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  minReplicaCount: 0
  cooldownPeriod: 300
  pollingInterval: 30
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "util.fullname" . }}-hasura-on-demand-node
  triggers:
{{- range (default .Values.global.hasura.onDemandNodeDeployment.cronScheduledScaling .Values.hasura.onDemandNodeDeployment.cronScheduledScaling) }}
  - type: cron
    metadata:
      timezone: {{ .timezone }}
      start: {{ .start }}
      end: {{ .end }}
      desiredReplicas: {{ .desiredReplicas | quote }}
{{- end }}
{{- end -}}

{{/*
This is similar to `util.keda.hasuraScaledObject` except that it targets hasurav2 deployments.
However, it still uses the configs from hasura.
*/}}
{{- define "util.keda.hasurav2ScaledObject" -}}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: {{ include "util.fullname" . }}-hasurav2
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "util.fullname" . }}-hasurav2
  minReplicaCount: {{ default 1 .Values.global.kedaScaledObjectMinReplicas }}
  maxReplicaCount: {{ default 2 .Values.global.kedaScaledObjectMaxReplicas }}
  triggers:
{{- range (default .Values.global.hasura.cronScheduledScaling .Values.hasura.cronScheduledScaling) }}
  - type: cron
    metadata:
      timezone: {{ .timezone }}
      start: {{ .start }}
      end: {{ .end }}
      desiredReplicas: {{ .desiredReplicas | quote }}
{{- end }}
{{- if .Values.hasura.hpa }}
{{- with .Values.hasura.hpa.averageCPUUtilization }}
  - type: cpu
    metricType: Utilization
    metadata:
      value: {{ . | quote }}
{{- end }}
{{- with .Values.hasura.hpa.averageMemoryValue }}
  - type: memory
    metricType: AverageValue
    metadata:
      value: {{ . | quote }}
{{- end }}
{{- end }}
{{- end -}}
