package config

// Private CIDR ranges
var DefaultPreferredRanges = PreferredRanges{
	"172.16.0.0/12",
	"192.168.0.0/16",
	"10.0.0.0/8",
	"0.0.0.0/8",
	"127.0.0.0/8",
}
