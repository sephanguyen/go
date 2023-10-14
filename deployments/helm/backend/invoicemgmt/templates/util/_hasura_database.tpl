{{/*
Return "true" if Hasura IAM authentication to Cloud SQL is enabled.
*/}}
{{- define "util.hasuraIAMAuthEnabled" -}}
{{- $env := (include "util.environment" .) -}}
{{- or (eq "stag" $env) (eq "uat" $env) -}}
{{- end -}}

{{/*
Hasura database user.
Currently it can only be used in stag & uat.
*/}}
{{- define "util.hasuraDatabaseUser" -}}
{{- $env := (include "util.environment" .) }}
{{- if eq "false" (include "util.hasuraIAMAuthEnabled" .) }}
{{- printf "IAM Authentication is only enabled in stag/uat (current env: %s)" $env | fail }}
{{- else }}
{{- printf "%s%s-h@%s.iam" .Values.global.dbUserPrefix (include "util.fullname" .) .Values.global.serviceAccountEmailSuffix }}
{{- end }}
{{- end -}}

{{/*
Fully-qualified database connection string for Hasura.
*/}}
{{- define "util.hasuraDatabaseConnectionString" -}}
{{- $env := (include "util.environment" .) }}
{{- if eq "false" (include "util.hasuraIAMAuthEnabled" .) }}
{{- printf "IAM Authentication is only enabled in stag/uat (current env: %s)" $env | fail }}
{{- else }}
{{- $urlEncodedUser := (include "util.hasuraDatabaseUser" . | urlquery) }}
{{- printf "postgres://%s@127.0.0.1:5432/%s%s?sslmode=disable&application_name=%s" $urlEncodedUser .Values.global.dbPrefix (include "util.fullname" .) $urlEncodedUser }}
{{- end }}
{{- end }}

{{/*
Fully-qualified database connection string for Hasura's metadata database.
*/}}
{{- define "util.hasuraMetadataDatabaseConnectionString" -}}
{{- $env := (include "util.environment" .) }}
{{- if eq "false" (include "util.hasuraIAMAuthEnabled" .) }}
{{- printf "IAM Authentication is only enabled in stag/uat (current env: %s)" $env | fail }}
{{- else }}
{{- $urlEncodedUser := (include "util.hasuraDatabaseUser" . | urlquery) }}
{{- printf "postgres://%s@127.0.0.1:5432/%s%s_hasura_metadata?sslmode=disable&application_name=%s" $urlEncodedUser .Values.global.dbPrefix (include "util.fullname" .) $urlEncodedUser }}
{{- end }}
{{- end }}
