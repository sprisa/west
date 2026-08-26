package west

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPClientForAddressPreservesHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Host)
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse("http://west.invalid/api")
	if err != nil {
		t.Fatal(err)
	}
	client := httpClientForAddress(endpoint, serverURL.Host)
	response, err := client.Get(endpoint.String())
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != endpoint.Host {
		t.Fatalf("request host = %q, want %q", got, endpoint.Host)
	}
}

func TestHTTPClientForAddressSkipsRefusedConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse("http://west.invalid/api")
	if err != nil {
		t.Fatal(err)
	}

	badClient := httpClientForAddress(endpoint, "127.0.0.1:1")
	_, err = badClient.Get(endpoint.String())
	if err == nil {
		t.Fatal("expected refused connection to fail")
	}

	goodClient := httpClientForAddress(endpoint, serverURL.Host)
	resp, err := goodClient.Get(endpoint.String())
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
