package config

// Defines the path of each file required for a Nebula host: CA certificate, host certificate, and host key.
// Each of these files can also be stored inline as YAML multiline strings.
//
// See: https://nebula.defined.net/docs/config/pki/
type Pki struct {
	// The ca is a collection of one or more certificate authorities this host should trust.
	Ca string `yaml:"ca"`
	// The cert is a certificate unique to every host on a Nebula network. The certificate identifies a host’s IP address, name, and group membership within a Nebula network.
	// The certificate is signed by a certificate authority when created, which informs other hosts on whether to trust a particular host certificate.
	Cert string `yaml:"cert"`
	// The key is a private key unique to every host on a Nebula network.
	// It is used in conjunction with the host certificate to prove a host’s identity to other members of the Nebula network.
	// The private key should never be shared with other hosts.
	Key string `yaml:"key"`
	// The blocklist contains a list of individual hosts' certificate fingerprints which should be blocked even
	// if the certificate is otherwise valid (signed by a trusted CA and unexpired.).
	// This should be used if a host's credentials are stolen or compromised.
	//
	// NOTE: The blocklist is not distributed via Lighthouses. To ensure access to your entire network is blocked you must distribute the
	// full blocklist to every host in your network. This is typically done via tooling such as Ansible, Chef, or Puppet.
	Blocklist []string `yaml:"blocklist,omitempty"`
}

// The static host map defines a set of hosts with fixed IP addresses on the internet (or any network).
// A host can have multiple fixed IP addresses defined here, and nebula will try each when establishing a tunnel.
// The syntax is: "<nebula ip>": ["<routable ip/dns name>:<routable port>"]
//
// See: https://nebula.defined.net/docs/config/static-host-map/
type StaticHostMap = map[string][]string

// The static map defines options which control the interpretation of the static_host_map.
type StaticMap struct {
	// Select the IP version used to communicate with hosts in the static map.
	// Valid values are ip4, ip6, and ip (for both.)
	Network StaticMapNetwork `yaml:"network,omitempty"`
	// Interval to re-query DNS for hostnames listed in the static map
	// Default: 30s
	Cadence string `yaml:"cadence,omitempty"`
	// Timeout for DNS lookup of static hosts
	// Default: 250ms
	LookupTimeout string `yaml:"lookup_timeout,omitempty"`
}

type StaticMapNetwork string

const (
	StaticMapNetworkIPv4      StaticMapNetwork = "ip4"
	StaticMapNetworkIPv6      StaticMapNetwork = "ip6"
	StaticMapNetworkIPv4AndV6 StaticMapNetwork = "ip"
)

type LightHouseDns struct {
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
}

// See: https://nebula.defined.net/docs/config/lighthouse/
type Lighthouse struct {
	// am_lighthouse is used to enable lighthouse functionality for a node.
	// This should ONLY be true on nodes you have configured to be lighthouses in your network.
	Am_lighthouse bool `yaml:"am_lighthouse"`
	// serve_dns optionally starts a DNS listener that responds to A and TXT queries and can even be delegated to for name resolution by external DNS hosts.
	// The DNS listener can only respond to requests about hosts it's aware of. For this reason, it can only be enabled on Lighthouses.
	// A records contain the Nebula IP for a host name and can be queried by any host that can reach the DNS listener, regardless of whether it is communicating over the Nebula network.
	// TXT records can only be queried over the Nebula network, and contain certificate information for the requested host IP address.
	Serve_dns bool `yaml:"serve_dns"`
	// dns is used to configure the address (host) and port (port) the DNS server should listen on.
	// By listening on the host's Nebula IP, you can make the DNS server accessible only on the Nebula network.
	// Alternatively, listening on 0.0.0.0 will allow anyone that can reach the host to make queries.
	// The default value for dns.port is 53 but you must set an IP address.
	Dns LightHouseDns `yaml:"dns,omitempty"`
	// interval specifies how often a nebula host should report itself to a lighthouse.
	// By default, hosts report themselves to lighthouses once every 10 seconds.
	// Use caution when changing this interval, as it may affect host discovery times in a large nebula network.
	Interval int `yaml:"interval,omitempty"`
	// hosts is a list of lighthouse hosts this node should report to and query from.
	// The lighthouses listed here should be referenced by their nebula IP, not by the IPs of their physical network interfaces.
	// Note: This should be empty on lighthouse nodes.
	Hosts []string `yaml:"hosts,omitempty"`
	// remote_allow_list allows you to control ip ranges that this node will consider when handshaking to another node.
	// By default, any remote IPs are allowed. You can provide CIDRs here with true to allow and false to deny.
	// The most specific CIDR rule applies to each remote. If all rules are "allow", the default will be "deny", and vice-versa.
	// If both "allow" and "deny" rules are present, then you MUST set a rule for "0.0.0.0/0" as the default.
	// Similarly if both "allow" and "deny" IPv6 rules are present, then you MUST set a rule for "::/0" as the default.
	Remote_allow_list map[string]bool `yaml:"remote_allow_list,omitempty"`
	// local_allow_list allows you to filter which local IP addresses we advertise to the lighthouses.
	// This uses the same logic as remote_allow_list, but additionally, you can specify an interfaces map of regular expressions to match against interface names.
	// The regexp must match the entire name. All interface rules must be either true or false (and the default will be the inverse).
	// CIDR rules are matched after interface name rules. Default is all local IP addresses.
	//
	// TODO: Need to allow "interfaces" sub object
	Local_allow_list map[string]bool `yaml:"local_allow_list,omitempty"`
}

