{{/*
Expand the name of the chart.
*/}}
{{- define "import-map-deployer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "import-map-deployer.fullname" -}}
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
{{- define "import-map-deployer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "import-map-deployer.labels" -}}
helm.sh/chart: {{ include "import-map-deployer.chart" . }}
{{ include "import-map-deployer.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "import-map-deployer.selectorLabels" -}}
app.kubernetes.io/name: {{ include "import-map-deployer.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "import-map-deployer.serviceAccountName" -}}
{{- if eq .Values.global.environment "dorp" -}}
{{- printf "prod-import-map-deployer" }}
{{- else }}
{{- printf "%s-import-map-deployer" .Values.global.environment }}
{{- end }}
{{- end }}

{{/*
Annotate service account
*/}}
{{- define "import-map-deployer.serviceAccountAnnotations" -}}
iam.gke.io/gcp-service-account: {{ printf "%s@%s.iam.gserviceaccount.com" (include "import-map-deployer.serviceAccountName" .) .Values.serviceAccountEmailSuffix }}
{{- end }}
