package west

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/sprisa/west/west/gql"
)

func provisionWithFailover(ctx context.Context, endpoint *url.URL, addresses []string, input gql.ProvisionDeviceInput) (*gql.ProvisionDeviceResponse, error) {
	perAttempt := 10 * time.Second
	var errs error

	for _, addr := range addresses {
		attemptCtx, cancel := context.WithTimeout(ctx, perAttempt)
		httpClient := httpClientForAddress(endpoint, addr)
		client := graphql.NewClient(endpoint.String(), httpClient)
		data, err := gql.ProvisionDevice(attemptCtx, client, input)
		cancel()
		if err == nil {
			return data, nil
		}
		errs = errors.Join(errs, err)
	}

	attemptCtx, cancel := context.WithTimeout(ctx, perAttempt)
	defer cancel()
	client := graphql.NewClient(endpoint.String(), &http.Client{Timeout: perAttempt})
	data, err := gql.ProvisionDevice(attemptCtx, client, input)
	if err == nil {
		return data, nil
	}
	return nil, errors.Join(errs, err)
}

func httpClientForAddress(endpoint *url.URL, address string) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, address)
	}
	if endpoint.Scheme == "https" {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.ServerName = endpoint.Hostname()
	}
	return &http.Client{Transport: transport, Timeout: 10 * time.Second}
}
