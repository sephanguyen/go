{{/*
Expand the name of the chart.
*/}}
{{- define "elastic.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "elastic.fullname" -}}
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
{{- define "elastic.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "elastic.labels" -}}
helm.sh/chart: {{ include "elastic.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Get node addresses
*/}}
{{- define "elastic.nodeAddresses" -}}
  {{- $fullname := include "elastic.fullname" . }}
  {{- $replicas := .Values.elasticsearch.replicas | int -}}
  {{- $prefix := "elasticsearch-" -}}
  {{- range $index := until $replicas -}}
    {{- if ne $index 0 }},{{ end -}}
    {{- $prefix }}{{ $fullname }}-{{ $index }}.{{ $prefix }}{{ $fullname }}-headless.{{ $.Release.Namespace }}.svc.cluster.local
  {{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "elastic.serviceAccountName" -}}
{{- if eq "dorp" .Values.environment -}}
{{- printf "prod-elasticsearch" }}
{{- else if and (eq "stag" .Values.environment) (eq "jprep" .Values.vendor) }}
{{- "stag-jprep-elasticsearch" }}
{{- else if and (eq "prod" .Values.environment) (eq "jprep" .Values.vendor) }}
{{- "prod-jprep-elasticsearch" }}
{{- else }}
{{- printf "%s-elasticsearch" .Values.environment }}
{{- end }}
{{- end }}

{{/*
Annotate service account
*/}}
{{- define "elastic.serviceAccountAnnotations" -}}
iam.gke.io/gcp-service-account: {{ printf "%s@%s.iam.gserviceaccount.com" (include "elastic.serviceAccountName" .) .Values.serviceAccountEmailSuffix }}
{{- end }}

{{/*
Value for admin_dn. Customized for preproduction.
For more information, see: https://manabie.slack.com/archives/C025EN333K8/p1653383023222469?thread_ts=1653381819.426799&cid=C025EN333K8
*/}}
{{- define "elastic.adminDn" -}}
{{- if eq "dorp" .Values.environment -}}
{{- printf "admin.%s.prod.search" .Values.vendor }}
{{- else -}}
{{- printf "admin.%s.%s.search" .Values.vendor .Values.environment }}
{{- end -}}
{{- end }}

{{/*
Address to shamir service. Used for authn.
*/}}
{{- define "elastic.shamirAddress" -}}
{{- printf "shamir.%s-%s-backend.svc.cluster.local:5680" .Values.environment .Values.vendor -}}
{{- end -}}
