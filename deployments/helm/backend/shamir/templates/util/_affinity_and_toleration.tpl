{{/*
Common template for deployments/statefulsets' tolerations.
Note that if .tolerations is an empty slice, it is not overriden, while
a nil .tolerations will be overriden by .Values.global.tolerations.
*/}}
{{- define "util.tolerations" -}}
{{- if kindIs "slice" .tolerations -}}
  {{- .tolerations | toYaml }}
{{- else if kindIs "slice" .Values.tolerations -}}
  {{- .Values.tolerations | toYaml }}
{{- else if .Values.global -}}
  {{- if kindIs "slice" .Values.global.tolerations -}}
    {{- .Values.global.tolerations | toYaml }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Similar to util.affinityNew, but for hasura.
The following is for hasura deployment affinity.
*/}}
{{- define "util.hasuraAffinity" -}}
{{- $podAffinityIdentifier := printf "%s-hasura" (include "util.name" $) -}}
{{- $commonPodAntiAffinity := dict "preferredDuringSchedulingIgnoredDuringExecution" (list (dict
  "podAffinityTerm" (dict
    "labelSelector" (dict "matchLabels" (dict "app.kubernetes.io/name" $podAffinityIdentifier))
    "topologyKey" "kubernetes.io/hostname"
  )
  "weight" 100
))
-}}
{{- if eq "true" (include "util.requirePodAntiAffinity" .) -}}
{{- $commonPodAntiAffinity = dict "requiredDuringSchedulingIgnoredDuringExecution" (list (dict
  "labelSelector" (dict "matchLabels" (dict "app.kubernetes.io/name" $podAffinityIdentifier))
  "topologyKey" "kubernetes.io/hostname"
))
-}}
{{- end -}}
{{- $affinity := dict -}}
{{- if $.Values.hasura -}}
  {{- if $.Values.hasura.affinityOverride -}}
    {{- $affinity = deepCopy $.Values.hasura.affinityOverride -}}
  {{- end -}}
{{- end -}}
{{- $affinity = merge $affinity (deepCopy (default dict $.Values.global.hasura.affinityOverride)) -}}
{{- if not $affinity -}}
  {{- if $.Values.hasura -}}
    {{- if $.Values.hasura.affinity -}}
      {{- $affinity = deepCopy $.Values.hasura.affinity -}}
    {{- end -}}
  {{- end -}}
  {{- $affinity = merge $affinity (deepCopy (default dict $.Values.global.hasura.affinity)) -}}
{{- end -}}
{{- mustMerge $affinity (dict "podAntiAffinity" $commonPodAntiAffinity) | toYaml }}
{{- end }}

{{/*
Common template for hasura deployment tolerations.
*/}}
{{- define "util.hasuraTolerations" -}}
{{- $tolerations := $.Values.global.hasura.tolerations -}}
{{- if $.Values.hasura -}}
    {{- if $.Values.hasura.tolerations -}}
        {{- $tolerations = $.Values.hasura.tolerations -}}
    {{- end -}}
{{- end -}}
{{- toYaml $tolerations }}
{{- end }}

{{/*
Similar to util.affinityNew, but for hasura v2 deployment affinity.
*/}}
{{- define "util.hasurav2Affinity" -}}
{{- $podAffinityIdentifier := printf "%s-hasurav2" (include "util.name" $) -}}
{{- $commonPodAntiAffinity := dict "preferredDuringSchedulingIgnoredDuringExecution" (list (dict
  "podAffinityTerm" (dict
    "labelSelector" (dict "matchLabels" (dict "app.kubernetes.io/name" $podAffinityIdentifier))
    "topologyKey" "kubernetes.io/hostname"
  )
  "weight" 100
))
-}}
{{- if eq "true" (include "util.requirePodAntiAffinity" .) -}}
{{- $commonPodAntiAffinity = dict "requiredDuringSchedulingIgnoredDuringExecution" (list (dict
  "labelSelector" (dict "matchLabels" (dict "app.kubernetes.io/name" $podAffinityIdentifier))
  "topologyKey" "kubernetes.io/hostname"
))
-}}
{{- end -}}
{{- $affinity := dict -}}
{{- if $.Values.hasura -}}
  {{- if $.Values.hasura.affinityOverride -}}
    {{- $affinity = deepCopy $.Values.hasura.affinityOverride -}}
  {{- end -}}
{{- end -}}
{{- $affinity = merge $affinity (deepCopy (default dict $.Values.global.hasura.affinityOverride)) -}}
{{- if not $affinity -}}
  {{- if $.Values.hasura -}}
    {{- if $.Values.hasura.affinity -}}
      {{- $affinity = deepCopy $.Values.hasura.affinity -}}
    {{- end -}}
  {{- end -}}
  {{- $affinity = merge $affinity (deepCopy (default dict $.Values.global.hasura.affinity)) -}}
{{- end -}}
{{- mustMerge $affinity (dict "podAntiAffinity" $commonPodAntiAffinity) | toYaml }}
{{- end }}

