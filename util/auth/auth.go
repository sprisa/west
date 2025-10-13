package auth

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	Endpoint string `json:"endpoint"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	jwt.RegisteredClaims
}
