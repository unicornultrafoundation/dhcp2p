package main

import (
	"fmt"
	"os"

	"github.com/unicornultrafoundation/dhcp2p/cmd/dhcp2p/cmd"
)

// A version string that can be set with
//
//	-ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var Build string

func main() {
	cmd := cmd.RootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
