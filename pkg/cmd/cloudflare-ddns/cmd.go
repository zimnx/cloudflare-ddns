package cloudflareddns

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/cobra"
	"github.com/zimnx/cloudflare-ddns/pkg/addressprovider"
	"github.com/zimnx/cloudflare-ddns/pkg/ipify"
)

type CloudflareDDNSOptions struct {
	CloudflareAPIToken string
	ZoneID             string
	ARecordID          string
	AAAARecordID       string
	RecordName         string

	RefreshInterval           time.Duration
	PublicAddressProviderType addressprovider.PublicAddressProviderType

	httpClient       *http.Client
	cloudflareClient *cloudflare.API
	addressProviders map[addressprovider.PublicAddressProviderType]addressprovider.PublicAddressProviderFactory
	addressProvider  addressprovider.PublicAddressProvider
}

func NewCloudflareDDNSOptions() *CloudflareDDNSOptions {
	return &CloudflareDDNSOptions{
		RefreshInterval:           5 * time.Minute,
		PublicAddressProviderType: ipify.IpifyAddressProvider,

		addressProviders: map[addressprovider.PublicAddressProviderType]addressprovider.PublicAddressProviderFactory{
			ipify.IpifyAddressProvider: ipify.New,
		},
	}
}

func NewCloudflareDDNSCommand() *cobra.Command {
	o := NewCloudflareDDNSOptions()

	cmd := &cobra.Command{
		Use:   "cloudflare-ddns",
		Short: "Run the cloudflare-ddns",
		Long:  `Run the cloudflare-ddns`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Complete()
			if err != nil {
				return err
			}

			err = o.Validate()
			if err != nil {
				return err
			}

			err = o.Run()
			if err != nil {
				return err
			}

			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,
	}

	o.AddFlags(cmd)

	return cmd
}

func (o *CloudflareDDNSOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.CloudflareAPIToken, "cloudflare-api-token", "", o.CloudflareAPIToken, "Cloudflare API Token")
	cmd.Flags().StringVarP(&o.ZoneID, "zone-id", "", o.ZoneID, "Cloudflare DNS zone ID")
	cmd.Flags().StringVarP(&o.ARecordID, "a-record-id", "", o.ARecordID, "Cloudflare A record ID")
	cmd.Flags().StringVarP(&o.AAAARecordID, "aaaa-record-id", "", o.AAAARecordID, "Cloudflare AAAA record ID")
	cmd.Flags().StringVarP(&o.RecordName, "record-name", "", o.RecordName, "Cloudflare record name ID")

	cmd.Flags().DurationVarP(&o.RefreshInterval, "refresh-interval", "", o.RefreshInterval, "Record refresh interval")
}

func (o *CloudflareDDNSOptions) Validate() error {
	var errs []error

	if len(o.CloudflareAPIToken) == 0 {
		errs = append(errs, errors.New("cloudflare-api-token can't be empty"))
	}
	if len(o.ZoneID) == 0 {
		errs = append(errs, errors.New("zone-id can't be empty"))
	}
	if len(o.ARecordID) == 0 {
		errs = append(errs, errors.New("a-record-id can't be empty"))
	}
	if len(o.AAAARecordID) == 0 {
		errs = append(errs, errors.New("aaaa-record-id can't be empty"))
	}
	if len(o.RecordName) == 0 {
		errs = append(errs, errors.New("record-name can't be empty"))
	}

	return errors.Join(errs...)
}

func (o *CloudflareDDNSOptions) Complete() error {
	api, err := cloudflare.NewWithAPIToken(o.CloudflareAPIToken)
	if err != nil {
		return fmt.Errorf("can't initalize cloudflare client: %w", err)
	}

	o.cloudflareClient = api
	o.httpClient = http.DefaultClient

	o.addressProvider = o.addressProviders[o.PublicAddressProviderType](o.httpClient)

	return nil
}

func (o *CloudflareDDNSOptions) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	return o.Execute(ctx)
}

func (o *CloudflareDDNSOptions) Execute(ctx context.Context) error {
	t := time.NewTicker(o.RefreshInterval)
	defer t.Stop()

	var ipv4AddressCache string
	var ipv6AddressCache string

	for {
		select {
		case <-ctx.Done():
			slog.Info("Finished refreshing DNS records, closing...")
			return nil
		case <-t.C:
			var err error
			ipv4AddressCache, err = o.refreshDNS(ctx, ipv4AddressCache, "A", o.ARecordID, o.addressProvider.GetPublicIPV4)
			if err != nil {
				slog.Warn("Failed to refresh DNS A record", "error", err)
			}
			ipv6AddressCache, err = o.refreshDNS(ctx, ipv6AddressCache, "AAAA", o.AAAARecordID, o.addressProvider.GetPublicIPV6)
			if err != nil {
				slog.Warn("Failed to refresh DNS AAAA record", "error", err)
			}
		}
	}
}

type getPublicAddressFunc func(context.Context) (string, error)

func (o *CloudflareDDNSOptions) refreshDNS(ctx context.Context, lastAddress, recordType, recordID string, addressFunc getPublicAddressFunc) (string, error) {
	address, err := addressFunc(ctx)
	if err != nil {
		return "", fmt.Errorf("can't get address: %w", err)
	}
	if lastAddress != address {
		slog.Info("Refreshing DNS record", "type", recordType, "address", address)
		err = o.cloudflareClient.UpdateDNSRecord(ctx, o.ZoneID, recordID, cloudflare.DNSRecord{
			Type:    recordType,
			Name:    o.RecordName,
			Content: address,
			TTL:     3360,
		})
		if err != nil {
			return "", fmt.Errorf("update A record: %w", err)
		}
	}

	return address, nil
}
