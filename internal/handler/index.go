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
<p><a href="{{.MetricsPath}}">Metrics</a></p>
<p>Configured targets:</p>
<ul>
{{- range .Targets }}
<li><a href="{{$.ProbePath}}?target={{.}}">{{.}}</a></li>
{{- end }}
</ul>
</body>
</html>
`

var indexTemplate = template.Must(template.New("index").Parse(indexPageTemplate))

type indexPageData struct {
	Targets     []string
	MetricsPath string
	ProbePath   string
}

// IndexHandler serves a landing page at "/" linking to the exporter's own
// metrics endpoint and, for each configured FortiWeb target, a link to its
// probe endpoint.
type IndexHandler struct {
	data indexPageData
}

// NewIndexHandler builds an IndexHandler listing targetNames as probePath
// links, alongside a link to metricsPath.
func NewIndexHandler(targetNames []string, metricsPath, probePath string) *IndexHandler {
	sorted := append([]string(nil), targetNames...)
	sort.Strings(sorted)
	return &IndexHandler{data: indexPageData{Targets: sorted, MetricsPath: metricsPath, ProbePath: probePath}}
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
