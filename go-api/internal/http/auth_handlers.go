package http

import (
	"encoding/json"
	"fullstack-oracle/go-api/internal/repo"
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
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (a *AuthHandlers) validator() *validator.Validate {
	if a.Val == nil {
		a.Val = validator.New()
	}
	return a.Val
}

func (a *AuthHandlers) Register(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	if err := a.Val.Struct(in); err != nil {
		writeValidation(w, r, map[string]string{"email": "required,email", "password": "min=6"})
		return
	}
	if err := a.S.Register(r.Context(), in.Email, in.Password); err != nil {
		if err == repo.ErrConflict {
			writeError(w, r, 409, "email_exists", "email already registered")
			return
		}
		writeError(w, r, 500, "register_failed", err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusCreated, map[string]any{"ok": true})
}

func (a *AuthHandlers) Login(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	if err := a.validator().Struct(in); err != nil {
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

func (a *AuthHandlers) Logout(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	ref := r.Header.Get("X-Refresh-Token")
	if ref != "" {
		_ = a.S.Logout(r.Context(), ref)
	}
	w.WriteHeader(stdhttp.StatusNoContent)
}

func (a *AuthHandlers) Me(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"user_id": userIDFrom(r.Context()),
		"role":    roleFrom(r.Context()),
	})
}
