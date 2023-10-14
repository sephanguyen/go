{{/*
Define the docker tag to use for the service. Business service only.
Use the local chart's image tag if available, else fallback to the global image tag.
*/}}
{{- define "util.imageTag" -}}
{{- if .Values.image -}}
{{- default .Values.global.image.tag .Values.image.tag -}}
{{- else -}}
{{ .Values.global.image.tag }}
{{- end -}}
{{- end -}}

{{/*
Define the docker image to use for the service.
Use the local chart's image tag if available, else fallback to the global image tag.
*/}}
{{- define "util.image" -}}
{{ .Values.global.image.repository }}:{{ include "util.imageTag" . }}
{{- end -}}

{{/*
Defines the image that contains sops binary to decrypt secrets.
Usually it is "mozilla/sops:<tag>".
*/}}
{{- define "util.sopsImage" -}}
{{- $image := "mozilla/sops" -}} {{/* default image value */}}
{{- $tag := "v3.7.3-alpine" -}} {{/* default tag value */}}
{{- if and .Values.global .Values.global.sopsImage -}}
{{ default $image .Values.global.sopsImage.repository }}:{{ default $tag .Values.global.sopsImage.tag }}
{{- else -}}
{{ printf "%s:%s" $image $tag }}
{{- end -}}
{{- end -}}

{{/*
Defines the image that contains the wait-for.sh script.
This is usually used in init containers to make pods wait for certain services.
*/}}
{{- define "util.waitForImage" -}}
{{- $image := "asia.gcr.io/student-coach-e1e95/wait-for" -}} {{/* default image value */}}
{{- $tag := "0.0.2" -}} {{/* default tag value */}}
{{- if and .Values.global .Values.global.waitForImage -}}
{{ default $image .Values.global.waitForImage.repository }}:{{ default $tag .Values.global.waitForImage.tag }}
{{- else -}}
{{ printf "%s:%s" $image $tag }}
{{- end -}}
{{- end -}}
