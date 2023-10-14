{{- define "util.jobs" -}}
{{- range $jobName, $jobValues := .Values.jobs }}
{{- if $jobValues.enabled }}
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "util.fullname" $ }}-{{ $jobName }}
  labels:
{{ include "util.labels" $ | indent 4 }}
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ printf "%s-%s" (include "util.name" $) $jobName }}
        app.kubernetes.io/instance: {{ $.Release.Name }}
        sidecar.istio.io/inject: "false"
    spec:
      restartPolicy: {{ $jobValues.restartPolicy | default "OnFailure"}}
      serviceAccountName: {{ include "util.serviceAccountName" $ }}
      volumes:
{{- if eq "local" $.Values.global.environment }}
      - name: service-credential
        secret:
          secretName: {{ include "util.fullname" $ }}
          items:
          - key: service_credential.json
            path: service_credential.json
{{- end }}
      - name: config-volume
        configMap:
          name: {{ include "util.fullname" $ }}
          items:
          - key: {{ $.Chart.Name }}.common.config.yaml
            path: {{ $.Chart.Name }}.common.config.yaml
          - key: {{ $.Chart.Name }}.config.yaml
            path: {{ $.Chart.Name }}.config.yaml
      - name: secrets-volume
        secret:
          secretName: {{ include "util.fullname" $ }}
          items:
          - key: {{ $.Chart.Name }}.secrets.encrypted.yaml
            path: {{ $.Chart.Name }}.secrets.encrypted.yaml
      containers:
        - name: {{ $jobName }}
          image: {{ include "util.image" $ }}
          imagePullPolicy: IfNotPresent
          args:
            - gjob
            - {{ $jobValues.cmd }}
            - --commonConfigPath=/configs/{{ $.Chart.Name }}.common.config.yaml
            - --configPath=/configs/{{ $.Chart.Name }}.config.yaml
            - --secretsPath=/configs/{{ $.Chart.Name }}.secrets.encrypted.yaml
          {{- range $key, $arg := $jobValues.args }}
            - --{{ $key }}={{ $arg }}
          {{- end }}
          volumeMounts:
          - name: config-volume
            mountPath: /configs/{{ $.Chart.Name }}.common.config.yaml
            subPath: {{ $.Chart.Name }}.common.config.yaml
            readOnly: true
          - name: config-volume
            mountPath: /configs/{{ $.Chart.Name }}.config.yaml
            subPath: {{ $.Chart.Name }}.config.yaml
            readOnly: true
          - name: secrets-volume
            mountPath: /configs/{{ $.Chart.Name }}.secrets.encrypted.yaml
            subPath: {{ $.Chart.Name }}.secrets.encrypted.yaml
            readOnly: true
          {{- if eq "local" $.Values.global.environment }}
          - name: service-credential
            mountPath: /configs/service_credential.json
            subPath: service_credential.json
            readOnly: true
          env:
          - name: GOOGLE_APPLICATION_CREDENTIALS
            value: "/configs/service_credential.json"
          {{- end }}
          resources:
            {{- toYaml $.Values.resources | nindent 12 }}
      {{- with $.Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
    {{- with $.Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
    {{- end }}
    {{- with $.Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
    {{- end }}
{{- end }}
{{- end }}
{{- end }}
