{{- define "util.workloadContainers" }}
- name: {{ .Chart.Name }}
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
          {{ .Chart.Name }} \\
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
    - "{{ .Chart.Name }}"
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
  {{- if .Values.httpPort }}
    - name: http
      protocol: TCP
      containerPort: {{ .Values.httpPort }}
  {{- end }}
  {{- if .Values.grpcPort }}
    - name: grpc
      protocol: TCP
      containerPort: {{ .Values.grpcPort }}
  {{- end }}
{{- with .Values.readinessProbe }}
{{- if .enabled }}
  readinessProbe:
  {{- if .command }}
  {{- toYaml .command | nindent 4 }}
  {{- else }}
  {{- $port := "" }}
  {{- if $.Values.grpcPort }}
  {{- $port = $.Values.grpcPort }}
  {{- else if $.Values.service }}
  {{- $port = $.Values.service.port }}
  {{- end }}
    exec:
      command:
        - sh
        - -c
        - /bin/grpc_health_probe -addr=localhost:{{ $port }} -connect-timeout 750ms -rpc-timeout 750ms
  {{- end }}
    initialDelaySeconds: {{ default 10 .initialDelaySeconds }}
    periodSeconds: {{ default 10 .periodSeconds }}
    timeoutSeconds: {{ default 5 .timeoutSeconds }}
    successThreshold: {{ default 1 .successThreshold }}
    failureThreshold: {{ default 5 .failureThresold }}
{{- end }}
{{- end }}
  resources:
    {{- toYaml .Values.resources | nindent 4 }}
{{- if not .Values.disableScanRLS }}
- name: {{ .Chart.Name }}-scan-rls
  image: {{ include "util.image" . }}
  imagePullPolicy: IfNotPresent
  command:
    - /server
  args:
    - gjob
    - rls_check
    - --commonConfigPath=/configs/{{ .Chart.Name }}.common.config.yaml
    - --configPath=/configs/{{ .Chart.Name }}.config.yaml
    - --secretsPath=/configs/{{ .Chart.Name }}.secrets.encrypted.yaml
  resources:
    {{- toYaml .Values.global.scanRLSResources | nindent 4 }}
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
{{- end }}
{{- end }}
