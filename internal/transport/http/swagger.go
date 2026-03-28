package http

import "net/http"

const openAPISpec = `openapi: 3.0.3
info:
  title: EventHub API
  version: 0.2.0
  description: Lab 1-2 endpoints for healthcheck and anonymous sessions.
servers:
  - url: http://localhost:8080
paths:
  /health:
    get:
      summary: Service healthcheck
      description: Returns service status and echoes X-Session-Id cookie when provided.
      parameters:
        - in: cookie
          name: X-Session-Id
          required: false
          schema:
            type: string
      responses:
        '200':
          description: Service is healthy
          headers:
            Set-Cookie:
              description: Present only when request includes valid X-Session-Id cookie.
              schema:
                type: string
          content:
            application/json:
              schema:
                type: object
                required: [status]
                properties:
                  status:
                    type: string
                    example: ok
  /session:
    post:
      summary: Create or refresh anonymous session
      description: Creates session on first visit or refreshes existing session TTL.
      parameters:
        - in: cookie
          name: X-Session-Id
          required: false
          schema:
            type: string
      responses:
        '201':
          description: New session created
          headers:
            Set-Cookie:
              description: X-Session-Id cookie with HttpOnly, Path=/ and Max-Age.
              schema:
                type: string
        '200':
          description: Existing session refreshed
          headers:
            Set-Cookie:
              description: X-Session-Id cookie with HttpOnly, Path=/ and Max-Age.
              schema:
                type: string
        '500':
          description: Internal server error
`

const swaggerHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>EventHub Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/openapi.yaml',
      dom_id: '#swagger-ui'
    });
  </script>
</body>
</html>
`

func (h *Handler) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(openAPISpec))
}

func (h *Handler) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerHTML))
}
