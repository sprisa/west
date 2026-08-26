package westport

import "testing"

func TestValidateInstallRequest(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
		datastore   string
		ip          string
		wantAdd     bool
		wantError   bool
	}{
		{name: "first install", datastore: "sqlite"},
		{name: "first install cannot add", datastore: "postgres://db/west", ip: "10.10.10.2", wantError: true},
		{name: "existing requires IP", initialized: true, datastore: "postgres://db/west", wantError: true},
		{name: "SQLite cannot add", initialized: true, datastore: "sqlite", ip: "10.10.10.2", wantError: true},
		{name: "Postgres adds", initialized: true, datastore: "postgres://db/west", ip: "10.10.10.2", wantAdd: true},
		{name: "MySQL adds", initialized: true, datastore: "mysql://db/west", ip: "10.10.10.2", wantAdd: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateInstallRequest(test.initialized, test.datastore, test.ip)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError %v", err, test.wantError)
			}
			if got != test.wantAdd {
				t.Fatalf("add = %v, want %v", got, test.wantAdd)
			}
		})
	}
}

func TestResolvePortEndpoint(t *testing.T) {
	for _, endpoint := range []string{"203.0.113.10:4242", "lh2.example.com:4242"} {
		got, err := resolvePortEndpoint(endpoint)
		if err != nil {
			t.Fatalf("resolvePortEndpoint(%q): %v", endpoint, err)
		}
		if got != endpoint {
			t.Fatalf("resolvePortEndpoint(%q) = %q", endpoint, got)
		}
	}
	for _, endpoint := range []string{"example.com", "example.com:0", ":4242"} {
		if _, err := resolvePortEndpoint(endpoint); err == nil {
			t.Fatalf("expected %q to fail", endpoint)
		}
	}
}
