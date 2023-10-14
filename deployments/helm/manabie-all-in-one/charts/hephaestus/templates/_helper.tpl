{{/*
This is copied from "util.serviceAccountName". It represents the name of kafka-connect service account.
*/}}
{{- define "hephaestus.kafkaConnectServiceAccountName" -}}
{{- if eq "dorp" (include "util.environment" .) -}}
{{- printf "prod-kafka-connect" }}
{{- else if and (eq "stag" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "stag-jprep-kafka-connect" }}
{{- else if and (eq "prod" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "prod-jprep-kafka-connect" }}
{{- else }}
{{- printf "%s-kafka-connect" (include "util.environment" .) }}
{{- end }}
{{- end }}


{{/*
This is copied from "util.serviceAccountName"
*/}}
{{- define "hephaestus.dwhKafkaConnectServiceAccountName" -}}
{{- if eq "dorp" (include "util.environment" .) -}}
{{- printf "prod-dwh-kafka-connect" }}
{{- else if and (eq "stag" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "stag-jprep-dwh-kafka-connect" }}
{{- else if and (eq "prod" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "prod-jprep-dwh-kafka-connect" }}
{{- else }}
{{- printf "%s-dwh-kafka-connect" (include "util.environment" .) }}
{{- end }}
{{- end }}
