{{/*
Expand the name of the chart.
*/}}
{{- define "infras.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "infras.fullname" -}}
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
{{- define "infras.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "infras.labels" -}}
helm.sh/chart: {{ include "infras.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "virtualservice.minio.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: minio-{{ .Chart.Name }}
spec:
  hosts:
{{ toYaml .Values.dnsNames.minio | indent 4 }}
  gateways:
    - istio-system/{{ .Values.minio.environment }}-{{ .Values.minio.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.minio.adminHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end }}
