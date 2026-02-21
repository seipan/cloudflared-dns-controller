{{/*
Expand the name of the chart.
*/}}
{{- define "cloudflared-dns-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "cloudflared-dns-controller.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cloudflared-dns-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cloudflared-dns-controller.labels" -}}
helm.sh/chart: {{ include "cloudflared-dns-controller.chart" . }}
{{ include "cloudflared-dns-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cloudflared-dns-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cloudflared-dns-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "cloudflared-dns-controller.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cloudflared-dns-controller.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Determine the secret name for Cloudflare credentials
*/}}
{{- define "cloudflared-dns-controller.secretName" -}}
{{- if .Values.cloudflare.existingSecret }}
{{- .Values.cloudflare.existingSecret }}
{{- else }}
{{- include "cloudflared-dns-controller.fullname" . }}-cloudflare
{{- end }}
{{- end }}

{{/*
Container image
*/}}
{{- define "cloudflared-dns-controller.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}
