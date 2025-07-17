package ipify

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/zimnx/cloudflare-ddns/pkg/addressprovider"
)

type Ipify struct {
	httpClient *http.Client
}

const (
	IpifyAddressProvider addressprovider.PublicAddressProviderType = "ipify"

	ipv4IpifyAPIURL = "https://api.ipify.org"
	ipv6IpifyAPIURL = "https://api6.ipify.org"
)

func New(httpClient *http.Client) addressprovider.PublicAddressProvider {
	return &Ipify{
		httpClient: httpClient,
	}
}

func (ip *Ipify) GetPublicIPV4(ctx context.Context) (string, error) {
	return ip.queryIpify(ctx, ipv4IpifyAPIURL)
}

func (ip *Ipify) GetPublicIPV6(ctx context.Context) (string, error) {
	return ip.queryIpify(ctx, ipv6IpifyAPIURL)
}

func (ip *Ipify) queryIpify(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("can't make HTTP request: %w", err)
	}
	resp, err := ip.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("can't send GET request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("can't read response body: %w", err)
	}

	return string(body), nil
}
