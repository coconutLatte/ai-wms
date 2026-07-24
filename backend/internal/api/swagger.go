// Package api provides HTTP handlers and route registration for the WMS API.
package api

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed swagger_spec.json
var swaggerSpec []byte

// swaggerUITemplate is the HTML page for Swagger UI.
// It loads swagger-ui-dist from unpkg CDN and points to the embedded spec.
const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>AI-WMS API Documentation</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" crossorigin></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: './swagger.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        plugins: [SwaggerUIBundle.plugins.DownloadUrl],
        layout: "StandaloneLayout",
        defaultModelsExpandDepth: -1,
        docExpansion: "list"
      });
    };
  </script>
</body>
</html>
`

// RegisterSwaggerRoutes registers Swagger UI API documentation routes on the given mux.
// It serves the OpenAPI 3.0 spec at /api/docs/swagger.json and the Swagger UI at /api/docs.
// These routes are public (no auth required) so developers can browse the API docs.
func RegisterSwaggerRoutes(mux *http.ServeMux) {
	// Swagger JSON spec
	mux.HandleFunc("GET /api/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(swaggerSpec)
	})

	// Swagger UI page
	mux.HandleFunc("GET /api/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		tmpl, _ := template.New("swagger").Parse(swaggerUITemplate)
	_ = tmpl.Execute(w, nil)
	})
}
