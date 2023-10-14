{{/*
Service resource for normal application deployments.
*/}}
{{- define "util.service" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "util.fullname" . }}
  labels:
    {{- include "util.labels" . | nindent 4 }}
spec:
  type: ClusterIP
  ports:
  {{- if .Values.httpPort }}
    - name: http-port
      protocol: TCP
      targetPort: http
      port: {{ .Values.httpPort }}
  {{- end }}
  {{- if .Values.grpcPort }}
    - name: grpc-web-port
      protocol: TCP
      targetPort: grpc
      port: {{ .Values.grpcPort }}
  {{- end }}
  selector:
    {{- include "util.selectorLabels" . | nindent 4 }}
{{- end }}

{{/*
Service resource for hasura.
*/}}
{{- define "util.hasuraService" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "util.fullname" . }}-hasura
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.hasura.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app.kubernetes.io/name: {{ include "util.name" . }}-hasura
    app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service resource for hasura v2.
*/}}
{{- define "util.hasurav2Service" -}}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "util.fullname" . }}-hasurav2
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  type: {{ .Values.hasurav2.service.type }}
  ports:
    - port: {{ .Values.hasurav2.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app.kubernetes.io/name: {{ include "util.name" . }}-hasurav2
    app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
