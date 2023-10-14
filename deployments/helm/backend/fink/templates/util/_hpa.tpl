{{/*
Returns the appropriate apiVersion for Horizontal Pod Autoscaler.
*/}}
{{- define "util.autoscaling.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "autoscaling/v2" -}}
{{- print "autoscaling/v2" -}}
{{- else -}}
{{- print "autoscaling/v2beta2" -}}
{{- end -}}
{{- end -}}

{{/*
Returns HPA for normal application deployments.
*/}}
{{- define "util.hpa" -}}
{{- if or (.Capabilities.APIVersions.Has "autoscaling/v2") (.Capabilities.APIVersions.Has "autoscaling/v2beta2") }}
apiVersion: {{ include "util.autoscaling.apiVersion" . }}
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "util.fullname" . }}
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "util.fullname" . }}
  minReplicas: {{ .Values.hpa.minReplicas }}
  maxReplicas: {{ .Values.hpa.maxReplicas }}
  metrics:
{{- with .Values.hpa.averageCPUUtilization }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ . }}
{{- end }}
{{- with .Values.hpa.averageMemoryValue }}
    - type: Resource
      resource:
        name: memory
        target:
          type: AverageValue
          averageValue: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Returns HPA for hasura deployments.
*/}}
{{- define "util.hasuraHpa" -}}
{{- if or (.Capabilities.APIVersions.Has "autoscaling/v2") (.Capabilities.APIVersions.Has "autoscaling/v2beta2") }}
apiVersion: {{ include "util.autoscaling.apiVersion" . }}
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "util.fullname" . }}-hasura
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "util.fullname" . }}-hasura
  minReplicas: {{ .Values.hasura.hpa.minReplicas }}
  maxReplicas: {{ .Values.hasura.hpa.maxReplicas }}
  metrics:
{{- with .Values.hasura.hpa.averageCPUUtilization }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ . }}
{{- end }}
{{- with .Values.hasura.hpa.averageMemoryValue }}
    - type: Resource
      resource:
        name: memory
        target:
          type: AverageValue
          averageValue: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Returns HPA for hasura v2 deployments.
*/}}
{{- define "util.hasurav2Hpa" -}}
{{- if or (.Capabilities.APIVersions.Has "autoscaling/v2") (.Capabilities.APIVersions.Has "autoscaling/v2beta2") }}
apiVersion: {{ include "util.autoscaling.apiVersion" . }}
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "util.fullname" . }}-hasurav2
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "util.fullname" . }}-hasurav2
  minReplicas: {{ .Values.hasurav2.hpa.minReplicas }}
  maxReplicas: {{ .Values.hasurav2.hpa.maxReplicas }}
  metrics:
{{- with .Values.hasurav2.hpa.averageCPUUtilization }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ . }}
{{- end }}
{{- with .Values.hasurav2.hpa.averageMemoryValue }}
    - type: Resource
      resource:
        name: memory
        target:
          type: AverageValue
          averageValue: {{ . }}
{{- end }}
{{- end }}
{{- end }}
