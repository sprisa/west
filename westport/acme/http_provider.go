package acme

import (
	"net/http"
	"sync"

	l "github.com/sprisa/west/util/log"
)

// HTTPProvider implements the challenge.Provider interface for HTTP-01 challenges
type HTTPProvider struct {
	mu     sync.RWMutex
	tokens map[string]string // maps token to key authorization
}

func NewHTTPProvider() *HTTPProvider {
	return &HTTPProvider{
		tokens: make(map[string]string),
	}
}

// Present creates the challenge response
func (p *HTTPProvider) Present(domain, token, keyAuth string) error {
	l.Log.Info().
		Str("domain", domain).
		Str("token", token).
		Str("keyAuth", token).
		Msg("HTTP challenge present")

	p.mu.Lock()
	p.tokens[token] = keyAuth
	p.mu.Unlock()

	return nil
}

// CleanUp removes the challenge response
func (p *HTTPProvider) CleanUp(domain, token, keyAuth string) error {
	p.mu.Lock()
	delete(p.tokens, token)
	p.mu.Unlock()

	return nil
}

// ServeHTTP handles ACME HTTP-01 challenge requests
func (p *HTTPProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract token from path: /.well-known/acme-challenge/{token}
	token := r.URL.Path[len("/.well-known/acme-challenge/"):]

	l.Log.Info().
		Str("token", token).
		Msg("HTTP challenge request")

	p.mu.RLock()
	keyAuth, ok := p.tokens[token]
	p.mu.RUnlock()

	if !ok {
		l.Log.Warn().
			Str("token", token).
			Msg("HTTP challenge token not found")
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(keyAuth))
}
