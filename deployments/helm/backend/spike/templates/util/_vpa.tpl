{{/*
vpa defines a VerticalPodAutoscaler for the current service's Deployment.

It should be disabled in local, as VerticalPodAutoscaler CRD is not installed there.
*/}}
{{- define "util.vpa" -}}
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: {{ include "util.fullname" . }}
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: {{ include "util.fullname" . }}
  updatePolicy:
    updateMode: {{ include "util.vpaUpdatePolicy" . | quote }}
{{- end }}

{{/*
Similar to util.vpa, but for hasura deployments.
*/}}
{{- define "util.hasuraVpa" -}}
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: {{ include "util.fullname" . }}-hasura
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: {{ include "util.fullname" . }}-hasura
  updatePolicy:
    updateMode: {{ include "util.vpaUpdatePolicy" . | quote }}
{{- end }}

{{/*
Similar to util.vpa, but for hasura v2 deployments.
*/}}
{{- define "util.hasurav2Vpa" -}}
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: {{ include "util.fullname" . }}-hasurav2
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: {{ include "util.fullname" . }}-hasurav2
  updatePolicy:
    updateMode: {{ include "util.vpaUpdatePolicy" . | quote }}
{{- end }}

{{/*
vpaUpdatePolicy returns the first non-empty value of:
  - .Values.vpa.UpdateMode
  - .Values.global.vpa.UpdateMode
  -  "Off", otherwise
*/}}
{{- define "util.vpaUpdatePolicy" -}}
{{- $vpaUpdateMode := "Off" -}}
{{- if .Values.vpa -}}
  {{- if .Values.vpa.updateMode -}}
    {{- $vpaUpdateMode = .Values.vpa.updateMode -}}
  {{- else if .Values.global.vpa -}}
    {{- if .Values.global.vpa.updateMode -}}
      {{- $vpaUpdateMode = .Values.global.vpa.updateMode -}}
    {{- end -}}
  {{- end -}}
{{- else if .Values.global.vpa -}}
  {{- if .Values.global.vpa.updateMode -}}
    {{- $vpaUpdateMode = .Values.global.vpa.updateMode -}}
  {{- end -}}
{{- end -}}
{{ $vpaUpdateMode }}
{{- end -}}
