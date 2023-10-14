{{/*
PodDisruptionBudget definitions.
*/}}
{{- define "util.pdb" -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "util.fullname" . }}
spec:
  maxUnavailable: {{ default 1 .Values.pdbMaxUnavailable }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "util.hasuraPdb" -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "util.fullname" . }}-hasura
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}-hasura
      app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "util.hasurav2Pdb" -}}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "util.fullname" . }}-hasurav2
spec:
  maxUnavailable: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}-hasurav2
      app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
