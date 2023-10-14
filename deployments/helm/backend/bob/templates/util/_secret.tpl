{{/*
util.secret is the template for the k8s secret object for most manabie-all-in-one
business services.

For more information, see https://manabie.atlassian.net/wiki/spaces/TECH/pages/503940027/Config+Secret+management
*/}}
{{- define "util.secret" -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "util.fullname" . }}
{{- if or .Values.preHookUpsertStream .Values.preHookUpsertKafkaTopic }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade,post-upgrade
    "helm.sh/hook-weight": "-49"
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
{{- end }}
type: Opaque
data:
{{- if eq "local" .Values.global.environment }}
  service_credential.json: {{ include "util.serviceCredential" . }}
{{- end }}
{{ (.Files.Glob (printf "secrets/%s/%s/*.encrypted.yaml" .Values.global.vendor .Values.global.environment)).AsSecrets | indent 2 }}
{{- end -}}
