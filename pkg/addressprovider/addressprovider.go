package addressprovider

import (
	"context"
	"net/http"
)

type PublicAddressProvider interface {
	GetPublicIPV4(context.Context) (string, error)
	GetPublicIPV6(context.Context) (string, error)
}

type PublicAddressProviderType string

type PublicAddressProviderFactory func(*http.Client) PublicAddressProvider
