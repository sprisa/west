package info

import (
	"bytes"
	"io"
	"net"
	"net/http"
)

func GetPublicIP() (addr net.IP, err error) {
	res, err := http.Get("https://checkip.amazonaws.com")
	if err != nil {
		return
	}
	defer res.Body.Close()

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	publicIP := bytes.TrimSpace(body)
	return net.ParseIP(string(publicIP)), nil
}