// listen sets the UDP port Nebula will use for sending/receiving traffic and for handshakes.
//
// See: https://nebula.defined.net/docs/config/listen/
type Listen struct {
	// host is the ip of the interface to use when binding the listener.
	// The default is 0.0.0.0, which is what most people should use.
	// To listen on both any ipv4 and ipv6 use "::".
	Host string `yaml:"host,omitempty"`
	// port is the UDP port nebula should use on a host.
	// For a lighthouse node, the port should be defined, conventionally to 4242,
	// however using port 0 or leaving port unset will dynamically assign a port and is recommended for roaming nodes.
	// Using 0 on lighthouses and relay hosts will likely lead to connectivity issues.
	Port int `yaml:"port,omitempty"`
	// Sets the max number of packets to pull from the kernel for each syscall (under systems that support recvmmsg).
	Batch        int `yaml:"batch,omitempty"`
	Read_buffer  int `yaml:"read_buffer,omitempty"`
	Write_buffer int `yaml:"write_buffer,omitempty"`
	// The so_sock option is a Linux-specific feature that allows all outgoing Nebula packets to be tagged with a specific identifier.
	// This tagging enables IP rule-based filtering. For example, it supports 0.0.0.0/0 unsafe_routes,
	// allowing for more precise routing decisions based on the packet tags. Default is 0 meaning no mark is set.
	// This setting is reloadable.
	So_Mark int `yaml:"so_mark,omitempty"`
}

// punchy configures the sending of inbound/outbound packets at a
// regular interval to avoid expiration of firewall nat mappings.
//
// See: https://nebula.defined.net/docs/config/punchy/
type Punchy struct {
	// When enabled, Nebula will periodically send "empty" packets
	// to the underlay IP addresses of hosts it has established tunnels
	// to in order to maintain the "hole" punched in the NAT's firewall.
	Punch bool `yaml:"punch"`
	// When enabled, the node will attempt a handshake to
	// the initiating peer in response to the Lighthouse's notification
	// of the peer attempting to handshake with it.
	// This can be useful when a node is behind a difficult NAT
	// for which regular hole punching does not work.
	// Some combinations of NAT still will not work and relays can be used for this scenario.
	Respond bool `yaml:"respond"`
	// Delay is the period of time Nebula waits between receiving
	// a Lighthouse handshake notification and sending an empty packet
	// in order to try to punch a hole in the NAT firewall.
	// This is helpful in some NAT race condition situations.
	Delay string `yaml:"delay,omitempty"`
}

// cipher allows you to choose between the available ciphers for your network.
// You may choose chachapoly to use the ChaCha20-Poly1305 cipher or aes for AES256-GCM.
// Note: This value must be identical on ALL nodes and lighthouses. Nebula does not support the use of different ciphers simultaneously!
//
// See: https://nebula.defined.net/docs/config/cipher/
type Cipher string

const (
	CipherChaChaPoly Cipher = "chachapoly"
	CipherAes        Cipher = "aes"
)

// preferred_ranges sets the priority order for underlay IP addresses.
// Two hosts on the same LAN would likely benefit from having their tunnels use the LAN IP addresses rather than the
// public IP addresses the lighthouse may have learned for them.
// preferred_ranges accepts a list of CIDR ranges admitted as a set of preferred ranges of IP addresses.
//
// See: https://nebula.defined.net/docs/config/preferred-ranges/
type Preferred_ranges = []string

