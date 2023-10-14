{{/*
Defines the deployment resource for a typical application.
*/}}
{{- define "util.deployment" -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "util.fullname" . }}
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  replicas: {{ default .Values.global.replicaCount .Values.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
    {{- include "util.workloadMetadata" . | indent 6 }}
    spec:
    {{- with (default .Values.global.imagePullSecrets .Values.imagePullSecrets) }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      serviceAccountName: {{ include "util.serviceAccountName" . }}
      volumes:
        {{- include "util.workloadVolumes" . | nindent 8 }}
      initContainers:
        {{- include "util.workloadInitContainers" . | nindent 8 }}
      containers:
        {{- include "util.workloadContainers" . | nindent 8 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values }}
      {{- $context := (mustMerge (deepCopy .) $) }}
      affinity: {{- include "util.affinityNew" $context | nindent 8 }}
      tolerations: {{- include "util.tolerations" $context | nindent 8 }}
      {{- end }}
{{- if .Values.onDemandNodeDeployment }}
{{- if .Values.onDemandNodeDeployment.enabled }}
---
{{/*
The template for the on-demand node deployment.
We have several modifications for the init containers in this deployment:
- We don't need to run the {{service}}-migrate init container.
  The reason is in our migration scripts, there are some commands that cannot be run
  concurrently, like `CREATE INDEX CONCURRENTLY`...etc, so if we put that init container
  here, the migration scripts will be run concurrently on both spot and on-demand pods,
  which will cause migration errors.
- The wair-for-services init container is modified to wait for the same service that this
  deployment will deploy for. The reason is if that service is ready, that means the migration
  runs successfully somewhere (on the pods in spot nodes), so we can safely let this on-demand
  node deployment run after that.
- We also don't need to run the hasura migration init container, for same reason with above.
*/}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "util.fullname" . }}-on-demand-node
  labels:
{{ include "util.labels" . | indent 4 }}
spec:
  replicas: {{ default .Values.global.onDemandNodeDeployment.replicaCount .Values.onDemandNodeDeployment.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "util.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
    {{- include "util.workloadMetadata" . | indent 6 }}
    spec:
    {{- with (default .Values.global.imagePullSecrets .Values.imagePullSecrets) }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
    {{- end }}
      serviceAccountName: {{ include "util.serviceAccountName" . }}
      volumes:
        {{- include "util.workloadVolumes" . | nindent 8 }}
      {{- $svcName := .Chart.Name -}}
      {{- $svcPort := 0 -}}
      {{- if .Values.service -}}
      {{- $svcPort = .Values.service.port -}}
      {{- else if .Values.grpcPort -}}
      {{- $svcPort = .Values.grpcPort -}}
      {{- else if .Values.httpPort -}}
      {{- $svcPort = .Values.httpPort -}}
      {{- end -}}
      {{- $overwrites := dict "Values" (dict
          "waitForServices" (list (dict "name" $svcName "port" $svcPort ))
          "migrationEnabled" false
          "hasuraEnabled" false
        )
      }}
      {{- $context := mustMergeOverwrite (deepCopy .) $overwrites }}
      initContainers:
        {{- include "util.workloadInitContainers" $context | nindent 8 }}
      containers:
        {{- include "util.workloadContainers" . | nindent 8 }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- $overwrites := (dict
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
        )
      }}
      {{- with .Values }}
      {{- $context := mustMergeOverwrite (deepCopy $) $overwrites }}
      affinity: {{- include "util.affinityNew" $context | nindent 8 }}
      tolerations: {{- include "util.tolerations" $context | nindent 8 }}
      {{- end }}
{{- end -}}
{{- end -}}
{{- end -}}
