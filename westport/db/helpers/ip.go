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

	var v string
	switch value := value.(type) {
	case string:
		v = value
	case []byte:
		v = string(value)
	default:
		return fmt.Errorf("unexpected type for IpCidr: %T", value)
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
