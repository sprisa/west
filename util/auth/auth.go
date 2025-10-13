package auth

import "github.com/golang-jwt/jwt/v5"


type TokenClaims struct {
	Endpoint string `json:"endpoint"`
	jwt.RegisteredClaims
}
