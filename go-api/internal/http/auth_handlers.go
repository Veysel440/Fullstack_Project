package http

import (
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"strings"

	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"
	"fullstack-oracle/go-api/internal/service"

	"github.com/go-playground/validator/v10"
)

type AuthHandlers struct {
	Cfg config.Config
	S   *service.AuthService
	Val *validator.Validate
}

func (h *AuthHandlers) Login(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var dto domain.LoginDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeValidation(w, r, map[string]string{"body": "invalid json"})
		return
	}
	if err := h.Val.Struct(dto); err != nil {
		writeValidation(w, r, toFields(err))
		return
	}
	u, tokens, err := h.S.Login(dto.Email, dto.Password)
	if errors.Is(err, repo.ErrNotFound) {
		writeError(w, r, 401, "unauthorized", "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, r, 500, "login_failed", err.Error())
		return
	}
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"user": u, "access_token": tokens.AccessToken, "refresh_token": tokens.RefreshToken,
	})
}

func (h *AuthHandlers) Refresh(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	rt := r.Header.Get("X-Refresh-Token")
	if rt == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		rt = strings.TrimSpace(body.RefreshToken)
	}
	if rt == "" {
		writeError(w, r, 400, "bad_request", "missing refresh_token")
		return
	}

	uid, role, err := ParseRefreshToken(rt, []byte(h.Cfg.JWTRefreshSecret))
	if err != nil {
		writeError(w, r, 401, "unauthorized", "invalid refresh")
		return
	}

	tokens := h.S.Refresh(uid, role)
	writeJSON(w, stdhttp.StatusOK, tokens)
}

func (h *AuthHandlers) Me(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]any{
		"user_id": userIDFrom(r.Context()),
		"role":    roleFrom(r.Context()),
	})
}
