{{- define "util.cronjobs" -}}
{{- range $cronjobName, $cronjobValues := .Values.cronjobs }}
{{- if not $cronjobValues.disabled }}
---
{{- if $.Capabilities.APIVersions.Has "batch/v1/CronJob" }}
apiVersion: batch/v1
{{- else }}
apiVersion: batch/v1beta1
{{- end }}
kind: CronJob
metadata:
  name: {{ include "util.fullname" $ }}-{{ $cronjobName }}
  labels:
{{ include "util.labels" $ | indent 4 }}
spec:
  concurrencyPolicy: Forbid
  schedule: "{{ $cronjobValues.schedule }}"
  successfulJobsHistoryLimit: 1
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
    {{- if or (eq "stag" (include "util.environment" $)) (eq "uat" (include "util.environment" $)) }}
      ttlSecondsAfterFinished: 259200 #3days
    {{- else }}
      ttlSecondsAfterFinished: 604800 #7days
    {{- end }}
      template:
        metadata:
          labels:
            app.kubernetes.io/name: {{ printf "%s-%s" (include "util.name" $) $cronjobName }}
            app.kubernetes.io/instance: {{ $.Release.Name }}
            sidecar.istio.io/inject: "false"
        spec:
          restartPolicy: Never
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
          - name: secrets-volume
            secret:
              secretName: {{ include "util.fullname" $ }}
              items:
              - key: {{ $.Chart.Name }}.secrets.encrypted.yaml
                path: {{ $.Chart.Name }}.secrets.encrypted.yaml
          - name: config-volume
            configMap:
              name: {{ include "util.fullname" $ }}
              items:
              - key: {{ $.Chart.Name }}.common.config.yaml
                path: {{ $.Chart.Name }}.common.config.yaml
              - key: {{ $.Chart.Name }}.config.yaml
                path: {{ $.Chart.Name }}.config.yaml
          containers:
          - name: {{ $.Chart.Name }}-{{ $cronjobName }}
            image: {{ include "util.image" $ }}
            imagePullPolicy: IfNotPresent
            args:
              - gjob
              - {{ $cronjobValues.cmd }}
              - --commonConfigPath=/configs/{{ $.Chart.Name }}.common.config.yaml
              - --configPath=/configs/{{ $.Chart.Name }}.config.yaml
              - --secretsPath=/configs/{{ $.Chart.Name }}.secrets.encrypted.yaml
            {{- range $key, $arg := $cronjobValues.args }}
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
              requests:
                cpu: 10m
{{- end }}
{{- end }}
{{- end }}
