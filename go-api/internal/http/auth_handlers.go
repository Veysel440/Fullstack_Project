package http

import (
	"encoding/json"
	stdhttp "net/http"

	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/service"

	"github.com/go-playground/validator/v10"
)

type AuthHandlers struct {
	Cfg config.Config
	S   *service.AuthService
	Val *validator.Validate
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *AuthHandlers) Login(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	if err := a.Val.Struct(in); err != nil {
		writeValidation(w, r, map[string]string{"email": "required,email", "password": "required"})
		return
	}
	tok, err := a.S.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		writeError(w, r, 401, "unauthorized", "invalid credentials")
		return
	}
	writeJSON(w, stdhttp.StatusOK, tok)
}

func (a *AuthHandlers) Refresh(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	ref := r.Header.Get("X-Refresh-Token")
	if ref == "" {
		ref = body.RefreshToken
	}
	if ref == "" {
		writeValidation(w, r, map[string]string{"refresh_token": "required"})
		return
	}
	tok, err := a.S.Refresh(r.Context(), ref)
	if err != nil {
		writeError(w, r, 401, "unauthorized", "invalid refresh")
		return
	}
	writeJSON(w, stdhttp.StatusOK, tok)
}

func (a *AuthHandlers) Me(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"id":   userIDFrom(r.Context()),
		"role": roleFrom(r.Context()),
	})
}
