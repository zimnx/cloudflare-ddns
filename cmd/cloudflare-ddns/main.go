package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
)

const (
	cloudflareAPIToken = "Dh8neVjrZudkxhV0apBPDuh4fuXXGYTUlPIVKsyc"
	zoneID             = "a3f2243d67e039da9b18c84ff5dd8a94"
	aRecordID          = "e022c3fc61f9b221e656952865f97c0e"
	aaaaRecordID       = "938ed0dbc7d1f9e8267ac59870644b61"
	recordName         = "nas"

	ipv4Ipfy = "https://api.ipify.org"
	ipv6Ipfy = "https://api64.ipify.org"

	refreshInterval = 5 * time.Minute
)

var (
	lastIpv4 = ""
	lastIpv6 = ""
)

func queryIpify(ctx context.Context, client *http.Client, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("make request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GET: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return string(body), nil
}

func getPublicIPV4(ctx context.Context, client *http.Client) (string, error) {
	return queryIpify(ctx, client, ipv4Ipfy)
}

func getPublicIPV6(ctx context.Context, client *http.Client) (string, error) {
	return queryIpify(ctx, client, ipv6Ipfy)
}

func refreshDNS(ctx context.Context, client *http.Client, api *cloudflare.API) error {
	ipv4, err := getPublicIPV4(ctx, client)
	if err != nil {
		return fmt.Errorf("get public ipv4: %w", err)
	}

	ipv6, err := getPublicIPV6(ctx, client)
	if err != nil {
		return fmt.Errorf("get public ipv6: %w", err)
	}

	if lastIpv4 != ipv4 {
		fmt.Println("Updating A with", ipv4)

		err = api.UpdateDNSRecord(ctx, zoneID, aRecordID, cloudflare.DNSRecord{
			Type:    "A",
			Name:    recordName,
			Content: ipv4,
			TTL:     3360,
		})
		if err != nil {
			return fmt.Errorf("update A record: %w", err)
		}
		lastIpv4 = ipv4
	}

	if lastIpv6 != ipv6 {
		fmt.Println("Updating AAAA with", ipv6)

		err = api.UpdateDNSRecord(ctx, zoneID, aaaaRecordID, cloudflare.DNSRecord{
			Type:    "AAAA",
			Name:    recordName,
			Content: ipv6,
			TTL:     3360,
		})
		if err != nil {
			return fmt.Errorf("update AAAA record: %w", err)
		}
		lastIpv6 = ipv6
	}

	return nil
}

func realMain(ctx context.Context) error {
	client := http.DefaultClient

	api, err := cloudflare.NewWithAPIToken(cloudflareAPIToken)
	if err != nil {
		return fmt.Errorf("cloudflare client init: %w", err)
	}

	if err := refreshDNS(ctx, client, api); err != nil {
		fmt.Println("Error during DNS refresh:", err)
	}

	t := time.NewTicker(refreshInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done")
			return nil
		case <-t.C:
			if err := refreshDNS(ctx, client, api); err != nil {
				fmt.Println("Error during DNS refresh:", err)
			}
		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := realMain(ctx); err != nil {
		panic(err)
	}
}
