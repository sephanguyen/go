{{- define "util.workloadMetadata" }}
annotations:
  checksum/{{ .Chart.Name }}.common.config.yaml: {{ tpl ("configs/{{ .Chart.Name }}.common.config.yaml" | .Files.Get) . | sha256sum }}
  checksum/{{ .Chart.Name }}.config.yaml: {{ tpl (printf "configs/%s/%s/{{ .Chart.Name }}.config.yaml" .Values.global.vendor .Values.global.environment | .Files.Get) . | sha256sum }}
{{- if eq "local" .Values.global.environment }}
  checksum/service_credential.json: {{ include "util.serviceCredential" . | sha256sum }}
{{- end }}
  checksum/{{ .Chart.Name }}.secrets.encrypted.yaml: {{ printf "secrets/%s/%s/{{ .Chart.Name }}.secrets.encrypted.yaml" .Values.global.vendor .Values.global.environment | .Files.Get | sha256sum }}
{{- if .Values.migrationEnabled }}
  checksum/{{ .Chart.Name }}_migrate.secrets.encrypted.yaml: {{ printf "secrets/%s/%s/{{ .Chart.Name }}_migrate.secrets.encrypted.yaml" .Values.global.vendor .Values.global.environment | .Files.Get | sha256sum }}
{{- end }}
{{- if .Values.hasuraEnabled }}
  checksum/hasura-metadata: {{ (.Files.Glob "files/hasura/metadata/*").AsConfig | sha256sum }}
{{- end }}
  cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
{{- if .Values.podAnnotations }}
{{ toYaml .Values.podAnnotations | indent 2 }}
{{- else }}
  sidecar.istio.io/proxyCPU: "10m"
  sidecar.istio.io/proxyMemory: "50Mi"
{{- end }}
{{- if .Values.metrics }}
{{- if .Values.metrics.enabled }}
{{- if .Values.metrics.podAnnotations }}
{{ toYaml .Values.metrics.podAnnotations | indent 2 }}
{{- else }}
  prometheus.io/scheme: "http"
  prometheus.io/port: "8888"
  prometheus.io/scrape: "true"
{{- end }}
{{- end }}
{{- end }}
labels:
  app.kubernetes.io/name: {{ .Chart.Name }}
  app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
