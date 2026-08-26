package auth

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	Endpoint          string   `json:"endpoint"`
	EndpointAddresses []string `json:"endpoint_addresses"`
	IP                string   `json:"ip"`
	Ca                string   `json:"ca"`
	jwt.RegisteredClaims
}
