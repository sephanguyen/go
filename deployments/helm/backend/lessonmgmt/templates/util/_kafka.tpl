{{/*
"is_local" config value for kafka config.
*/}}
{{- define "util.kafkaIsLocal" -}}
{{- eq "local" (include "util.environment" .) -}}
{{- end -}}

{{/*
"object_name_prefix" config value for kafka topic name prefix depend on environment and organization config.
*/}}
{{- define "util.kafkaObjectNamePrefix" -}}
  {{- if .Values.global -}}
    {{- if and .Values.global.environment .Values.global.vendor -}}
      {{ printf "%s.%s." .Values.global.environment .Values.global.vendor }}
    {{- else if .Values.global.environment -}}
      {{ printf "%s." .Values.global.environment }}
    {{- else if .Values.global.vendor -}}
      {{ printf "%s." .Values.global.vendor }}
    {{- end -}}
  {{- end -}}
{{- end -}}
