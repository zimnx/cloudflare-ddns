package main

import (
	"fmt"
	"os"

	cmd "github.com/zimnx/cloudflare-ddns/pkg/cmd/cloudflare-ddns"
)

func main() {
	command := cmd.NewCloudflareDDNSCommand()
	err := command.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
