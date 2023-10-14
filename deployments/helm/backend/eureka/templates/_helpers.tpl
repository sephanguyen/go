{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "eureka.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "eureka.fullname" -}}
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
{{- define "eureka.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "eureka.labels" -}}
helm.sh/chart: {{ include "eureka.chart" . }}
{{ include "eureka.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "eureka.allConsumersLabels" -}}
helm.sh/chart: {{ include "eureka.chart" . }}
{{ include "eureka.selectorAllConsumersLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "eureka.JPREPSyncCourseStudentLabels" -}}
helm.sh/chart: {{ include "eureka.chart" . }}
{{ include "eureka.selectorAllConsumersLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "eureka.monitorsLabels" -}}
helm.sh/chart: {{ include "eureka.chart" . }}
{{ include "eureka.selectorMonitorsLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "eureka.selectorLabels" -}}
app.kubernetes.io/name: {{ include "eureka.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "eureka.selectorAllConsumersLabels" -}}
app.kubernetes.io/name: {{ include "eureka.name" . }}-all-consumers
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: all-consumers
{{- end }}

{{- define "eureka.selectorJPREPSyncCourseStudentLabels" -}}
app.kubernetes.io/name: {{ include "eureka.name" . }}-jprep-sync-course-student
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: all-consumers
{{- end }}

{{- define "eureka.selectorMonitorsLabels" -}}
app.kubernetes.io/name: {{ include "eureka.name" . }}-monitors
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: monitors
{{- end }}
