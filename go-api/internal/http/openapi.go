package http

import (
	"net/http"
)

const openAPISpecYAML = `openapi: 3.0.3
info:
  title: Fullstack-PG API
  version: 1.0.0
servers:
  - url: /api
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
paths:
  /health:
    get:
      summary: Liveness probe
      responses:
        '200': { description: OK }
  /auth/login:
    post:
      summary: Login
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password]
              properties:
                email: { type: string, format: email }
                password: { type: string, minLength: 6 }
      responses:
        '200':
          description: tokens
          content:
            application/json:
              schema:
                type: object
                properties:
                  access_token: { type: string }
                  refresh_token: { type: string }
        '401': { description: invalid credentials }
  /auth/refresh:
    post:
      summary: Refresh tokens
      responses:
        '200': { description: new tokens }
        '401': { description: invalid refresh }
  /auth/me:
    get:
      security: [{ bearerAuth: [] }]
      responses: { '200': { description: user info } }
  /items/:
    get:
      summary: List items
      parameters:
        - in: query
          name: page
          schema: { type: integer, minimum: 1, default: 1 }
        - in: query
          name: size
          schema: { type: integer, minimum: 1, maximum: 100, default: 20 }
      responses: { '200': { description: array of items } }
    post:
      security: [{ bearerAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, price]
              properties:
                name: { type: string }
                price: { type: number }
      responses: { '201': { description: created } }
  /items/{id}:
    parameters:
      - in: path
        name: id
        required: true
        schema: { type: integer }
    get:
      responses: { '200': { description: item }, '404': { description: not found } }
    put:
      security: [{ bearerAuth: [] }]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, price]
              properties:
                name: { type: string }
                price: { type: number }
      responses: { '200': { description: updated } }
    delete:
      security: [{ bearerAuth: [] }]
      responses: { '204': { description: deleted }, '403': { description: forbidden } }
`

func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write([]byte(openAPISpecYAML))
}

const redocHTML = `<!doctype html>
<html><head><meta charset="utf-8"/><title>API docs</title>
<link rel="icon" href="data:,">
<style>body,html,#rd{height:100%;margin:0}</style>
</head><body>
<div id="rd"></div>
<script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
<script>Redoc.init('/api/openapi.yaml', {}, document.getElementById('rd'));</script>
</body></html>`

func Docs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(redocHTML))
}
