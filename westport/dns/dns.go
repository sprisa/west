package dns

import (
	"context"
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/sprisa/west/util/info"
	"github.com/sprisa/west/westport/acme"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/ent/device"
	"github.com/sprisa/x/errutil"
	l "github.com/sprisa/x/log"
)

var publicIp net.IP

func StartCompassDNSServer(
	ctx context.Context,
	addr string,
	client *ent.Client,
	settings *ent.Settings,
	acme *acme.DNSProvider,
) error {
	if settings.DomainZone == "" {
		l.Log.Warn().Msg("No domain zone configured. Compass DNS server disabled.")
		return nil
	}

	ip, err := info.GetPublicIP()
	if err != nil {
		l.Log.Err(err).Msg("error finding public ip")
	} else {
		l.Log.Info().Msgf("Public IP: %s", ip.String())
		publicIp = ip
	}

	dnsServer := &dns.Server{Addr: addr, Net: "udp"}
	dns.HandleFunc(".", func(res dns.ResponseWriter, msg *dns.Msg) {
		handleDnsRequest(ctx, res, msg, client, settings, acme)
	})

	var closeError error
	go func() {
		<-ctx.Done()
		closeError = dnsServer.Shutdown()
		l.Log.Err(closeError).Msg("Compass DNS shutdown")
	}()
	l.Log.Info().Str("addr", dnsServer.Addr).Msg("Starting Compass DNS Server")
	err = dnsServer.ListenAndServe()
	if err != nil {
		return errutil.WrapErr(err, "failed to start Compass DNS Server")
	}
	return closeError
}

func handleDnsRequest(
	ctx context.Context,
	res dns.ResponseWriter,
	msg *dns.Msg,
	client *ent.Client,
	settings *ent.Settings,
	acme *acme.DNSProvider,
) {
	m := new(dns.Msg)
	m.SetReply(msg)
	m.Compress = false

	switch msg.Opcode {
	case dns.OpcodeQuery:
		parseQuery(ctx, m, client, settings, acme)
	}

	res.WriteMsg(m)
}

func parseQuery(
	ctx context.Context,
	msg *dns.Msg,
	client *ent.Client,
	settings *ent.Settings,
	acme *acme.DNSProvider,
) {
	for _, q := range msg.Question {
		qName := strings.ToLower(q.Name)
		// Cut off the trailing dot
		host, _ := strings.CutSuffix(qName, ".")
		isDomainZone := strings.HasSuffix(host, settings.DomainZone)
		// l.Log.Info().
		// 	Str("host", host).
		// 	Bool("isDomainZone", isDomainZone).
		// 	Msg("dns")
		// Skip if for external domain
		if !isDomainZone {
			return
		}
		switch q.Qtype {
		case dns.TypeA:
			// Handle API Record
			if host == settings.DomainZone {
				rr := &dns.A{
					A: publicIp,
					Hdr: dns.RR_Header{
						Name:   qName,
						Rrtype: dns.TypeA,
						// "IN" stands for internet. Standard class.
						Class: dns.ClassINET,
						// 5min
						Ttl: 300,
					},
				}
				msg.Answer = append(msg.Answer, rr)
			} else {
				dvcName, _ := strings.CutSuffix(host, "."+settings.DomainZone)
				dvc, err := client.Device.Query().
					Select(device.FieldIP).
					Where(device.Name(dvcName)).
					First(ctx)
				if err != nil {
					l.Log.Err(err).
						Str("dvc", dvcName).
						Msg("dns error fetching device")

					// NXDOMAIN
					msg.Rcode = dns.RcodeNameError
					return
				}
				l.Log.Info().
					Str("host", host).
					Str("ip", dvc.IP.ToIpAddr().String()).
					Msg("DNS query")

				rr := &dns.A{
					A: dvc.IP.ToIPV4(),
					Hdr: dns.RR_Header{
						Name:   qName,
						Rrtype: dns.TypeA,
						// "IN" stands for internet. Standard class.
						Class: dns.ClassINET,
						// 5min
						Ttl: 300,
					},
				}
				msg.Answer = append(msg.Answer, rr)
			}

		case dns.TypeTXT:
			// l.Log.Info().
			// 	Str("q", q.String()).
			// 	Msg("DNS TXT query")
			// Handle ACME DNS-01 challenges
			if acme != nil {
				if value, ok := acme.GetTXTRecord(qName); ok {
					l.Log.Info().
						Str("host", host).
						Str("value", value).
						Msg("DNS TXT query (ACME)")

					rr := &dns.TXT{
						Txt: []string{value},
						Hdr: dns.RR_Header{
							Name:   qName,
							Rrtype: dns.TypeTXT,
							Class:  dns.ClassINET,
							Ttl:    60, // Short TTL for challenges
						},
					}
					msg.Answer = append(msg.Answer, rr)
					// addAuthorityRecords(msg, settings, name)
					return
				}
			}
		}
	}
}
