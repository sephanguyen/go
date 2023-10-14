{{- define "corsPolicy" -}}
allowOrigins:
  - regex: ".*"
allowMethods:
  - POST
  - GET
  - OPTIONS
  - PUT
  - DELETE
allowHeaders:
  - grpc-timeout
  - content-type
  - keep-alive
  - user-agent
  - cache-control
  - content-type
  - content-transfer-encoding
  - token
  - x-accept-content-transfer-encoding
  - x-accept-response-streaming
  - x-user-agent
  - x-grpc-web
  - pkg
  - version
maxAge: 100s
exposeHeaders:
  - grpc-status
  - grpc-message
{{- end -}}

{{- define "util.destinationrule.api.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: {{ .Chart.Name }}-api
spec:
  host: {{ .Values.service.grpcHost }}
  trafficPolicy:
    portLevelSettings:
      - port:
          number: {{ .Values.service.port }}
{{- toYaml .Values.trafficPolicy | nindent 8 }}
{{- end -}}

{{- define "util.destinationrule.web.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: {{ .Chart.Name }}-web
spec:
  host: {{ .Values.service.grpcWebHost }}
  trafficPolicy:
    portLevelSettings:
      - port:
          number: {{ .Values.service.grpcWebPort }}
{{- toYaml .Values.trafficPolicy | nindent 8 }}
{{- end -}}

{{- define "virtualservice.api.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-api
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.api | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.apiHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{- define "virtualservice.web.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-web
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.webApi | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.webHttp }}
  http:
{{- range $i, $http := . }}
  - match:
  {{- toYaml $http.match | nindent 4 }}
    route:
  {{- toYaml $http.route | nindent 4 }}
    corsPolicy:
  {{- if $http.corsPolicy }}
  {{- toYaml $http.corsPolicy | nindent 6 }}
  {{- else }}
  {{- include "corsPolicy" . | nindent 6 }}
  {{- end }}
{{- end }}
{{- end }}
{{- end -}}

{{- define "virtualservice.admin.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-admin
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.admin | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.adminHttp }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}


{{- define "virtualservice.adminv2.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-adminv2
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.admin | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.adminHttpV2 }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}


{{- define "virtualservice.teacher.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-web
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.teacher | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.httpRoute }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{- define "virtualservice.learner.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-web
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.learner | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.httpRoute }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{- define "virtualservice.backoffice.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-web
spec:
  hosts:
{{ if eq .Values.hostname "backoffice" }}
{{ toYaml .Values.global.dnsNames.backoffice | indent 4 }}
{{ else }}
{{ toYaml .Values.global.dnsNames.backofficeMfe | indent 4 }}
{{ end }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
{{- with .Values.httpRoute }}
  http:
{{ toYaml . | indent 4 }}
{{- end }}
{{- end -}}

{{- define "virtualservice.grafana.tpl" -}}
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: {{ .Chart.Name }}-grafana
spec:
  hosts:
{{ toYaml .Values.global.dnsNames.grafana | indent 4 }}
  gateways:
    - istio-system/{{ .Values.global.environment }}-{{ .Values.global.vendor }}-gateway
  exportTo:
    - istio-system
  http:
  - match:
    - uri:
        prefix: "/"
    route:
    - destination:
        host: grafana
        port:
          number: {{ .Values.service.port }}
{{- end -}}
