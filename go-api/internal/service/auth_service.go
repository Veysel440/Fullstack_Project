package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/repo"
)

type RevocationStore interface {
	IsRevoked(ctx context.Context, jti string) (bool, error)
	RevokeJTI(ctx context.Context, jti string, ttl time.Duration) error
}

type AuthService struct {
	cfg config.Config
	r   *repo.UserRepo
	c   RevocationStore
}

func NewAuthService(cfg config.Config, r *repo.UserRepo, c RevocationStore) *AuthService {
	return &AuthService{cfg: cfg, r: r, c: c}
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Register(ctx context.Context, email, pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.r.Create(ctx, email, string(hash), "user")
	return err
}

func (s *AuthService) Login(ctx context.Context, email, pass string) (Tokens, error) {
	u, hash, err := s.r.FindByEmail(ctx, email)
	if err != nil {
		return Tokens{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)) != nil {
		return Tokens{}, repo.ErrNotFound
	}
	return s.issue(u.ID, u.Role)
}

func (s *AuthService) Refresh(ctx context.Context, oldRefresh string) (Tokens, error) {
	tok, err := jwt.ParseWithClaims(oldRefresh, jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(s.cfg.JWTRefreshSecret), nil
	})
	if err != nil || !tok.Valid {
		return Tokens{}, repo.ErrNotFound
	}
	claims := tok.Claims.(jwt.MapClaims)
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return Tokens{}, repo.ErrNotFound
	}
	if s.c != nil {
		rev, err := s.c.IsRevoked(ctx, jti)
		if err != nil || rev {
			return Tokens{}, repo.ErrNotFound
		}
	}

	uid := int64(claims["sub"].(float64))
	role, _ := claims["role"].(string)

	if s.c != nil {
		_ = s.c.RevokeJTI(ctx, jti, time.Duration(s.cfg.JWTRefreshTTLDays)*24*time.Hour)
	}
	return s.issue(uid, role)
}

func (s *AuthService) Logout(ctx context.Context, refresh string) error {
	tok, _, err := jwt.NewParser().ParseUnverified(refresh, jwt.MapClaims{})
	if err != nil {
		return err
	}
	claims := tok.Claims.(jwt.MapClaims)
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return nil
	}
	return s.c.RevokeJTI(ctx, jti, time.Duration(s.cfg.JWTRefreshTTLDays)*24*time.Hour)
}

func (s *AuthService) issue(uid int64, role string) (Tokens, error) {
	now := time.Now()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid, "role": role, "typ": "access", "iat": now.Unix(),
		"exp": now.Add(time.Duration(s.cfg.JWTAccessTTLMin) * time.Minute).Unix(),
	})
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid, "role": role, "typ": "refresh", "iat": now.Unix(),
		"exp": now.Add(time.Duration(s.cfg.JWTRefreshTTLDays) * 24 * time.Hour).Unix(),
		"jti": randString(32),
	})
	ak, err := access.SignedString([]byte(s.cfg.JWTAccessSecret))
	if err != nil {
		return Tokens{}, err
	}
	rk, err := refresh.SignedString([]byte(s.cfg.JWTRefreshSecret))
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: ak, RefreshToken: rk}, nil
}

func randString(n int) string {
	const a = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	seed := time.Now().UnixNano()
	for i := range b {
		seed = seed*1664525 + 1013904223
		b[i] = a[seed%int64(len(a))]
	}
	return string(b)
}
