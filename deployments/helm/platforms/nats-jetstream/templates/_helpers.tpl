{{/*
Expand the name of the chart.
*/}}
{{- define "nats-jetstream.name" }}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nats-jetstream.fullname" -}}
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
{{- define "nats-jetstream.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nats-jetstream.labels" -}}
helm.sh/chart: {{ include "nats-jetstream.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nats-jetstream.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nats-jetstream.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/*
Return the Nats JetStream cluster advertise
*/}}
{{- define "nats-jetstream.clusterAdvertise" -}}
{{- printf "$(POD_NAME).%s-headless.$(POD_NAMESPACE).svc.cluster.local" (include "nats-jetstream.fullname" . ) }}
{{- end }}

{{/*
Return the Nats JetStream cluster routes
*/}}
{{- define "nats-jetstream.clusterRoutes" -}}
{{- $name := (include "nats-jetstream.fullname" . ) }}
{{- range $i, $e := until (.Values.jetstream.cluster.replicas | int) -}}
{{- printf "nats://%s-%d.%s-headless.%s.svc.cluster.local:6223," $name $i $name $.Release.Namespace -}}
{{- end -}}
{{- end }}

{{/*
Nats url
*/}}
{{- define "nats-jetstream.NatsURL" -}}
{{- printf "nats://nats-jetstream.%s.svc.cluster.local:4223" $.Release.Namespace -}}
{{- end -}}

{{/*
Create the name of the service account to use.
Referenced from "./deployments/helm/manabie-all-in-one/templates/_serviceaccount.yaml"
*/}}
{{- define "nats-jetstream.serviceAccountName" -}}
{{- if eq "dorp" .Values.environment -}}
{{- printf "prod-%s" (include "nats-jetstream.fullname" .) }}
{{- else if and (eq "prod" .Values.environment) (eq "jprep" .Values.vendor) -}}
{{- printf "prod-jprep-%s" (include "nats-jetstream.fullname" .) }}
{{- else -}}
{{- printf "%s-%s" .Values.environment (include "nats-jetstream.fullname" .) }}
{{- end -}}
{{- end -}}

{{- define "nats-jetstream.serviceAccountEmail" -}}
{{- printf "%s@%s.iam.gserviceaccount.com" (include "nats-jetstream.serviceAccountName" .) .Values.serviceAccountEmailSuffix }}
{{- end -}}

{{/*
Annotate service account.
Referenced from "./deployments/helm/manabie-all-in-one/templates/_serviceaccount.yaml"
*/}}
{{- define "nats-jetstream.serviceAccountAnnotations" -}}
iam.gke.io/gcp-service-account: {{ include "nats-jetstream.serviceAccountEmail" . }}
{{- end -}}
