{{/*
Expand the name of the chart.
*/}}
{{- define "kserve.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kserve.fullname" -}}
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
{{- define "kserve.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kserve.labels" -}}
helm.sh/chart: {{ include "kserve.chart" . }}
{{ include "kserve.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kserve.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kserve.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kserve.serviceAccountName" -}}
{{- if eq "dorp" .Values.environment -}}
{{- printf "prod-%s" (include "kserve.fullname" .) }}
{{- else if and (eq "stag" .Values.environment) (eq "jprep" .Values.vendor) }}
{{- printf "stag-jprep-%s" (include "kserve.fullname" .) }}
{{- else }}
{{- printf "%s-%s" .Values.environment (include "kserve.fullname" .) }}
{{- end }}
{{- end }}

{{- define "kserve.serviceAccountEmail" -}}
{{- printf "%s@%s.iam.gserviceaccount.com" (include "kserve.serviceAccountName" .) .Values.serviceAccountEmailSuffix }}
{{- end }}

{{/*
Annotate service account
*/}}
{{- define "kserve.serviceAccountAnnotations" -}}
iam.gke.io/gcp-service-account: {{ include "kserve.serviceAccountEmail" . }}
{{- end }}

{{/*
Service credential name secrect
*/}}
{{- define "kserve.secrectName" -}}
{{- default (include "kserve.fullname" .) .Values.serviceAccount.name }}
{{- end -}}
