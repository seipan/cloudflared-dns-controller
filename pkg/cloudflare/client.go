package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

type Client interface {
	ListDNSRecords(ctx context.Context, zoneID string) ([]DNSRecord, error)
	CreateDNSRecord(ctx context.Context, zoneID string, record DNSRecord) error
	DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error
}

type DNSRecord struct {
	ID      string
	Name    string // hostname (e.g., "hoge.exmple.com")
	Type    string // "CNAME"
	Content string // target (e.g., "<tunnel-id>.cfargotunnel.com")
	Proxied bool
	TTL     int
}

type client struct {
	cf *cloudflare.Client

	zoneID string
}

func NewClient(token string) *client {
	cfClient := cloudflare.NewClient(
		option.WithAPIToken(token),
	)
	return &client{
		cf: cfClient,
	}
}
