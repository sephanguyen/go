{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "yugabyte.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "yugabyte.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "yugabyte.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "yugabyte.labels" -}}
helm.sh/chart: {{ include "yugabyte.chart" . }}
{{ include "yugabyte.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "yugabyte.selectorLabels" -}}
app.kubernetes.io/name: {{ include "yugabyte.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "yugabyte.serviceAccountName" -}}
{{ printf "%s-%s" .Values.environment (include "yugabyte.fullname" .) }}
{{- end }}

{{/*
Annotate service account
*/}}
{{- define "yugabyte.serviceAccountAnnotations" -}}
iam.gke.io/gcp-service-account: {{ printf "%s@%s.iam.gserviceaccount.com" (include "yugabyte.serviceAccountName" .) .Values.serviceAccountEmailSuffix }}
{{- end }}

{{/*
Get YugaByte master addresses
*/}}
{{- define "yugabyte.masterAddresses" -}}
  {{- $fullname := include "yugabyte.fullname" . }}
  {{- $masterReplicas := .Values.replicas.master | int -}}
  {{- $prefix := "" -}}
  {{- range $index := until $masterReplicas -}}
    {{- if ne $index 0 }},{{ end -}}
    {{- $prefix }}{{ $fullname }}-{{ $index }}.{{ $fullname }}-headless.{{ $.Release.Namespace }}.svc.cluster.local:7100
  {{- end -}}
{{- end -}}

{{/*
{{- $prefix }}{{ $fullname }}-master-{{ $index }}.{{ $fullname }}-master-headless.{{ $.Release.Namespace }}.svc.cluster.local:7100
*/}}

{{/*
See https://github.com/yugabyte/charts/blob/master/stable/yugabyte/templates/_helpers.tpl#L27
*/}}
{{- define "yugabyte.memoryHardLimit" -}}
{{- printf "%d" .limits.memory | regexFind "\\d+" | mul 1024 | mul 1024 | mul 870 }}
{{- end -}}
