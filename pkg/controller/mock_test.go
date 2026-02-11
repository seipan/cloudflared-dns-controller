package controller

import (
	"context"

	"github.com/seipan/cloudflared-dns-controller/pkg/cloudflare"
)

type fakeCloudflareClient struct {
	records []cloudflare.DNSRecord

	createdRecords []cloudflare.DNSRecord
	deletedIDs     []string

	listErr   error
	createErr error
	deleteErr error
}

func (f *fakeCloudflareClient) ListDNSRecords(_ context.Context) ([]cloudflare.DNSRecord, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.records, nil
}

func (f *fakeCloudflareClient) CreateDNSRecord(_ context.Context, record cloudflare.DNSRecord) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdRecords = append(f.createdRecords, record)
	f.records = append(f.records, record)
	return nil
}

func (f *fakeCloudflareClient) DeleteDNSRecord(_ context.Context, recordID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deletedIDs = append(f.deletedIDs, recordID)
	filtered := f.records[:0]
	for _, rec := range f.records {
		if rec.ID != recordID {
			filtered = append(filtered, rec)
		}
	}
	f.records = filtered
	return nil
}

func (f *fakeCloudflareClient) IsTunnelRecord(rec cloudflare.DNSRecord, tunnelID string) bool {
	return rec.Type == "CNAME" && rec.Content == tunnelID+".cfargotunnel.com"
}
