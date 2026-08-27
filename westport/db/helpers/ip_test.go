package helpers

import "testing"

func TestIpCidrScan(t *testing.T) {
	for _, value := range []any{"10.10.10.1/24", []byte("10.10.10.1/24")} {
		var cidr IpCidr
		if err := cidr.Scan(value); err != nil {
			t.Fatalf("Scan(%T): %v", value, err)
		}
		if got := cidr.String(); got != "10.10.10.1/24" {
			t.Fatalf("Scan(%T) = %q", value, got)
		}
	}
}

func TestIpCidrScanRejectsUnsupportedType(t *testing.T) {
	var cidr IpCidr
	if err := cidr.Scan(42); err == nil {
		t.Fatal("expected unsupported scan type to fail")
	}
}
