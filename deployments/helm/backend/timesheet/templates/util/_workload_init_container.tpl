{{- define "util.workloadInitContainers" }}
{{- $depSvcList := (list (dict "name" "shamir" "port" "5680")) }}
{{- if hasKey .Values "waitForServices" }}
{{- $depSvcList = .Values.waitForServices }}
{{- end }}
{{- if not (empty $depSvcList) }}
- name: wait-for-shamir
  image: "{{ include "util.waitForImage" . }}"
  imagePullPolicy: IfNotPresent
  command:
  - /bin/sh
  - -c
  - |
    set -e
  {{- range $depSvcList }}
    ./scripts/wait-for.sh {{ .name }}.{{ $.Release.Namespace }}.svc.cluster.local:{{ .port }} -t 100
  {{- end }}
{{- end }}
{{- $isPartnerMigrationDisabled := or (eq "synersia" .Values.global.vendor) -}}
{{- if and .Values.migrationEnabled (not $isPartnerMigrationDisabled) }}
- name: {{ .Chart.Name }}-migrate
  image: {{ include "util.image" . }}
  imagePullPolicy: IfNotPresent
  command: ["/server"]
  args:
    - gjob
    - sql_migrate
    - --commonConfigPath=/configs/{{ .Chart.Name }}.common.config.yaml
    - --configPath=/configs/{{ .Chart.Name }}.config.yaml
    - --secretsPath=/configs/{{ .Chart.Name }}_migrate.secrets.encrypted.yaml
  volumeMounts:
  - name: config-volume
    mountPath: /configs/{{ .Chart.Name }}.common.config.yaml
    subPath: {{ .Chart.Name }}.common.config.yaml
    readOnly: true
  - name: config-volume
    mountPath: /configs/{{ .Chart.Name }}.config.yaml
    subPath: {{ .Chart.Name }}.config.yaml
    readOnly: true
  - name: secrets-volume
    mountPath: /configs/{{ .Chart.Name }}_migrate.secrets.encrypted.yaml
    subPath: {{ .Chart.Name }}_migrate.secrets.encrypted.yaml
    readOnly: true
{{- if eq "local" .Values.global.environment }}
  - name: service-credential
    mountPath: /configs/service_credential.json
    subPath: service_credential.json
    readOnly: true
  env:
  - name: GOOGLE_APPLICATION_CREDENTIALS
    value: "/configs/service_credential.json"
{{- end }}
{{- if and .Values.hasuraEnabled (ne "draft" .Chart.Name) }}
- name: hasura-decrypt-secret
  image: "{{ include "util.sopsImage" . }}"
  imagePullPolicy: IfNotPresent
  command:
    - /bin/sh
    - -c
    - |

      set -e
      startTime=$(date +"%T")
      echo "Hasura decrypt secret start time: $startTime"

      sops --decrypt /configs/hasura.secrets.encrypted.yaml > /hasura/config.env

      endTime=$(date +"%T")
      echo "Hasura decrypt secret end time: $endTime"

      duration=$(date -d @$(( $(date -d "$endTime" +%s) - $(date -d "$startTime" +%s) )) -u +'%H:%M:%S')
      echo "Duration: $duration"
  volumeMounts:
  - name: hasura-secrets-decrypted-volume
    mountPath: /hasura
  - name: secrets-volume
    mountPath: /configs/hasura.secrets.encrypted.yaml
    subPath: hasura.secrets.encrypted.yaml
    readOnly: true
{{- if eq "local" .Values.global.environment }}
  - name: service-credential
    mountPath: /configs/service_credential.json
    subPath: service_credential.json
    readOnly: true
  env:
  - name: GOOGLE_APPLICATION_CREDENTIALS
    value: "/configs/service_credential.json"
{{- end }}
- name: hasura-migration
  image: "{{ .Values.global.hasura.migrationImage.repository }}:{{ .Values.global.hasura.migrationImage.tag }}"
  imagePullPolicy: IfNotPresent
  command:
    - /bin/sh
    - -c
    - |

      set -e
    {{- if .Values.global.sqlProxy.enabled }}
      echo "Running cloud-sql-proxy asynchronously"
      cloud_sql_proxy2 {{ include "util.databaseInstance" . }} \
        --port 5432 \
        --auto-iam-authn \
        --structured-logs \
      {{- if not .Values.global.cloudSQLUsePublicIP }}
        --private-ip \
      {{- end }}
      {{- if or (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment) }}
        --impersonate-service-account {{ include "util.hasurav2ServiceAccountEmail" . }} \
      {{- end }}
        --max-sigterm-delay=30s &

      until nc -z 127.0.0.1 5432; do
        echo "Waiting for the proxy to run..."
        sleep 2
      done
    {{- end }}

      startTime=$(date +"%T")
      echo "Hasura migration start time: $startTime"
      export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
      /bin/docker-entrypoint.sh
      exitcode=$?
    {{- if .Values.global.sqlProxy.enabled }}
      echo "Sending SIGTERM to cloud-sql-proxy process"
      fuser -k -TERM 5432/tcp
    {{- end }}
      endTime=$(date +"%T")
      echo "Hasura migration end time: $endTime"

      duration=$(date -d @$(( $(date -d "$endTime" +%s) - $(date -d "$startTime" +%s) )) -u +'%H:%M:%S')
      echo "Duration: $duration"
      exit $exitcode
  env:
    - name: HASURA_GRAPHQL_ENABLE_TELEMETRY
      value: "false"
    - name: HASURA_GRAPHQL_MIGRATIONS_SERVER_TIMEOUT
      value: "180"
  resources:
    {{- toYaml .Values.hasura.resources | nindent 4 }}
  volumeMounts:
    - name: hasura-metadata
      mountPath: /hasura-metadata
    - name: hasura-secrets-decrypted-volume
      mountPath: /hasura
{{- end }}
{{- end }}
{{- end }}
