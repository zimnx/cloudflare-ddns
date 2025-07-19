# Cloudflare DDNS

Cloudflare DDNS is a simple Dynamic DNS (DDNS) client that automatically updates DNS records (A/AAAA) in your Cloudflare zone based on your server's public IP address. This ensures your domain record always points to the current IP, even if it changes dynamically.

## Features

- Automatic update of A and AAAA records in Cloudflare
- Supports IPv4 and IPv6
- Configurable refresh interval
- Uses Cloudflare API and [ipify](https://www.ipify.org/) to fetch public IP

## Requirements

- Cloudflare API Token with DNS edit permissions

## Installation

### Build locally

```bash
git clone https://github.com/zimnx/cloudflare-ddns.git
cd cloudflare-ddns
go build -o cloudflare-ddns ./cmd/cloudflare-ddns
```

### Run in a container

You can use the provided `Containerfile` (Dockerfile):

```bash
docker build -t cloudflare-ddns -f images/cloudflare-ddns/Containerfile .
docker run --rm cloudflare-ddns [options]
```

## Usage

Run the application with the following flags:

```bash
./cloudflare-ddns \
  --cloudflare-api-token <TOKEN> \
  --zone-id <ZONE_ID> \
  --a-record-id <A_RECORD_ID> \
  --aaaa-record-id <AAAA_RECORD_ID> \
  --record-name <RECORD_NAME> \
  --refresh-interval <INTERVAL>
```

**Example:**

```bash
./cloudflare-ddns \
  --cloudflare-api-token=cf_api_token \
  --zone-id=cf_zone_id \
  --a-record-id=cf_a_record_id \
  --aaaa-record-id=cf_aaaa_record_id \
  --record-name=example.com \
  --refresh-interval=5m
```

## Parameters

- `--cloudflare-api-token` – Cloudflare API Token
- `--zone-id` – Cloudflare DNS zone ID
- `--a-record-id` – Cloudflare A record ID
- `--aaaa-record-id` – Cloudflare AAAA record ID
- `--record-name` – DNS record name (e.g. `example.com`)
- `--refresh-interval` – Refresh interval (e.g. `5m`)

## License

This project is licensed under the MIT License.
