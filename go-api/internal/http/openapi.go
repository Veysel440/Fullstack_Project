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
    Simple Items API with JWT auth (access/refresh), rate limiting and metrics.
    - Auth: Bearer access token (Authorization: Bearer <token>)
    - Refresh: send **X-Refresh-Token** header.
servers:
  - url: /api
    description: Reverse-proxied behind web (nginx)
  - url: /
    description: Direct API (dev)

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
        error: { type: string, example: not_found }
        fields:
          type: object
          additionalProperties: { type: string }
        request_id: { type: string, description: Optional request correlation id }
    Tokens:
      type: object
      properties:
        access_token: { type: string }
        refresh_token: { type: string }
      required: [access_token, refresh_token]
    User:
      type: object
      properties:
        id: { type: integer, format: int64 }
        email: { type: string, format: email }
        role: { type: string, example: user }
        created_at: { type: string, format: date-time }
    Item:
      type: object
      properties:
        id: { type: integer, format: int64 }
        name: { type: string }
        price: { type: number, format: float }
        created_at: { type: string, format: date-time }
      required: [id, name, price, created_at]
    CreateItemDTO:
      type: object
      properties:
        name: { type: string, minLength: 1 }
        price: { type: number, format: float, minimum: 0 }
      required: [name, price]

paths:
  /auth/login:
    post:
      summary: Login and receive access/refresh tokens
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [email, password]
              properties:
                email: { type: string, format: email }
                password: { type: string, format: password }
      responses:
        '200':
          description: Tokens
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Tokens' }
        '401':
          description: Unauthorized
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }

  /auth/refresh:
    post:
      summary: Rotate tokens using refresh token in header
      parameters:
        - in: header
          name: X-Refresh-Token
          required: true
          schema: { type: string }
          description: Refresh JWT
      responses:
        '200':
          description: New rotated tokens
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Tokens' }
        '401':
          description: Invalid/expired/blacklisted refresh
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }

  /auth/me:
    get:
      summary: Current user info
      security: [ { bearerAuth: [] } ]
      responses:
        '200':
          description: User
          content:
            application/json:
              schema: { $ref: '#/components/schemas/User' }
        '401':
          description: Unauthorized
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }

  /items/:
    get:
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
          schema:
            type: string
            enum: [created_at, -created_at, price, -price, name, -name]
          description: >
            Optional sorting. A leading '-' means descending.
            **Not all deployments may implement sorting yet.**
      responses:
        '200':
          description: Items
          content:
            application/json:
              schema:
                type: array
                items: { $ref: '#/components/schemas/Item' }
        '429':
          description: Rate limited
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
    post:
      summary: Create item
      security: [ { bearerAuth: [] } ]
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateItemDTO' }
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Item' }
        '400':
          description: Validation error
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
        '401':
          description: Unauthorized
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }

  /items/{id}:
    get:
      summary: Get item by id
      parameters:
        - in: path
          name: id
          required: true
          schema: { type: integer, format: int64 }
      responses:
        '200':
          description: OK
          headers:
            ETag: { description: Weak ETag for conditional GET, schema: { type: string } }
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Item' }
        '304': { description: Not Modified }
        '404':
          description: Not found
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
    put:
      summary: Update item
      security: [ { bearerAuth: [] } ]
      parameters:
        - in: path
          name: id
          required: true
          schema: { type: integer, format: int64 }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/CreateItemDTO' }
      responses:
        '200':
          description: Updated
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Item' }
        '404':
          description: Not found
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
    delete:
      summary: Delete item (admin only)
      description: Requires role **admin** (enforced by middleware).
      security: [ { bearerAuth: [] } ]
      parameters:
        - in: path
          name: id
          required: true
          schema: { type: integer, format: int64 }
      responses:
        '204': { description: Deleted }
        '403':
          description: Forbidden
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
        '404':
          description: Not found
          content: { application/json: { schema: { $ref: '#/components/schemas/Error' } } }
`)

// Serves /openapi.yaml
func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	_, _ = w.Write(openapiYAML)
}

// Very small docs page (Redoc)
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
