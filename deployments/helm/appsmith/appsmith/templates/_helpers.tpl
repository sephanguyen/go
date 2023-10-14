{{/*
Expand the name of the chart.
*/}}
{{- define "appsmith.name" -}}
{{- default .Chart.Name .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "appsmith.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.fullnameOverride }}
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
{{- define "appsmith.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "appsmith.labels" -}}
appsmith.sh/chart: {{ include "appsmith.chart" . }}
{{ include "appsmith.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "appsmith.selectorLabels" -}}
app.kubernetes.io/name: {{ include "appsmith.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Allow the release namespace to be overridden for multi-namespace deployments in combined charts.
*/}}
{{- define "appsmith.namespace" -}}
    {{- if .Values.global -}}
        {{- if .Values.global.namespaceOverride }}
            {{- .Values.global.namespaceOverride -}}
        {{- else -}}
            {{- .Release.Namespace -}}
        {{- end -}}
    {{- else -}}
        {{- .Release.Namespace -}}
    {{- end }}
{{- end -}}

{{/*
Kubernetes standard labels
*/}}
{{- define "labels.standard" -}}
app.kubernetes.io/name: {{ include "names.name" . }}
helm.sh/chart: {{ include "names.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Expand the name of the chart.
*/}}
{{- define "names.name" -}}
{{- default .Chart.Name .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "names.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Return  the proper Storage Class
*/}}
{{- define "storage.class" -}}

{{- $storageClass := .persistence.storageClass -}}
{{- if .global -}}
    {{- if .global.storageClass -}}
        {{- $storageClass = .global.storageClass -}}
    {{- end -}}
{{- end -}}

{{- if $storageClass -}}
  {{- if (eq "-" $storageClass) -}}
      {{- printf "storageClassName: \"\"" -}}
  {{- else }}
      {{- printf "storageClassName: %s" $storageClass -}}
  {{- end -}}
{{- end -}}

{{- end -}}

{{/*
Renders a value that contains template.
*/}}
{{- define "tplvalues.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{- define "virtualservice.appsmith.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-appsmith
spec:
  hosts:
{{ toYaml .Values.dnsNames.appsmith | indent 4 }}
  gateways:
    - istio-system/{{ .Values.environment }}-{{ .Values.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.appsmithHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{/*
This defines istio virtualservice that works with internal.staging.manabie.io and similar domains.
*/}}
{{- define "virtualservice.internal.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-internal
spec:
  hosts:
{{ toYaml .Values.dnsNames.internal | indent 4 }}
  gateways:
    - istio-system/{{ .Values.environment }}-{{ .Values.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.internalHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{/*
This defines istio virtualservice that works with internal.uat.manabie.io and similar domains.
*/}}
{{- define "virtualservice.uatInternal.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-uat-internal
spec:
  hosts:
{{ toYaml .Values.dnsNames.uatInternal | indent 4 }}
  gateways:
    - istio-system/{{ .Values.environment }}-{{ .Values.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.uatInternalHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{- define "virtualservice.internalTool.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-prod-internaltool
spec:
  hosts:
{{ toYaml .Values.dnsNames.internalTool | indent 4 }}
  gateways:
    - istio-system/{{ .Values.environment }}-{{ .Values.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.internalToolHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{- define "appsmith.serviceAccountName" -}}
{{- if eq "dorp" .Values.environment -}}
{{- printf "prod-%s" (include "appsmith.fullname" .) }}
{{- else -}}
{{- printf "%s-%s" .Values.environment (include "appsmith.fullname" .) }}
{{- end -}}
{{- end -}}


{{- define "appsmith.serviceAccountEmail" -}}
{{- printf "%s@%s.iam.gserviceaccount.com" (include "appsmith.serviceAccountName" .) .Values.serviceAccountEmailSuffix }}
{{- end -}}

{{/*
Annotate service account.
Referenced from "./deployments/helm/manabie-all-in-one/templates/_serviceaccount.yaml"
*/}}
{{- define "appsmith.serviceAccountAnnotations" -}}
iam.gke.io/gcp-service-account: {{ include "appsmith.serviceAccountEmail" . }}
{{- end -}}
