package http

import (
	"encoding/json"
	stdhttp "net/http"
)

type ErrorBody struct {
	Error     string            `json:"error"`
	Code      string            `json:"code"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

func writeJSON(w stdhttp.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w stdhttp.ResponseWriter, r *stdhttp.Request, status int, code, msg string) {
	writeJSON(w, status, ErrorBody{
		Error: msg, Code: code, RequestID: RequestIDFrom(r.Context()),
	})
}

func writeValidation(w stdhttp.ResponseWriter, r *stdhttp.Request, fields map[string]string) {
	writeJSON(w, stdhttp.StatusBadRequest, ErrorBody{
		Error: "validation failed", Code: "bad_request",
		Fields: fields, RequestID: RequestIDFrom(r.Context()),
	})
}
