{{/*
Fall back to .Values.serviceAccountEmailSuffix when .Values.global.serviceAccountEmailSuffix is not set.
*/}}
{{- define "util.serviceAccountEmailSuffix" -}}
{{- if .Values.global -}}
    {{- if .Values.global.serviceAccountEmailSuffix -}}
        {{- .Values.global.serviceAccountEmailSuffix -}}
    {{- else -}}
        {{- .Values.serviceAccountEmailSuffix -}}
    {{- end -}}
{{- else -}}
    {{- .Values.serviceAccountEmailSuffix -}}
{{- end -}}
{{- end -}}

{{/*
Service account attributes.
We are using GKE Workload Identity, so there are a few things to set up in here:
  - creating k8s service account
  - annotating that service account
See: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
*/}}
{{- define "util.serviceAccount" -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "util.serviceAccountName" . }}
  labels:
    {{- include "util.labels" . | nindent 4 }}
  annotations:
    {{- include "util.serviceAccountAnnotations" . | nindent 4 }}
{{- end }}

{{/*
Name of the k8s service account object.
*/}}
{{- define "util.serviceAccountName" -}}
{{- if and (eq "stag" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "stag-jprep-%s" (include "util.fullname" .) }}
{{- else if and (eq "dorp" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "prod-jprep-%s" (include "util.fullname" .) }}
{{- else if and (eq "prod" (include "util.environment" .)) (eq "jprep" (include "util.vendor" .)) }}
{{- printf "prod-jprep-%s" (include "util.fullname" .) }}
{{- else if eq "dorp" (include "util.environment" .) }}
{{- printf "prod-%s" (include "util.fullname" .) }}
{{- else }}
{{- printf "%s-%s" (include "util.environment" .) (include "util.fullname" .) }}
{{- end }}
{{- end }}

{{- define "util.serviceAccountEmail" -}}
{{ printf "%s@%s.iam.gserviceaccount.com" (include "util.serviceAccountName" .) (include "util.serviceAccountEmailSuffix" .) }}
{{- end }}


{{- define "util.serviceAccountAnnotations" -}}
{{- if or .Values.preHookUpsertStream .Values.preHookUpsertKafkaTopic }}
"helm.sh/hook": pre-install,pre-upgrade
"helm.sh/hook-weight": "-50"
"helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
{{- end }}
iam.gke.io/gcp-service-account: {{ include "util.serviceAccountEmail" . }}
{{- end }}

{{/*
Service account attributes for hasura v2.
Similar to "util.serviceAccount" but using a different IAM account customized for hasura v2.
*/}}
{{- define "util.hasurav2ServiceAccount" -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "util.hasurav2ServiceAccountName" . }}
  labels:
    {{- include "util.labels" . | nindent 4 }}
  annotations:
    iam.gke.io/gcp-service-account: {{ include "util.hasurav2ServiceAccountEmail" . }}
{{- end }}

{{- define "util.hasurav2ServiceAccountName" -}}
{{- if or (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment)  -}}
{{ printf "%s%s-h" .Values.global.dbUserPrefix (include "util.name" .) }}
{{- else -}}
{{ printf "%s%s-hasura" .Values.global.dbUserPrefix (include "util.name" .) }}
{{- end -}}
{{- end }}

{{- define "util.hasurav2ServiceAccountEmail" -}}
{{ printf "%s@%s.iam.gserviceaccount.com" (include "util.hasurav2ServiceAccountName" .) (include "util.serviceAccountEmailSuffix" .) }}
{{- end }}

{{/*
Service account attributes for hasura v1.
Old envs/orgs/services are still using the old service accounts,
but we are migrating them to use the same service account as hasura v2.
*/}}
{{- define "util.hasuraServiceAccountName" -}}
{{- if or (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment)  -}}
{{ include "util.hasurav2ServiceAccountName" . }}
{{- else -}}
{{ include "util.serviceAccountName" . }}
{{- end -}}
{{- end }}

