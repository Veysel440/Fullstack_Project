package http

import (
	"context"
	"errors"
	"log/slog"
	stdhttp "net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type authKey string

const (
	keyUserID authKey = "uid"
	keyRole   authKey = "role"
)

func userIDFrom(ctx context.Context) int64 {
	v, _ := ctx.Value(keyUserID).(int64)
	return v
}
func roleFrom(ctx context.Context) string {
	v, _ := ctx.Value(keyRole).(string)
	return v
}

type JWTVerifier struct {
	AccessSecret  []byte
	RefreshSecret []byte
	Logger        *slog.Logger
}

func (v *JWTVerifier) AuthRequired(roles ...string) func(stdhttp.Handler) stdhttp.Handler {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			raw := r.Header.Get("Authorization")
			if !strings.HasPrefix(raw, "Bearer ") {
				writeError(w, r, 401, "unauthorized", "missing bearer token")
				return
			}
			tokenStr := strings.TrimPrefix(raw, "Bearer ")
			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) { return v.AccessSecret, nil })
			if err != nil || !tok.Valid {
				writeError(w, r, 401, "unauthorized", "invalid token")
				return
			}
			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, r, 401, "unauthorized", "invalid claims")
				return
			}

			idAny, ok1 := claims["sub"]
			roleAny, ok2 := claims["role"]
			if !ok1 || !ok2 {
				writeError(w, r, 401, "unauthorized", "claims missing")
				return
			}

			var uid int64
			switch t := idAny.(type) {
			case float64:
				uid = int64(t)
			case int64:
				uid = t
			default:
				writeError(w, r, 401, "unauthorized", "bad sub")
				return
			}
			role, _ := roleAny.(string)
			if len(allowed) > 0 {
				if _, ok := allowed[role]; !ok {
					writeError(w, r, 403, "forbidden", "insufficient role")
					return
				}
			}
			ctx := context.WithValue(r.Context(), keyUserID, uid)
			ctx = context.WithValue(ctx, keyRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ParseRefreshToken(refresh string, secret []byte) (int64, string, error) {
	tok, err := jwt.Parse(refresh, func(t *jwt.Token) (any, error) { return secret, nil })
	if err != nil || !tok.Valid {
		return 0, "", errors.New("invalid")
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("invalid")
	}
	if claims["typ"] != "refresh" {
		return 0, "", errors.New("invalid")
	}
	var uid int64
	switch t := claims["sub"].(type) {
	case float64:
		uid = int64(t)
	case int64:
		uid = t
	default:
		return 0, "", errors.New("invalid")
	}
	role, _ := claims["role"].(string)
	return uid, role, nil
}
