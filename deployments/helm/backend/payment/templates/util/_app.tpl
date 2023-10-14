{{- define "util.appWithToggle" -}}
{{- if .Values.enabled -}}
{{- include "util.app" . -}}
{{- end -}}
{{- end -}}

{{- define "util.app" -}}
{{ include "util.configMap" . }}
---
{{ include "util.secret" . }}
---
{{ include "util.pdb" . }}
---
{{ include "util.serviceAccount" . }}
---
{{ include "util.deployment" . }}
---
{{ include "util.service" . }}
{{- if .Values.apiHttp }}
---
{{ include "virtualservice.api.tpl" . }}
{{- end }}
{{- if .Values.webHttp }}
---
{{ include "virtualservice.web.tpl" . }}
{{- end }}
{{- if .Values.global.vpa.enabled }}
---
{{ include "util.vpa" . }}
{{- end }}
{{ include "util.jobs" . }}
---
{{- if .Capabilities.APIVersions.Has "keda.sh/v1alpha1" }}
{{- if .Values.onDemandNodeDeployment }}
{{- if .Values.onDemandNodeDeployment.enabled }}
{{ include "util.keda.scaledObject" . }}
{{- end }}
{{- end }}
{{/*
Keda is using HPA under the hood, so if we use Keda, we can't also using HPA.
See https://keda.sh/docs/2.9/faq/.
*/}}
{{- else if .Values.hpa }}
{{ include "util.hpa" . }}
{{- end }}
{{- if or .Values.hasuraEnabled .Values.hasurav2Enabled }}
---
{{ include "util.hasurav2ServiceAccount" . }}
{{- end }}
{{- if .Values.hasuraEnabled }}
---
{{ include "util.hasura" . }}
{{- end }}
{{- if .Values.hasurav2Enabled }}
---
{{ include "util.hasurav2" . }}
{{- end }}
{{ include "util.cronjobs" . }}
{{- if .Values.caching }}
{{- if .Values.caching.enabled }}
---
{{ include "util.caching" . }}
{{- end }}
{{- end }}
{{- end -}}
