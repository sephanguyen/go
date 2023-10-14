{{/*
Database user which runs the SQL migration for the current service.
*/}}
{{- define "util.databaseMigrationUser" -}}
{{- if or (eq "stag" (include "util.environment" .)) (eq "uat" (include "util.environment" .)) }}
{{- printf "%s-m@%s.iam" (include "util.serviceAccountName" .) (include "util.serviceAccountEmailSuffix" .) }}
{{- else }}
{{- printf "postgres" }}
{{- end }}
{{- end }}

{{/*
Email of the service account used in impersonation when connecting to Cloud SQL.
Mainly used for database migration.
*/}}
{{- define "util.databaseMigrationServiceAccountEmail" -}}
{{- if or (eq "stag" (include "util.environment" .)) (eq "uat" (include "util.environment" .)) }}
{{- printf "%s.gserviceaccount.com" (include "util.databaseMigrationUser" .) }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{/*
Database username for a service. Generally, one service has one database username to use.
*/}}
{{- define "util.databaseUser" -}}
{{- if eq "local" (include "util.environment" .) }}
{{- printf "%s%s" .Values.global.dbUserPrefix (include "util.fullname" .) }}
{{- else }}
{{- printf "%s%s@%s.iam" .Values.global.dbUserPrefix (include "util.fullname" .) .Values.global.serviceAccountEmailSuffix }}
{{- end }}
{{- end }}

{{/*
Database hostname for a Go service. For non-Go services, do not use this.
Generally:
  - In local, it should points to the postgres service running in emulator namespace
  - Otherwise, it should be null (as we use cloudsqlconn library to connect)
*/}}
{{- define "util.databaseHost" -}}
{{- if eq "local" (include "util.environment" .) }}
{{- printf "postgres-infras.emulator.svc.cluster.local" }}
{{- end }}
{{- end }}

{{/*
Database instance.
  - `eureka` uses LMS instance.
  - `auth` uses Auth instance.
  - The rest use Common instance.
*/}}
{{- define "util.databaseInstance" -}}
{{- if eq .Chart.Name "eureka" }}
{{- .Values.global.cloudSQLLMSInstance }}
{{- else if eq .Chart.Name "auth" }}
{{- .Values.global.cloudSQLAuthInstance }}
{{- else }}
{{- .Values.global.cloudSQLCommonInstance }}
{{- end }}
{{- end }}
