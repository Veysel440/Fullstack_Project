package http

import (
	"net/http"
)

var openapiYAML = []byte(`
openapi: 3.0.3
info:
  title: Fullstack Oracle API
  version: "1.0.0"
  description: |
    Items API with JWT (access/refresh), rate limiting and Prometheus metrics.
    - Auth: send **Authorization: Bearer <access>**
    - Refresh: send **X-Refresh-Token** header (or JSON body { "refresh_token": ... }).
servers:
  - url: /api
    description: Reverse-proxied behind web (nginx)
  - url: /
    description: Direct API (dev)
tags:
  - name: auth
  - name: items
  - name: health

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  schemas:
    Error:
      type: object
      properties:
        error:      { type: string, example: not_found }
        code:       { type: string, example: not_found }
        fields:
          type: object
          additionalProperties: { type: string }
        request_id: { type: string }
    Tokens:
      type: object
      properties:
        access_token:  { type: string }
        refresh_token: { type: string }
      required: [access_token, refresh_token]
    Me:
      type: object
      properties:
        id:   { type: integer, format: int64 }
        role: { type: string, example: user }
    Item:
      type: object
      properties:
        id:         { type: integer, format: int64 }
        name:       { type: string }
        price:      { type: number, format: float }
        created_at: { type: string, format: date-time }
      required: [id, name, price, created_at]
    CreateItemDTO:
      type: object
      properties:
        name:  { type: string, minLength: 1, maxLength: 100 }
        price: { type: number, minimum: 0 }
      required: [name, price]
    PagedItems:
      type: object
      properties:
        items:
          type: array
          items: { $ref: '#/components/schemas/Item' }
        page:  { type: integer }
        size:  { type: integer }
        total: { type: integer, format: int64 }
      required: [items, page, size, total]

paths:
  /health:
    get:
      tags: [health]
      summary: Liveness check
      responses:
        '200': { description: OK }

  /auth/login:
    post:
      tags: [auth]
      summary: Login with email & password
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                email:    { type: string, format: email }
                password: { type: string }
              required: [email, password]
      responses:
        '200':
          description: Tokens
          content: { application/json: { schema: { $ref: '#/components/schemas/Tokens' } } }
        '401':
          description: Unauthorized
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }

  /auth/refresh:
    post:
      tags: [auth]
      summary: Rotate refresh token
      parameters:
        - in: header
          name: X-Refresh-Token
          schema: { type: string }
          required: false
      requestBody:
        required: false
        content:
          application/json:
            schema:
              type: object
              properties:
                refresh_token: { type: string }
      responses:
        '200':
          description: Tokens
          content: { application/json: { schema: { $ref: '#/components/schemas/Tokens' } } }
        '401':
          description: Unauthorized
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }

  /auth/me:
    get:
      tags: [auth]
      security: [ { bearerAuth: [] } ]
      summary: Current user
      responses:
        '200':
          description: The current user payload
          content: { application/json: { schema: { $ref: '#/components/schemas/Me' } } }
        '401': { description: Unauthorized }

  /items/:
    get:
      tags: [items]
      security: [ { bearerAuth: [] } ]
      summary: List items (paged)
      parameters:
        - in: query
          name: page
          schema: { type: integer, minimum: 1, default: 1 }
        - in: query
          name: size
          schema: { type: integer, minimum: 1, maximum: 100, default: 20 }
        - in: query
          name: sort
          description: "Field and direction, e.g. 'price,asc' or 'created_at,desc'"
          schema: { type: string }
        - in: query
          name: q
          description: "Search by name (ILIKE)"
          schema: { type: string }
      responses:
        '200':
          description: Paged items
          content: { application/json: { schema: { $ref: '#/components/schemas/PagedItems' } } }
        '401': { description: Unauthorized }
        '429':
          description: Rate limited
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
    post:
      tags: [items]
      security: [ { bearerAuth: [] } ]
      summary: Create item
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateItemDTO' }
      responses:
        '201':
          description: Created
          content: { application/json: { schema: { $ref: '#/components/schemas/Item' } } }
        '400':
          description: Validation error
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
        '401': { description: Unauthorized }

  /items/{id}:
    parameters:
      - in: path
        name: id
        required: true
        schema: { type: integer, format: int64 }
    get:
      tags: [items]
      security: [ { bearerAuth: [] } ]
      summary: Get item by id
      responses:
        '200':
          description: OK
          headers:
            ETag:
              description: Weak ETag for conditional GET
              schema: { type: string }
          content: { application/json: { schema: { $ref: '#/components/schemas/Item' } } }
        '304': { description: Not Modified }
        '404':
          description: Not found
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
    put:
      tags: [items]
      security: [ { bearerAuth: [] } ]
      summary: Update item
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateItemDTO' }
      responses:
        '200':
          description: OK
          content: { application/json: { schema: { $ref: '#/components/schemas/Item' } } }
        '404':
          description: Not found
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
    delete:
      tags: [items]
      security: [ { bearerAuth: [] } ]
      summary: Delete item (admin)
      responses:
        '204': { description: No Content }
        '403': { description: Forbidden }
        '404':
          description: Not found
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
`)

func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	_, _ = w.Write(openapiYAML)
}

func Docs(w http.ResponseWriter, _ *http.Request) {
	const html = `<!doctype html>
<html>
<head><meta charset="utf-8"><title>API Docs</title></head>
<body>
  <redoc spec-url="/openapi.yaml"></redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
