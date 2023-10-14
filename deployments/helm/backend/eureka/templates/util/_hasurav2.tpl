{{- define "util.hasurav2" -}}
{{ include "util.hasurav2ConfigMap" . }}
---
{{ include "util.hasurav2Pdb" . }}
---
{{ include "util.hasurav2Deployment" . }}
---
{{ include "util.hasurav2Service" . }}

{{- if .Values.adminHttpV2 }}
---
{{ include "virtualservice.adminv2.tpl" . }}
{{- end }}
{{- if .Values.global.vpa.enabled }}
---
{{ include "util.hasurav2Vpa" . }}
{{- end }}
---
{{- if .Capabilities.APIVersions.Has "keda.sh/v1alpha1" }}
{{- if or (.Values.hasura.cronScheduledScaling) (and .Values.hasura.useGlobalCronScheduledScaling .Values.global.hasura.cronScheduledScaling) }}
{{ include "util.keda.hasurav2ScaledObject" . }}
{{- end }}
{{/*
Keda is using HPA under the hood, so if we use Keda, we can't also using HPA.
See https://keda.sh/docs/2.9/faq/.
*/}}
{{- else if .Values.hasurav2.hpa }}
{{ include "util.hasurav2Hpa" . }}
{{- end }}
{{- end -}}

