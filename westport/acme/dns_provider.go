package acme

import (
	"sync"

	"github.com/go-acme/lego/v4/challenge/dns01"
	l "github.com/sprisa/x/log"
)

// DNSProvider implements the challenge.Provider interface for DNS-01 challenges
type DNSProvider struct {
	mu      sync.RWMutex
	records map[string]string // maps FQDN to TXT record value
}

func NewDNSProvider() *DNSProvider {
	return &DNSProvider{
		records: make(map[string]string),
	}
}

// Present creates a TXT record for the ACME challenge
func (p *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)
	l.Log.Info().
		Str("domain", domain).
		Str("fqdn", info.FQDN).
		Str("value", info.Value).
		Msgf("dns challenge")

	p.mu.Lock()
	p.records[info.FQDN] = info.Value
	p.mu.Unlock()

	return nil
}

// CleanUp removes the TXT record after challenge completion
func (p *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	p.mu.Lock()
	delete(p.records, fqdn)
	p.mu.Unlock()

	return nil
}

// GetTXTRecord retrieves a TXT record for ACME challenges
func (p *DNSProvider) GetTXTRecord(fqdn string) (string, bool) {
	p.mu.RLock()
	value, ok := p.records[fqdn]
	p.mu.RUnlock()
	return value, ok
}
