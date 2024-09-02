{{- define "gvList" -}}
{{- $groupVersions := . -}}

# GreptimeDB Operator API Reference

## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
