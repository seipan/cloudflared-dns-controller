package cloudflare

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

type Client interface {
	ListDNSRecords(ctx context.Context) ([]DNSRecord, error)
	CreateDNSRecord(ctx context.Context, record DNSRecord) error
	DeleteDNSRecord(ctx context.Context, recordID string) error
	IsTunnelRecord(rec DNSRecord, tunnelID string) bool
}

type DNSRecord struct {
	ID      string
	Name    string // hostname (e.g., "hoge.example.com")
	Type    string // "CNAME"
	Content string // target (e.g., "<tunnel-id>.cfargotunnel.com")
	Proxied bool
	TTL     int
}

type client struct {
	cf     *cloudflare.Client
	zoneID string
}

func NewClient(token, zoneID string) Client {
	cfClient := cloudflare.NewClient(
		option.WithAPIToken(token),
	)
	return &client{
		cf:     cfClient,
		zoneID: zoneID,
	}
}

func (c *client) ListDNSRecords(ctx context.Context) ([]DNSRecord, error) {
	page, err := c.cf.DNS.Records.List(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(c.zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	records := make([]DNSRecord, 0, len(page.Result))
	for _, r := range page.Result {
		records = append(records, DNSRecord{
			ID:      r.ID,
			Name:    r.Name,
			Type:    string(r.Type),
			Content: r.Content,
			Proxied: r.Proxied,
			TTL:     int(r.TTL),
		})
	}

	return records, nil
}

func (c *client) CreateDNSRecord(ctx context.Context, record DNSRecord) error {
	_, err := c.cf.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(c.zoneID),
		Body: dns.CNAMERecordParam{
			Name:    cloudflare.F(record.Name),
			Content: cloudflare.F(record.Content),
			Type:    cloudflare.F(dns.CNAMERecordTypeCNAME),
			Proxied: cloudflare.F(record.Proxied),
			TTL:     cloudflare.F(dns.TTL(record.TTL)),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create DNS record %s: %w", record.Name, err)
	}
	return nil
}

func (c *client) DeleteDNSRecord(ctx context.Context, recordID string) error {
	_, err := c.cf.DNS.Records.Delete(ctx, recordID, dns.RecordDeleteParams{
		ZoneID: cloudflare.F(c.zoneID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete DNS record %s: %w", recordID, err)
	}
	return nil
}

func (r *client) IsTunnelRecord(rec DNSRecord, tunnelID string) bool {
	return rec.Type == "CNAME" &&
		rec.Content == tunnelID+".cfargotunnel.com"
}
