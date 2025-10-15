package auth

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	Endpoint string `json:"endpoint"`
	IP       string `json:"ip"`
	Ca       string `json:"ca"`
	PortIP   string `json:"port_ip"`
	jwt.RegisteredClaims
}
