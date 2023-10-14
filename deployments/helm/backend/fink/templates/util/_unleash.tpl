{{/*
Unleash URL for backend applications.

There are 2 versions of unleash: JPREP and non-JPREP.

Notes:
  - For preproduction/production, use public domain so that requests are properly forwarded to tokyo's unleash
    (except for tokyo and jprep themselves).
  - Production synersia does not have `prod` in the domain.
  - Otherwise, use k8s' domain to connect to unleash, to reduce network latency.
*/}}
{{- define "util.unleashURL" -}}
{{- if and (eq "dorp" (include "util.environment" .)) (and (ne "tokyo" (include "util.vendor" .)) (ne "jprep" (include "util.vendor" .))) -}}
{{- printf "https://admin.prep.%s.manabie.io/unleash/api" (include "util.vendor" .) -}}
{{- else if and (eq "prod" (include "util.environment" .)) (eq "synersia" (include "util.vendor" .)) -}}
{{- printf "https://admin.synersia.manabie.io/unleash/api" -}}
{{- else if and (eq "prod" (include "util.environment" .)) (and (ne "tokyo" (include "util.vendor" .)) (ne "jprep" (include "util.vendor" .))) -}}
{{- printf "https://admin.prod.%s.manabie.io/unleash/api" (include "util.vendor" .) -}}
{{- else if and (eq "local" (include "util.environment" .)) (eq "e2e" (include "util.vendor" .)) -}}
{{- printf "https://admin.staging-green.manabie.io/unleash/api" -}}
{{- else -}}
http://unleash.{{ .Values.global.environment }}-{{ .Values.global.vendor }}-unleash.svc.cluster.local:4242/unleash/api
{{- end -}}
{{- end -}}