{{/*
Common template for hasura v2 deployment tolerations.
*/}}
{{- define "util.hasurav2Tolerations" -}}
{{- include "util.hasuraTolerations" . }}
{{- end }}

{{/*
commonPodAntiAffinityStrategy returns either "requiredDuringSchedulingIgnoredDuringExecution"
or "preferredDuringSchedulingIgnoredDuringExecution", based on the flag requirePodAntiAffinity.
*/}}
{{- define "util.commonPodAntiAffinityStrategy" -}}
{{- if eq "true" (include "util.requirePodAntiAffinity" .) -}}
requiredDuringSchedulingIgnoredDuringExecution
{{- else -}}
preferredDuringSchedulingIgnoredDuringExecution
{{- end -}}
{{- end -}}

{{/*
Common template for deployments/statefulsets' affinity.

When .Values.global.affinity or .Values.affinity is nil, it defaults to a
nodeAffinity to use spot instances on GKE. This should be disabled in local.

By default, it also adds a common "podAntiAffinity" that prefers pods
of the same deployments to go to different nodes.
If you want to add/edit podAntiAffinity, careful with how helm's merge works.

The reason for having both affinity and affinityOverride is because we need to override .Values.global.affinity.
Here, we merge key-by-key, checking whether the key exists (non-nil) before merging
from global value.
*/}}
{{- define "util.affinityNew" -}}
{{- $affinity := dict -}}
{{- $podAffinityIdentifier := default (include "util.fullname" .) .podAffinityIdentifier -}}
{{- $commonPodAntiAffinity := dict "preferredDuringSchedulingIgnoredDuringExecution" (list (dict
  "podAffinityTerm" (dict
    "labelSelector" (dict "matchLabels" (dict "app.kubernetes.io/name" $podAffinityIdentifier))
    "topologyKey" "kubernetes.io/hostname"
  )
  "weight" 100
))
-}}
{{- if eq "true" (include "util.requirePodAntiAffinity" .) -}}
{{- $commonPodAntiAffinity = dict "requiredDuringSchedulingIgnoredDuringExecution" (list (dict
  "labelSelector" (dict "matchLabels" (dict "app.kubernetes.io/name" $podAffinityIdentifier))
  "topologyKey" "kubernetes.io/hostname"
))
-}}
{{- end -}}
{{- if .affinityOverride -}}
  {{- $affinity = deepCopy .affinityOverride -}}
{{- end -}}
{{- if .Values.global -}}
  {{- if .Values.global.affinityOverride -}}
    {{- if and (not (hasKey $affinity "nodeAffinity")) (hasKey .Values.global.affinityOverride "nodeAffinity") -}}
      {{- $_ := set $affinity "nodeAffinity" .Values.global.affinityOverride.nodeAffinity -}}
    {{- end -}}
    {{- if and (not (hasKey $affinity "podAffinity")) (hasKey .Values.global.affinityOverride "podAffinity") -}}
      {{- $_ := set $affinity "podAffinity" .Values.global.affinityOverride.podAffinity -}}
    {{- end -}}
    {{- if and (not (hasKey $affinity "podAntiAffinity")) (hasKey .Values.global.affinityOverride "podAntiAffinity") -}}
      {{- $_ := set $affinity "podAntiAffinity" .Values.global.affinityOverride.podAntiAffinity -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- if not $affinity -}}
  {{- if .affinity -}}
    {{- $affinity = deepCopy .affinity -}}
  {{- end -}}
  {{- if .Values.global -}}
    {{- if .Values.global.affinity -}}
      {{- if and (not (hasKey $affinity "nodeAffinity")) (hasKey .Values.global.affinity "nodeAffinity") -}}
        {{- $_ := set $affinity "nodeAffinity" .Values.global.affinity.nodeAffinity -}}
      {{- end -}}
      {{- if and (not (hasKey $affinity "podAffinity")) (hasKey .Values.global.affinity "podAffinity") -}}
        {{- $_ := set $affinity "podAffinity" .Values.global.affinity.podAffinity -}}
      {{- end -}}
      {{- if and (not (hasKey $affinity "podAntiAffinity")) (hasKey .Values.global.affinity "podAntiAffinity") -}}
        {{- $_ := set $affinity "podAntiAffinity" .Values.global.affinity.podAntiAffinity -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
{{- end -}}

{{- if not (hasKey $affinity "podAntiAffinity") -}}
  {{- $_ := set $affinity "podAntiAffinity" $commonPodAntiAffinity -}}
{{- end -}}
{{- toYaml $affinity }}
{{- end }}

{{/*
Convenience function
*/}}
{{- define "util.requirePodAntiAffinity" -}}
{{- if hasKey .Values "requirePodAntiAffinity" -}}
  {{- .Values.requirePodAntiAffinity -}}
{{- else if hasKey .Values "global" -}}
  {{- .Values.global.requirePodAntiAffinity -}}
{{- else -}}
  {{- "false" -}}
{{- end -}}
{{- end -}}
