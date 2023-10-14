{{- define "util.caching" -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "util.cache.name" . }}
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.cache.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "util.cache.redisName" . }}
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.cache.redisName" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "util.cache.redisName" . }}
  labels: {{ include "util.cache.redisLabels" . | nindent 4 }}
data:
{{ (.Files.Glob "configs/redis.conf").AsConfig | indent 2 }}
---
{{ include "util.cache.deployment" . }}
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ include "util.cache.redisName" . }}
  labels: {{ include "util.cache.redisLabels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.cache.redisName" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  replicas: 1
  serviceName: {{ include "util.cache.redisName" . }}
  template:
    metadata:
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
        sidecar.istio.io/inject: "false"
      labels:
        app.kubernetes.io/name: {{ include "util.cache.redisName" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
      volumes:
        - name: redis-conf
          configMap:
            name: {{ include "util.cache.redisName" . }}
            items:
              - key: redis.conf
                path: redis.conf
    {{- with (default .Values.global.imagePullSecrets .Values.imagePullSecrets) }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      containers:
        - name: redis
          image: "{{ .Values.global.caching.redis.image.repository }}:{{ .Values.global.caching.redis.image.tag }}"
          args: ["/configs/redis.conf"]
          volumeMounts:
            - name: {{ include "util.cache.redisName" . }}-pvc
              mountPath: /data
            - name: redis-conf
              mountPath: /configs
              readOnly: true
          ports:
            - name: redis
              protocol: TCP
              containerPort: 6379
          readinessProbe:
            exec:
              command: ["redis-cli", "ping"]
  volumeClaimTemplates:
    - metadata:
        name: {{ include "util.cache.redisName" . }}-pvc
      spec:
        accessModes:
          - ReadWriteOnce
        storageClassName: {{ .Values.global.caching.redis.storageClassName }}
        resources:
          requests:
            storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "util.cache.name" . }}
  labels:
    {{- include "util.cache.labels" . | nindent 4 }}
spec:
  type: ClusterIP
  ports:
  {{- with .Values.httpPort }}
    - name: http-port
      protocol: TCP
      targetPort: http
      port: {{ . }}
  {{- end }}
  selector:
    {{- include "util.cache.selectorLabels" . | nindent 4 }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "util.cache.redisName" . }}
  labels: {{- include "util.cache.redisLabels" . | nindent 4 }}
spec:
  type: ClusterIP
  ports:
    - name: redis
      protocol: TCP
      targetPort: redis
      port: 6379
  selector:
    app.kubernetes.io/name: {{ include "util.cache.redisName" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "util.cache.deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "util.cache.name" . }}
  labels:
{{ include "util.cache.labels" . | indent 4 }}
spec:
  replicas: {{ default 1 .Values.caching.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.cache.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      annotations:
        checksum/{{ .Chart.Name }}.common.config.yaml: {{ tpl ("configs/{{ .Chart.Name }}.common.config.yaml" | .Files.Get) . | sha256sum }}
        checksum/{{ .Chart.Name }}.config.yaml: {{ tpl (printf "configs/%s/%s/{{ .Chart.Name }}.config.yaml" .Values.global.vendor .Values.global.environment | .Files.Get) . | sha256sum }}
        checksum/{{ .Chart.Name }}.secrets.encrypted.yaml: {{ printf "secrets/%s/%s/{{ .Chart.Name }}.secrets.encrypted.yaml" .Values.global.vendor .Values.global.environment | .Files.Get | sha256sum }}
      {{- if eq "local" .Values.global.environment }}
        checksum/service_credential.json: {{ include "util.serviceCredential" . | sha256sum }}
      {{- end }}
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true"
        prometheus.io/scheme: "http"
        prometheus.io/port: "8888"
        prometheus.io/scrape: "true"
        sidecar.istio.io/inject: "false"
      labels:
        app.kubernetes.io/name: {{ include "util.cache.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
    {{- with (default .Values.global.imagePullSecrets .Values.imagePullSecrets) }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      serviceAccountName: {{ include "util.serviceAccountName" . }}
      volumes:
        - name: secrets-volume
          secret:
            secretName: {{ .Chart.Name }}
            items:
            - key: {{ .Chart.Name }}.secrets.encrypted.yaml
              path: {{ .Chart.Name }}.secrets.encrypted.yaml
        {{- if eq "local" .Values.global.environment }}
        - name: service-credential
          secret:
            secretName: {{ .Chart.Name }}
            items:
            - key: service_credential.json
              path: service_credential.json
        {{- end }}
        - name: config-volume
          configMap:
            name: {{ .Chart.Name }}
            items:
            - key: {{ .Chart.Name }}.common.config.yaml
              path: {{ .Chart.Name }}.common.config.yaml
            - key: {{ .Chart.Name }}.config.yaml
              path: {{ .Chart.Name }}.config.yaml
      containers:
        {{- include "util.cache.workloadContainers" . | nindent 8 }}
      {{- with .Values.caching.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.caching }}
      {{- $context := (mustMerge (deepCopy .) $) }}
      affinity: {{- include "util.affinityNew" $context | nindent 8 }}
      tolerations: {{- include "util.tolerations" $context | nindent 8 }}
      {{- end }}
{{- end -}}

{{- define "util.cache.name" -}}
{{- printf "%s-caching" .Chart.Name | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "util.cache.redisName" -}}
{{- printf "%s-caching-redis" (include "util.fullname" .) -}}
{{- end -}}

{{- define "util.cache.labels" -}}
helm.sh/chart: {{ include "util.chart" . }}
app.kubernetes.io/name: {{ include "util.cache.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "util.cache.redisLabels" -}}
helm.sh/chart: {{ include "util.chart" . }}
app.kubernetes.io/name: {{ include "util.cache.redisName" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "util.cache.selectorLabels" -}}
app.kubernetes.io/name: {{ include "util.cache.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "util.cache.workloadContainers" }}
- name: {{ .Chart.Name }}-caching
  image: {{ include "util.image" . }}
  imagePullPolicy: IfNotPresent
{{- if .Values.global.liveReloadEnabled }}
  command:
    - /bin/sh
    - -c
    - |
      #!/bin/bash
      set -eu
      cat <<EOF > modd.conf
      /server {
        daemon +sigterm: /server \\
          gserver \\
          jerry2 \\
          --commonConfigPath=/configs/{{ .Chart.Name }}.common.config.yaml \\
          --configPath=/configs/{{ .Chart.Name }}.config.yaml \\
          --secretsPath=/configs/{{ .Chart.Name }}.secrets.encrypted.yaml
      }
      EOF
      exec modd
{{- else }}
{{- if and .Values.global.debug }}
  command: ["/dlv","--listen=:40000", "--continue", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/server", "--"]
{{- end }}
  args:
    - "gserver"
    - "jerry2"
    - "--commonConfigPath=/configs/{{ .Chart.Name }}.common.config.yaml"
    - "--configPath=/configs/{{ .Chart.Name }}.config.yaml"
    - "--secretsPath=/configs/{{ .Chart.Name }}.secrets.encrypted.yaml"
{{- end }}
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
    mountPath: /configs/{{ .Chart.Name }}.secrets.encrypted.yaml
    subPath: {{ .Chart.Name }}.secrets.encrypted.yaml
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
  ports:
  {{- if .Values.global.debug }}
    - name: delve
      containerPort: 40000
      protocol: TCP
  {{- end }}
  {{- with .Values.httpPort }}
    - name: http
      protocol: TCP
      containerPort: {{ . }}
  {{- end }}
    - name: metrics
      protocol: TCP
      containerPort: 8888
  readinessProbe:
    httpGet:
      path: /healthz
      port: http
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
    successThreshold: 1
  resources:
    {{- toYaml .Values.caching.resources | nindent 4 }}
{{- end }}

{{- define "util.cache.hasuraHost" -}}
{{- printf "%s-hasura" (include "util.fullname" .) -}}
{{- end -}}

{{- define "util.cache.redisAddress" -}}
{{- printf "%s:6379" (include "util.cache.redisName" .) -}}
{{- end -}}
