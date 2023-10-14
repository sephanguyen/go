{{/*
util.configMap is the template for the k8s config map object for most manabie-all-in-one
business services. It comprises of common config and env/org-specific config.

modd.conf is for live-reloading, which should only be used in local development.

For more information, see https://manabie.atlassian.net/wiki/spaces/TECH/pages/503940027/Config+Secret+management
*/}}
{{- define "util.configMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "util.fullname" . }}
{{- if or .Values.preHookUpsertStream .Values.preHookUpsertKafkaTopic }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade,post-upgrade
    "helm.sh/hook-weight": "-48"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
{{- end }}
  labels:
    {{- include "util.labels" . | nindent 4 }}
data:
  {{ .Chart.Name }}.common.config.yaml: |
{{- tpl (printf "configs/%s.common.config.yaml" .Chart.Name | .Files.Get) . | nindent 4 }}
  {{ .Chart.Name }}.config.yaml: |
{{ tpl (printf "configs/%s/%s/%s.config.yaml" .Values.global.vendor .Values.global.environment .Chart.Name | .Files.Get) . | indent 4 }}
{{- end -}}

{{- define "util.hasuraConfigMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "util.fullname" . }}-hasura-metadata
  labels:
    {{- include "util.labels" . | nindent 4 }}
data:
{{- range $path, $_ := .Files.Glob "files/hasura/metadata/*" }}
{{- if not (hasSuffix "_stg.yaml" $path) }}
  {{ $path | base }}: |
{{- if and (eq "stag" (include "util.environment" $)) (eq "manabie" (include "util.vendor" $)) (eq "bob" $.Chart.Name) (hasSuffix "tables.yaml" $path) }}
{{ $.Files.Get "files/hasura/metadata/tables_stg.yaml" | indent 4 }}
{{- else if and (eq "stag" (include "util.environment" $)) (eq "manabie" (include "util.vendor" $)) (eq "bob" $.Chart.Name) (hasSuffix "functions.yaml" $path) }}
{{ $.Files.Get "files/hasura/metadata/functions_stg.yaml" | indent 4 }}
{{- else if and (eq "uat" (include "util.environment" $)) (eq "manabie" (include "util.vendor" $)) (eq "bob" $.Chart.Name) (hasSuffix "tables.yaml" $path) }}
{{ $.Files.Get "files/hasura/metadata/tables_stg.yaml" | indent 4 }}
{{- else if and (eq "uat" (include "util.environment" $)) (eq "manabie" (include "util.vendor" $)) (eq "bob" $.Chart.Name) (hasSuffix "functions.yaml" $path) }}
{{ $.Files.Get "files/hasura/metadata/functions_stg.yaml" | indent 4 }}
{{- else }}
{{ $.Files.Get $path | indent 4 }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{- define "util.hasurav2ConfigMap" -}}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "util.fullname" . }}-hasurav2-metadata
  labels:
    {{- include "util.labels" . | nindent 4 }}
data:
  {{- $files := .Files }}
  {{- range $path, $_ := .Files.Glob "files/hasurav2/metadata/**" }}
  {{ $path | replace "/" "-"}}: |
{{ tpl ($files.Get $path) $ | indent 4 }}
  {{- end }}
{{- end }}
