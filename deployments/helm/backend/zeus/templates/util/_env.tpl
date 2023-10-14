{{/*
Returns the current environment (local, stag, uat, dorp, prod) currently deploying.

Fallbacks to .Values.environment when .Values.global.environment is not set.
*/}}
{{- define "util.environment" -}}
{{- if .Values.global -}}
    {{- if .Values.global.environment -}}
        {{- .Values.global.environment -}}
    {{- else -}}
        {{- .Values.environment -}}
    {{- end -}}
{{- else -}}
    {{- .Values.environment -}}
{{- end -}}
{{- end -}}

{{/*
Returns `prod` for preproduction, otherwise returns value of "util.environment".
This is so that preproduction is treated identically to production.

Therefore, the output can be one of: local, stag, uat, prod.
*/}}
{{- define "util.runtimeEnvironment" -}}
{{- $e := include "util.environment" . -}}
{{- if eq "dorp" $e -}}
  {{- $e = "prod" -}}
{{- end -}}
{{- $e -}}
{{- end -}}

{{/*
Returns the current vendor (a.k.a organization) current deploying.

Fallbacks to .Values.vendor when .Values.global.vendor is not set.
*/}}
{{- define "util.vendor" -}}
{{- if .Values.global -}}
    {{- if .Values.global.vendor -}}
        {{- .Values.global.vendor -}}
    {{- else -}}
        {{- .Values.vendor -}}
    {{- end -}}
{{- else -}}
    {{- .Values.vendor -}}
{{- end -}}
{{- end -}}
