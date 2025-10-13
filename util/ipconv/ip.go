package ipconv

// TODO: Switch to using netip
import "net"

type IP uint32

func (ip IP) ToIPV4() net.IP {
  return IntToIPv4(uint32(ip))
}

func (ip IP) ToInt() uint32 {
  return uint32(ip)
}

func ParseToIP(ipString string) (IP, error) {
  val, err := IPv4ToInt(net.ParseIP(ipString))
  return IP(val), err
}
