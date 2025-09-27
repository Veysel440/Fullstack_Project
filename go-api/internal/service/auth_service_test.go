package service

import (
	"context"
	"testing"

	"fullstack-oracle/go-api/internal/cache"
	"fullstack-oracle/go-api/internal/config"
)

func TestRefresh_Rotates(t *testing.T) {
	cfg := config.Config{
		JWTAccessSecret:   "a",
		JWTRefreshSecret:  "b",
		JWTAccessTTLMin:   1,
		JWTRefreshTTLDays: 1,
	}
	s := &AuthService{cfg: cfg, r: nil, c: &cache.Store{}}

	tok, err := s.issue(1, "user")
	if err != nil {
		t.Fatal(err)
	}
	nt, err := s.Refresh(context.Background(), tok.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if nt.RefreshToken == tok.RefreshToken {
		t.Fatal("refresh token not rotated")
	}
}
