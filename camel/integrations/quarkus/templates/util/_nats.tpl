{{/*
"is_local" config value for nats jetstream config.
*/}}
{{- define "util.natsIsLocal" -}}
{{- eq "local" (include "util.environment" .) -}}
{{- end -}}
