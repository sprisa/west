package helpers

import (
	"database/sql/driver"
	"fmt"
	"net/netip"
)

func NewIpCidr(cidr string) (IpCidr, error) {
	prefix, err := netip.ParsePrefix(cidr)
	return IpCidr{prefix}, err
}

type IpCidr struct {
	netip.Prefix
}

// Reads
func (s *IpCidr) Scan(value any) error {
	if value == nil {
		return nil
	}

	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for EncryptedString: %T", value)
	}

	prefix, err := netip.ParsePrefix(v)
	if err != nil {
		return err
	}
	s.Prefix = prefix
	return nil
}

// Writes
func (s IpCidr) Value() (driver.Value, error) {
	return s.String(), nil
}
