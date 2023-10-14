{{- define "util.workloadVolumes" -}}
{{- if .Values.migrationEnabled }}
- name: secrets-volume
  secret:
    secretName: {{ .Chart.Name }}
    items:
    - key: {{ .Chart.Name }}.secrets.encrypted.yaml
      path: {{ .Chart.Name }}.secrets.encrypted.yaml
    - key: {{ .Chart.Name }}_migrate.secrets.encrypted.yaml
      path: {{ .Chart.Name }}_migrate.secrets.encrypted.yaml
{{- if and .Values.hasuraEnabled (ne "draft" .Chart.Name) }}
    - key: hasura.secrets.encrypted.yaml
      path: hasura.secrets.encrypted.yaml
- name: hasura-secrets-decrypted-volume
  emptyDir: {}
- name: hasura-metadata
  configMap:
    name: {{ .Chart.Name }}-hasura-metadata
{{- end }}
{{- else }}
- name: secrets-volume
  secret:
    secretName: {{ .Chart.Name }}
    items:
    - key: {{ .Chart.Name }}.secrets.encrypted.yaml
      path: {{ .Chart.Name }}.secrets.encrypted.yaml
  {{- if .Values.hasuraEnabled }}
    - key: hasura.secrets.encrypted.yaml
      path: hasura.secrets.encrypted.yaml
  {{- end }}
{{- end }}
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
{{- end }}