// See: https://nebula.defined.net/docs/config/relay/
type Relay struct {
	// relays are a list of Nebula IPs that peers can use to relay packets to this host.
	// IPs in this list must have am_relay set to true in their configs, otherwise they will reject relay requests.
	Relays []string `yaml:"relays,omitempty"`
	// Set am_relay to true to permit other hosts to list my IP in their relays config. The default is false.
	Am_relay bool `yaml:"am_relay"`
	// Set use_relays to false to prevent this instance from attempting to establish connections through relays. The default is true.
	Use_relays bool `yaml:"use_relays,omitempty"`
}

type TunRoute struct {
	Mtu   int    `yaml:"mtu"`
	Route string `yaml:"route"`
}

type TunUnsafeRoute struct {
	Route  string `yaml:"route"`
	Via    string `yaml:"via"`
	Mtu    int    `yaml:"mtu,omitempty"`
	Metric int    `yaml:"metric,omitempty"`
}

// See: https://nebula.defined.net/docs/config/tun/
type Tun struct {
	// Allows the nebula interface (tun) to be disabled, which lets you run a lighthouse
	// without a nebula interface (and therefore without root).
	// You will not be able to communiate over IP with a nebula node that uses this setting.
	Disabled bool `yaml:"disabled"`
	// dev sets the interface name for your nebula interface. If not set, a default will be chosen by the OS.
	// For macOS: Not required. If set, must be in the form utun[0-9]+. For FreeBSD: Required to be set, must be in the form tun[0-9]+.
	Dev string `yaml:"dev,omitempty"`
	// Toggles forwarding of local broadcast packets, the address of which depends on the ip/mask encoded in pki.cert
	Drop_local_broadcast bool `yaml:"drop_local_broadcast"`
	// Toggles forwarding of multicast packets
	Drop_multicast bool `yaml:"drop_multicast"`
	// Sets the transmit queue length, if you notice lots of transmit drops on the tun it may help to raise this number. Default is 500.
	Tx_queue int `yaml:"tx_queue,omitempty"`
	// Default MTU for every packet, safe setting is (and the default) 1300 for internet routed packets.
	Mtu int `yaml:"mtu,omitempty"`
	// Route based MTU overrides. If you have known VPN IP paths that can support larger MTUs you can increase/decrease them here.
	Routes []TunRoute `yaml:"routes,omitempty"`
	// Unsafe routes allows you to route traffic over nebula to non-nebula nodes.
	// Unsafe routes should be avoided unless you have hosts/services that cannot run nebula.
	Unsafe_routes []TunUnsafeRoute `yaml:"unsafe_routes,omitempty"`
}

type NebulaLogLevel = string

const (
	NebulaLogLevelDebug NebulaLogLevel = "debug"
	NebulaLogLevelInfo  NebulaLogLevel = "info"
	NebulaLogLevelWarn  NebulaLogLevel = "warning"
	NebulaLogLevelError NebulaLogLevel = "error"
	NebulaLogLevelFatal NebulaLogLevel = "fatal"
	NebulaLogLevelPanic NebulaLogLevel = "panic"
)

type NebulaLogFormat = string

const (
	NebulaLogFormatText NebulaLogFormat = "text"
	NebulaLogFormatJson NebulaLogFormat = "json"
)

// See: https://nebula.defined.net/docs/config/logging/
type Logging struct {
	// Controls the verbosity of logs. The options are panic, fatal, error, warning, info, or debug.
	Level NebulaLogLevel `yaml:"level,omitempty"`
	// Controls the logging format. The options are json or text
	Format NebulaLogFormat `yaml:"format,omitempty"`
	// Disables timestamp logging. Useful when output is redirected to logging system that already adds timestamps.
	Disable_timestamp bool `yaml:"disable_timestamp"`
	// timestamp_format is specified in Go time format, see: https://golang.org/pkg/time/#pkg-constants.
	Timestamp_format string `yaml:"timestamp_format,omitempty"`
}

type Proto string

const (
	ProtoAny  Proto = "any"
	ProtoTcp  Proto = "tcp"
	ProtoUdp  Proto = "udp"
	ProtoIcmp Proto = "icmp"
)

type FirewallRule struct {
	// Takes 0 or any as any, a single number (e.g. 80), a range (e.g. 200-901),
	// or fragment to match second and further fragments of fragmented packets (since there is no port available).
	Port int `yaml:"port"`
	// One of any, tcp, udp, or icmp
	Proto Proto `yaml:"proto"`
	// An issuing CA name
	Ca_name string `yaml:"ca_name,omitempty"`
	// An issuing CA shasum
	Ca_sha string `yaml:"ca_sha,omitempty"`
	// Can be any or a literal hostname, ie test-host
	Host string `yaml:"host,omitempty"`
	// Can be any or a literal group name, ie default-group
	Group string `yaml:"group,omitempty"`
	// Same as group but accepts a list of values.
	// Multiple values are AND'd together and a certificate must contain all groups to pass.
	Groups []string `yaml:"groups,omitempty"`
	// a CIDR, 0.0.0.0/0 is any.
	Cidr string `yaml:"cidr,omitempty"`
}

