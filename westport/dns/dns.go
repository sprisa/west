package dns

import (
	"context"
	"strings"

	"github.com/miekg/dns"
	"github.com/sprisa/west/util/errutil"
	l "github.com/sprisa/west/util/log"
	"github.com/sprisa/west/westport/db/ent"
	"github.com/sprisa/west/westport/db/ent/device"
)

func StartCompassDNSServer(ctx context.Context, client *ent.Client, settings *ent.Settings) error {
	if settings.DomainZone == "" {
		l.Log.Warn().Msg("No domain zone configured. Compass DNS server disabled.")
		return nil
	}

	dnsServer := &dns.Server{Addr: "0.0.0.0:53", Net: "udp"}
	dns.HandleFunc(".", func(res dns.ResponseWriter, msg *dns.Msg) {
		handleDnsRequest(ctx, res, msg, client, settings)
	})

	var closeError error
	go func() {
		<-ctx.Done()
		closeError = dnsServer.Shutdown()
		l.Log.Err(closeError).Msg("Compass DNS shutdown")
	}()
	l.Log.Info().Str("addr", dnsServer.Addr).Msg("Starting Compass DNS Server")
	err := dnsServer.ListenAndServe()
	if err != nil {
		return errutil.WrapError(err, "failed to start Compass DNS Server")
	}
	return closeError
}

func handleDnsRequest(
	ctx context.Context,
	res dns.ResponseWriter,
	msg *dns.Msg,
	client *ent.Client,
	settings *ent.Settings,
) {
	m := new(dns.Msg)
	m.SetReply(msg)
	m.Compress = false

	switch msg.Opcode {
	case dns.OpcodeQuery:
		parseQuery(ctx, m, client, settings)
	}

	res.WriteMsg(m)
}

func parseQuery(
	ctx context.Context,
	msg *dns.Msg,
	client *ent.Client,
	settings *ent.Settings,
) {
	for _, q := range msg.Question {
		// Cut off the trailing dot
		host, _ := strings.CutSuffix(q.Name, ".")
		dvcName, hasDomainZoneSuffix := strings.CutSuffix(host, "."+settings.DomainZone)
		// Skip if for external domain
		if !hasDomainZoneSuffix {
			return
		}
		switch q.Qtype {
		case dns.TypeA:
			dvc, err := client.Device.Query().
				Select(device.FieldIP).
				Where(device.Name(dvcName)).
				First(ctx)
			if err != nil {
				l.Log.Err(err).
					Str("dvc", dvcName).
					Msg("dns error fetching device")
				return
			}
			l.Log.Info().
				Str("host", host).
				Str("ip", dvc.IP.ToIpAddr().String()).
				Msg("DNS query")

			rr := &dns.A{
				A: dvc.IP.ToIPV4(),
				Hdr: dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					// "IN" stands for internet. Standard class.
					Class: dns.ClassINET,
					// 5min
					Ttl: 300,
				},
			}
			msg.Answer = append(msg.Answer, rr)
		}
	}
}
