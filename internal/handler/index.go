package handler

import (
	"html/template"
	"net/http"
	"sort"
)

const indexPageTemplate = `<!DOCTYPE html>
<html>
<head><title>FortiWeb Exporter</title></head>
<body>
<header>
<h1>FortiWeb Exporter</h1>
<p><a href="https://gitlab.com/endekasoft/fortiweb-exporter">gitlab.com/endekasoft/fortiweb-exporter</a></p>
</header>
<p>Configured targets:</p>
<ul>
{{- range .Targets }}
<li><a href="{{$.MetricsPath}}?target={{.}}">{{.}}</a></li>
{{- end }}
</ul>
</body>
</html>
`

var indexTemplate = template.Must(template.New("index").Parse(indexPageTemplate))

type indexPageData struct {
	Targets     []string
	MetricsPath string
}

// IndexHandler serves a landing page at "/" listing each configured
// FortiWeb target as a link to its /metrics?target=NAME endpoint.
type IndexHandler struct {
	data indexPageData
}

// NewIndexHandler builds an IndexHandler listing targetNames as links under
// metricsPath.
func NewIndexHandler(targetNames []string, metricsPath string) *IndexHandler {
	sorted := append([]string(nil), targetNames...)
	sort.Strings(sorted)
	return &IndexHandler{data: indexPageData{Targets: sorted, MetricsPath: metricsPath}}
}

// ServeHTTP implements http.Handler.
func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, h.data); err != nil {
		http.Error(w, "failed to render index page", http.StatusInternalServerError)
	}
}