{{- /*
Definitions for the blocks above.
*/}}
{{- define "util.hasurav2Deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "util.fullname" . }}-hasurav2
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  replicas: {{ .Values.global.hasura.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}-hasurav2
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      annotations:
        checksum/hasurav2-metadata: {{ (.Files.Glob "files/hasurav2/metadata/**").AsConfig | sha256sum }}
        checksum/hasura2.secrets.encrypted.yaml: {{ printf "secrets/%s/%s/hasura2.secrets.encrypted.yaml" .Values.global.vendor .Values.global.environment | .Files.Get | sha256sum }}
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
      {{- if ne "prod" .Values.global.environment }}
        sidecar.istio.io/inject: "false"
      {{- end }}
        sidecar.istio.io/proxyCPU: "10m"
        sidecar.istio.io/proxyMemory: "60Mi"
      labels:
        app.kubernetes.io/name: {{ include "util.name" . }}-hasurav2
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
      affinity: {{- include "util.hasurav2Affinity" . | nindent 8 }}
      tolerations: {{- include "util.hasurav2Tolerations" . | nindent 8 }}
      serviceAccountName: {{ include "util.hasurav2ServiceAccountName" . }}
      volumes:
      - name: hasurav2-secrets-decrypted-volume
        emptyDir: {}
      - name: hasurav2-metadata
        configMap:
          name: {{ include "util.name" . }}-hasurav2-metadata
{{- if eq "local" .Values.global.environment }}
      - name: service-credential
        secret:
          secretName: {{ include "util.fullname" . }}
          items:
          - key: service_credential.json
            path: service_credential.json
{{- end }}
      - name: secrets-volume
        secret:
          secretName: {{ include "util.fullname" . }}
          items:
          - key: hasura2.secrets.encrypted.yaml
            path: hasura2.secrets.encrypted.yaml
      initContainers:
{{- if eq "local" .Values.global.environment }}
        - name: wait-for-shamir
          image: "{{ include "util.waitForImage" . }}"
          imagePullPolicy: IfNotPresent
          command:
            - ./scripts/wait-for.sh
            - shamir:5680
            - --timeout=100
{{- end }}
        - name: hasurav2-decrypt-secret
          image: "{{ include "util.sopsImage" . }}"
          imagePullPolicy: IfNotPresent
          command:
            - sops
            - --decrypt
            - --output=/hasura/config.env
            - /configs/hasura2.secrets.encrypted.yaml
          volumeMounts:
          - name: hasurav2-secrets-decrypted-volume
            mountPath: /hasura
          - name: secrets-volume
            mountPath: /configs/hasura2.secrets.encrypted.yaml
            subPath: hasura2.secrets.encrypted.yaml
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
      containers:
        - name: hasura
          image: "{{ .Values.global.hasurav2.image.repository }}:{{ .Values.global.hasurav2.image.tag }}"
          ports:
            - name: http
              containerPort: {{ .Values.hasurav2.service.port }}
              protocol: TCP
          command:
            - /bin/sh
            - -c
            - |

              set -eu

              # sh script, not bash script
              if [ "{{ .Values.hasurav2.unauthorized.enable }}" = "true" ]; then
                export HASURA_GRAPHQL_UNAUTHORIZED_ROLE="{{ .Values.hasurav2.unauthorized.role }}"
              fi

              export HASURA_GRAPHQL_ALLOYDB_DATABASE_URL=$(grep HASURA_GRAPHQL_ALLOYDB_DATABASE_URL hasura/config.env | awk '{print$2}')
              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              export HASURA_GRAPHQL_METADATA_DATABASE_URL=$(grep HASURA_GRAPHQL_METADATA_DATABASE_URL hasura/config.env | awk '{print$2}')
              export HASURA_GRAPHQL_ADMIN_SECRET=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              export HASURA_GRAPHQL_JWT_SECRET=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              exec /bin/docker-entrypoint.sh graphql-engine serve
          env:
          {{- if eq "true" (include "util.hasuraIAMAuthEnabled" .) }}
            - name: HASURA_GRAPHQL_DATABASE_URL
              value: "{{ include "util.hasuraDatabaseConnectionString" . }}"
            - name: HASURA_GRAPHQL_METADATA_DATABASE_URL
              value: "{{ include "util.hasuraMetadataDatabaseConnectionString" . }}"
          {{- end }}
            - name: HASURA_GRAPHQL_METADATA_DIR
              value: "/files/hasurav2/metadata"
            - name: HASURA_GRAPHQL_ENABLE_TELEMETRY
              value: "false"
            - name: HASURA_GRAPHQL_ENABLED_LOG_TYPES
              value: "startup, http-log, webhook-log, websocket-log, query-log" #enable full logs
            - name: HASURA_GRAPHQL_ENABLED_APIS
              value: "{{ .Values.hasurav2.enabledApis }}"
            - name: HASURA_GRAPHQL_ENABLE_CONSOLE
              value: "{{ .Values.hasurav2.enableConsole }}"
            - name: HASURA_GRAPHQL_ENABLE_ALLOWLIST
              value: "{{ .Values.hasurav2.allowList }}"
            - name: HASURA_GRAPHQL_EXPERIMENTAL_FEATURES
              value: "{{ .Values.hasurav2.experimentFeatures }}"
            - name: HASURA_GRAPHQL_DEFAULT_NAMING_CONVENTION
              value: "{{ .Values.hasurav2.namingConvention }}"
          {{- if or (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment) }}
            - name: HASURA_GRAPHQL_LOG_LEVEL
              value: "warn"
          {{- end }}
          resources:
            {{- toYaml .Values.hasurav2.resources | nindent 12 }}
          volumeMounts:
    {{- range $path, $_ := .Files.Glob "files/hasurav2/metadata/**" }}
          - name: hasurav2-metadata
            mountPath: "/{{ $path }}"
            subPath: {{ $path | replace "/" "-" }}
    {{- end }}
          - name: hasurav2-secrets-decrypted-volume
            mountPath: /hasura
          readinessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            failureThreshold: 5
      {{- if .Values.global.sqlProxy.enabled }}
        - name: cloud-sql-proxy
          image: "{{ .Values.global.sqlProxy.image.repository }}:{{ .Values.global.sqlProxy.image.tag }}"
          imagePullPolicy: IfNotPresent
          args:
            - "{{ include "util.databaseInstance" . }}"
            - "--auto-iam-authn"
            - "--structured-logs"
          {{- if not .Values.global.cloudSQLUsePublicIP }}
            - "--private-ip"
          {{- end }}
            - "--max-sigterm-delay=30s"
          securityContext:
            runAsNonRoot: true
          resources:
            {{- toYaml .Values.global.sqlProxy.resources | nindent 12 }}
      {{- end }}
      {{- if .Values.alloydbProxy.enabled }}
        - name: alloydb-auth-proxy
          image: "{{ .Values.alloydbProxy.image.repository }}:{{ .Values.alloydbProxy.image.tag }}"
          imagePullPolicy: IfNotPresent
          command:
            - "/alloydb-auth-proxy"
            - {{ printf "%s" .Values.alloydbProxy.alloydbConnName }}
            - "--structured-logs"
          securityContext:
            runAsNonRoot: true
          resources:
            {{- toYaml .Values.alloydbProxy.resources | nindent 12 }}
      {{- end }}
{{- end }}
