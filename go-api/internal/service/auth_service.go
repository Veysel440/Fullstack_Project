package service

import (
	"time"

	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/domain"
	"fullstack-oracle/go-api/internal/repo"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	cfg config.Config
	ur  *repo.UserRepo
}

func NewAuthService(cfg config.Config, ur *repo.UserRepo) *AuthService {
	return &AuthService{cfg: cfg, ur: ur}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Login(email, password string) (*domain.User, TokenPair, error) {
	u, hash, err := s.ur.FindByEmail(nil, email)
	if err != nil {
		return nil, TokenPair{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, TokenPair{}, repo.ErrNotFound
	}
	return u, s.tokens(u), nil
}

func (s *AuthService) Refresh(userID int64, role string) TokenPair {
	u := &domain.User{ID: userID, Role: role}
	return s.tokens(u)
}

func (s *AuthService) tokens(u *domain.User) TokenPair {
	now := time.Now()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  u.ID,
		"role": u.Role,
		"exp":  now.Add(time.Duration(s.cfg.JWTAccessTTLMin) * time.Minute).Unix(),
		"iat":  now.Unix(),
	})
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  u.ID,
		"role": u.Role,
		"exp":  now.Add(time.Duration(s.cfg.JWTRefreshTTLDays) * 24 * time.Hour).Unix(),
		"iat":  now.Unix(),
		"typ":  "refresh",
	})
	ats, _ := access.SignedString([]byte(s.cfg.JWTAccessSecret))
	rts, _ := refresh.SignedString([]byte(s.cfg.JWTRefreshSecret))
	return TokenPair{AccessToken: ats, RefreshToken: rts}
}
