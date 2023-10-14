{{- define "util.hasura" -}}
{{ include "util.hasuraConfigMap" . }}
---
{{ include "util.hasuraDeployment" . }}
---
{{ include "util.hasuraPdb" . }}
---
{{ include "util.hasuraService" . }}

{{- if .Values.adminHttp }}
---
{{ include "virtualservice.admin.tpl" . }}
{{- end }}
{{- if .Values.global.vpa.enabled }}
---
{{ include "util.hasuraVpa" . }}
{{- end }}
---
{{- if .Capabilities.APIVersions.Has "keda.sh/v1alpha1" }}
{{- if .Values.hasura.onDemandNodeDeployment }}
{{- if .Values.hasura.onDemandNodeDeployment.enabled }}
{{ include "util.keda.hasuraScaledObject" . }}
{{- end }}
{{- end }}
{{/*
Keda is using HPA under the hood, so if we use Keda, we can't also using HPA.
See https://keda.sh/docs/2.9/faq/.
*/}}
{{- else if .Values.hasura.hpa }}
{{ include "util.hasuraHpa" . }}
{{- end }}
{{- end -}}

{{- /*
Definitions for the blocks above.
*/}}
{{- define "util.hasuraDeployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "util.fullname" . }}-hasura
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  replicas: {{ .Values.global.hasura.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}-hasura
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      annotations:
        checksum/hasura-metadata: {{ (.Files.Glob "files/hasura/metadata/*").AsConfig | sha256sum }}
        checksum/hasura.secrets.encrypted.yaml: {{ printf "secrets/%s/%s/hasura.secrets.encrypted.yaml" .Values.global.vendor .Values.global.environment | .Files.Get | sha256sum }}
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
      {{- if ne "prod" .Values.global.environment }}
        sidecar.istio.io/inject: "false"
      {{- end }}
        sidecar.istio.io/proxyCPU: "10m"
        sidecar.istio.io/proxyMemory: "60Mi"
      {{- if .Values.hasura.hasuraMetricAdapter }}
        prometheus.io/scheme: "http"
        prometheus.io/port: "9999"
        prometheus.io/scrape: "true"
      {{- end }}
      labels:
        app.kubernetes.io/name: {{ include "util.name" . }}-hasura
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
      affinity: {{- include "util.hasuraAffinity" . | nindent 8 }}
      tolerations: {{- include "util.hasuraTolerations" . | nindent 8 }}
      serviceAccountName: {{ include "util.hasuraServiceAccountName" . }}
    {{- with (default .Values.global.hasura.imagePullSecrets .Values.hasura.imagePullSecrets) }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      volumes:
      - name: hasura-secrets-decrypted-volume
        emptyDir: {}
      - name: hasura-metadata
        configMap:
          name: {{ include "util.name" . }}-hasura-metadata
      {{- if .Values.hasura.hasuraMetricAdapter }}
      - name: logs
        emptyDir: {}
      {{- end }}
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
          - key: hasura.secrets.encrypted.yaml
            path: hasura.secrets.encrypted.yaml
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
        - name: hasura-decrypt-secret
          image: "{{ include "util.sopsImage" . }}"
          imagePullPolicy: IfNotPresent
          command:
            - sops
            - --output=/hasura/config.env
            - --decrypt
            - /configs/hasura.secrets.encrypted.yaml
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
      containers:
        - name: hasura
          image: "{{ .Values.global.hasura.image.repository }}:{{ .Values.global.hasura.image.tag }}"
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: {{ .Values.hasura.service.port }}
              protocol: TCP
{{- if .Values.hasura.hasuraMetricAdapter }}
          command:
            - /bin/sh
            - -c
            - |

              set -euo pipefail
              adminSecret=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              jwtSecret=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              rm -rf /tmp/log/stdout.log
              mkdir -p /tmp/log/
              mkfifo /tmp/log/stdout.log
              exec graphql-engine serve \
                --admin-secret $adminSecret \
                --connections {{ .Values.hasura.pgConnections }} \
                --enable-allowlist \
                --enable-console \
                --enabled-apis {{ .Values.hasura.enabledApis }} \
                --jwt-secret $jwtSecret \
                {{- if .Values.hasura.anonymous.enabled }}
                --unauthorized-role anonymous \
                {{- end }}
                --timeout {{ .Values.hasura.pgTimeout }} | tee /tmp/log/stdout.log
{{- else }}
          command:
            - /bin/sh
            - -c
            - |

              set -euo pipefail
              adminSecret=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              jwtSecret=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              exec graphql-engine serve \
                --admin-secret $adminSecret \
                --connections {{ .Values.hasura.pgConnections }} \
                --enable-allowlist \
                --enable-console \
                --enabled-apis {{ .Values.hasura.enabledApis }} \
                --jwt-secret $jwtSecret \
                {{- if .Values.hasura.anonymous.enabled }}
                --unauthorized-role anonymous \
                {{- end }}
                --timeout {{ .Values.hasura.pgTimeout }}
{{- end }}
          env:
          {{- if eq "true" (include "util.hasuraIAMAuthEnabled" .) }}
            - name: HASURA_GRAPHQL_DATABASE_URL
              value: "{{ include "util.hasuraDatabaseConnectionString" . }}"
          {{- end }}
            - name: HASURA_GRAPHQL_ENABLE_TELEMETRY
              value: "false"
            - name: HASURA_GRAPHQL_ENABLED_LOG_TYPES
              value: "startup, http-log, webhook-log, websocket-log, query-log" #enable full logs
          {{- if or (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment) }}
            {{- if .Values.hasura.hasuraMetricAdapter }}
            - name: HASURA_GRAPHQL_LOG_LEVEL
              value: "info"
            {{- else }}
            - name: HASURA_GRAPHQL_LOG_LEVEL
              value: "warn"
            {{- end }}
          {{- end }}
          resources:
            {{- toYaml .Values.hasura.resources | nindent 12 }}
          volumeMounts:
          - name: hasura-metadata
            mountPath: /hasura-metadata
          - name: hasura-secrets-decrypted-volume
            mountPath: /hasura
          {{- if .Values.hasura.hasuraMetricAdapter }}
          - name: logs
            mountPath: /tmp/log
          {{- end }}
          readinessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            failureThreshold: 5
      {{- if .Values.hasura.hasuraMetricAdapter }}
      {{- if .Values.hasura.hasuraMetricAdapter.enabled }}
        - name: hasura-metric-adapter
          image: "{{ .Values.hasura.hasuraMetricAdapter.image.repository }}:{{ .Values.hasura.hasuraMetricAdapter.image.tag }}"
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
            - |
              adminSecret=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              jwtSecret=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              exec /metrics
          env:
          - name: LOG_FILE
            value: /tmp/log/stdout.log
          - name: LISTEN_ADDR
            value: 0.0.0.0:9999
          ports:
          - containerPort: 9999
            protocol: TCP
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 256Mi
          volumeMounts:
          - name: hasura-metadata
            mountPath: /hasura-metadata
          - name: hasura-secrets-decrypted-volume
            mountPath: /hasura
          - name: logs
            mountPath: /tmp/log
      {{- end }}
      {{- end }}
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
{{- if .Values.hasura.onDemandNodeDeployment }}
{{- if .Values.hasura.onDemandNodeDeployment.enabled }}
---
{{/*
The template for the Hasura on-demand node deployment.
*/}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "util.fullname" . }}-hasura-on-demand-node
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  replicas: {{ default .Values.global.hasura.onDemandNodeDeployment.replicaCount .Values.hasura.onDemandNodeDeployment.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}-hasura
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      annotations:
        checksum/hasura-metadata: {{ (.Files.Glob "files/hasura/metadata/*").AsConfig | sha256sum }}
        checksum/hasura.secrets.encrypted.yaml: {{ printf "secrets/%s/%s/hasura.secrets.encrypted.yaml" .Values.global.vendor .Values.global.environment | .Files.Get | sha256sum }}
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
      {{- if ne "prod" .Values.global.environment }}
        sidecar.istio.io/inject: "false"
      {{- end }}
        sidecar.istio.io/proxyCPU: "10m"
        sidecar.istio.io/proxyMemory: "60Mi"
      {{- if .Values.hasura.hasuraMetricAdapter }}
        prometheus.io/scheme: "http"
        prometheus.io/port: "9999"
        prometheus.io/scrape: "true"
      {{- end }}
      labels:
        app.kubernetes.io/name: {{ include "util.name" . }}-hasura
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
      serviceAccountName: {{ include "util.hasuraServiceAccountName" . }}
    {{- with (default .Values.global.hasura.imagePullSecrets .Values.hasura.imagePullSecrets) }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      volumes:
      - name: hasura-secrets-decrypted-volume
        emptyDir: {}
      - name: hasura-metadata
        configMap:
          name: {{ include "util.name" . }}-hasura-metadata
      {{- if .Values.hasura.hasuraMetricAdapter }}
      - name: logs
        emptyDir: {}
      {{- end }}
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
          - key: hasura.secrets.encrypted.yaml
            path: hasura.secrets.encrypted.yaml
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
        - name: hasura-decrypt-secret
          image: "{{ include "util.sopsImage" . }}"
          imagePullPolicy: IfNotPresent
          command:
            - sops
            - --output=/hasura/config.env
            - --decrypt
            - /configs/hasura.secrets.encrypted.yaml
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
      containers:
        - name: hasura
          image: "{{ .Values.global.hasura.image.repository }}:{{ .Values.global.hasura.image.tag }}"
          imagePullPolicy: IfNotPresent
          ports:
            - name: http
              containerPort: {{ .Values.hasura.service.port }}
              protocol: TCP
{{- if .Values.hasura.hasuraMetricAdapter }}
          command:
            - /bin/sh
            - -c
            - |

              set -euo pipefail
              adminSecret=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              jwtSecret=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              rm -rf /tmp/log/stdout.log
              mkdir -p /tmp/log/
              mkfifo /tmp/log/stdout.log
              exec graphql-engine serve \
                --admin-secret $adminSecret \
                --connections {{ .Values.hasura.pgConnections }} \
                --enable-allowlist \
                --enable-console \
                --enabled-apis {{ .Values.hasura.enabledApis }} \
                --jwt-secret $jwtSecret \
                {{- if .Values.hasura.anonymous.enabled }}
                --unauthorized-role anonymous \
                {{- end }}
                --timeout {{ .Values.hasura.pgTimeout }} | tee /tmp/log/stdout.log
{{- else }}
          command:
            - /bin/sh
            - -c
            - |

              set -euo pipefail
              adminSecret=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              jwtSecret=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              exec graphql-engine serve \
                --admin-secret $adminSecret \
                --connections {{ .Values.hasura.pgConnections }} \
                --enable-allowlist \
                --enable-console \
                --enabled-apis {{ .Values.hasura.enabledApis }} \
                --jwt-secret $jwtSecret \
                {{- if .Values.hasura.anonymous.enabled }}
                --unauthorized-role anonymous \
                {{- end }}
                --timeout {{ .Values.hasura.pgTimeout }}
{{- end }}
          env:
          {{- if eq "true" (include "util.hasuraIAMAuthEnabled" .) }}
            - name: HASURA_GRAPHQL_DATABASE_URL
              value: "{{ include "util.hasuraDatabaseConnectionString" . }}"
          {{- end }}
            - name: HASURA_GRAPHQL_ENABLE_TELEMETRY
              value: "false"
            - name: HASURA_GRAPHQL_ENABLED_LOG_TYPES
              value: "startup, http-log, webhook-log, websocket-log, query-log" #enable full logs
          {{- if or (eq "stag" .Values.global.environment) (eq "uat" .Values.global.environment) }}
            {{- if .Values.hasura.hasuraMetricAdapter }}
            - name: HASURA_GRAPHQL_LOG_LEVEL
              value: "info"
            {{- else }}
            - name: HASURA_GRAPHQL_LOG_LEVEL
              value: "warn"
            {{- end }}
          {{- end }}
          resources:
            {{- toYaml .Values.hasura.resources | nindent 12 }}
          volumeMounts:
          - name: hasura-metadata
            mountPath: /hasura-metadata
          - name: hasura-secrets-decrypted-volume
            mountPath: /hasura
          {{- if .Values.hasura.hasuraMetricAdapter }}
          - name: logs
            mountPath: /tmp/log
          {{- end }}
          readinessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            failureThreshold: 5
      {{- if .Values.hasura.hasuraMetricAdapter }}
      {{- if .Values.hasura.hasuraMetricAdapter.enabled }}
        - name: hasura-metric-adapter
          image: "{{ .Values.hasura.hasuraMetricAdapter.image.repository }}:{{ .Values.hasura.hasuraMetricAdapter.image.tag }}"
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
            - |
              adminSecret=$(grep HASURA_GRAPHQL_ADMIN_SECRET hasura/config.env | awk '{print$2}')
              jwtSecret=$(grep HASURA_GRAPHQL_JWT_SECRET hasura/config.env | tr -d "'" | awk '{print$2}')

              # If not set, means we have to get this value from secret file
              if [ -z ${HASURA_GRAPHQL_DATABASE_URL+x} ]; then
                export HASURA_GRAPHQL_DATABASE_URL=$(grep HASURA_GRAPHQL_DATABASE_URL hasura/config.env | awk '{print$2}')
              fi
              exec /metrics
          env:
          - name: LOG_FILE
            value: /tmp/log/stdout.log
          - name: LISTEN_ADDR
            value: 0.0.0.0:9999
          ports:
          - containerPort: 9999
            protocol: TCP
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 256Mi
          volumeMounts:
          - name: hasura-metadata
            mountPath: /hasura-metadata
          - name: hasura-secrets-decrypted-volume
            mountPath: /hasura
          - name: logs
            mountPath: /tmp/log
      {{- end }}
      {{- end }}
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
      {{- $overwrites := dict "Values" (dict "hasura" (dict
          "affinityOverride" (dict
            "nodeAffinity" (dict
              "requiredDuringSchedulingIgnoredDuringExecution" (dict
                "nodeSelectorTerms" (list (dict
                  "matchExpressions" (list (dict
                    "key" "backend-on-demand-node"
                    "operator" "In"
                    "values" (list "true")
                  ))
                ))
              )
            )
          )
          "tolerations" (list (dict
            "key" "backend-on-demand-node"
            "operator" "Exists"
            "effect" "NoSchedule"
          ))
        ))
      }}
      {{- $context := mustMergeOverwrite (deepCopy .) $overwrites }}
      affinity: {{- include "util.hasuraAffinity" $context | nindent 8 }}
      tolerations: {{- include "util.hasuraTolerations" $context | nindent 8 }}
{{- end -}}
{{- end -}}
{{- end }}