type FirewallConntrack struct {
	Tcp_timeout     string `yaml:"tcp_timeout,omitempty"`
	Udp_timeout     string `yaml:"udp_timeout,omitempty"`
	Default_timeout string `yaml:"default_timeout,omitempty"`
}

// The default state of the Nebula interface host firewall is deny all for all inbound and outbound traffic.
// Firewall rules can be added to allow traffic for specified ports and protocols, but it is not possible to explicitly define a deny rule.
//
// See: https://nebula.defined.net/docs/config/firewall/
type Firewall struct {
	// It is quite common to allow any outbound traffic to flow from a host.
	// This simply means that the host can use any port or protocol to attempt to connect to any other host in the overlay network.
	// (Whether or not those other hosts allow that inbound traffic is up to them.)
	Outbound []FirewallRule `yaml:"outbound,omitempty"`
	// At a minimum, it is recommended to enable ICMP so that ping can be used to verify connectivity.
	// Additionally, if enabling the built-in Nebula SSH server, you may wish to grant access over the Nebula network via firewall rules.
	Inbound   []FirewallRule    `yaml:"inbound,omitempty"`
	Conntrack FirewallConntrack `yaml:"conntrack,omitempty"`
}

type Handshakes struct {
	// Handshakes are sent to all known addresses at each interval with a linear backoff,
	// waiting try_interval after the 1st attempt, 2 * try_interval after the 2nd, etc,
	// until the handshake is older than timeout.
	// Default: 100ms
	TryInterval string `yaml:"try_interval,omitempty"`
	// A 100ms interval with the default 10 retries will
	// give a handshake 5.5 seconds to resolve before timing out.
	// Default: 10
	Retries int `yaml:"retries,omitempty"`
	// trigger_buffer is the size of the buffer channel for quickly
	// sending handshakes after receiving the response for lighthouse queries.
	// Default: 64
	TriggerBuffer int `yaml:"trigger_buffer,omitempty"`
}

// Tunnel manager settings
type Tunnels struct {
	// drop_inactive controls whether inactive tunnels are maintained
	// or dropped after the inactive_timeout period has elapsed.
	// This setting is reloadable.
	// Default: false
	DropInactive bool `yaml:"drop_inactive"`
	// inactivity_timeout controls how long a tunnel MUST NOT see any inbound
	// or outbound traffic before being considered inactive and eligible to be dropped.
	// This setting is reloadable
	// Default: 10m
	InactivityTimeout string `yaml:"inactivity_timeout,omitempty"`
}

// This option is only supported on Linux.
// Routines is the number of thread pairs to run that consume from the tun and UDP queues.
// Currently, this defaults to 1 which means we have 1 tun queue reader and 1 UDP queue reader.
// Setting this above 1 will set IFF_MULTI_QUEUE on the tun device and SO_REUSEPORT on the UDP socket to allow multiple queues.
// The maximum recommended setting is half of the available CPU cores.
// It's recommended to set this to a lower value still, to avoid resource starvation.
type Routines uint8

// Config for Nebula
//
// See: https://nebula.defined.net/docs/config/
type Config struct {
	Pki              Pki              `yaml:"pki"`
	StaticHostMap    StaticHostMap    `yaml:"static_host_map,omitempty"`
	StaticMap        StaticMap        `yaml:"static_map,omitempty"`
	Lighthouse       Lighthouse       `yaml:"lighthouse,omitempty"`
	Listen           Listen           `yaml:"listen,omitempty"`
	Punchy           Punchy           `yaml:"punchy,omitempty"`
	Cipher           Cipher           `yaml:"cipher,omitempty"`
	Preferred_ranges Preferred_ranges `yaml:"preferred_ranges,omitempty"`
	Relay            Relay            `yaml:"relay,omitempty"`
	Tun              Tun              `yaml:"tun,omitempty"`
	Logging          Logging          `yaml:"logging,omitempty"`
	Firewall         Firewall         `yaml:"firewall,omitempty"`
	Handshakes       Handshakes       `yaml:"handshakes,omitempty"`
	Tunnels          Tunnels          `yaml:"tunnels,omitempty"`
	Routines         Routines         `yaml:"routines,omitempty"`
	// TODO: Add sshd, stats
}
