{{- define "gvList" -}}
{{- $groupVersions := . -}}

# GreptimeDB Operator API References

## Packages
{{- range $groupVersions }}
- {{ markdownRenderGVLink . }}
{{- end }}

{{ range $groupVersions }}
{{ template "gvDetails" . }}
{{ end }}

{{- end -}}
