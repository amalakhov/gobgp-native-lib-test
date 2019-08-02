package main

import (
	"bgpush/bgp"
	"flag"
	"fmt"
	"os"
)

func configuration(path string) *bgp.Configuration {
	return &bgp.Configuration{
		Neighbor: bgp.Neighbor{
			NeighborAddress: "78.140.132.189",
			LocalAs:         65000,
			PeerAs:          65000,
		},
		RouterId: "188.42.217.230",
		As:       65001,
	}
}

func runParameters() (help bool, apiListen string, configuration string) {
	flag.BoolVar(&help, "help", false, "Shows this help message.")

	flag.StringVar(&configuration, "configuration", "", "Path to configuration file.")
	flag.StringVar(&apiListen, "api-listen", "127.0.0.1:8080", "IP[:Port] to use for REST API (default port is 8080, address is 127.0.0.1).")

	flag.Parse()

	return help, apiListen, configuration
}

func printHelp() {
	fmt.Println("  Usage of bgpush [option...]")
	fmt.Println("")

	fmt.Println("  -help")
	fmt.Println("\t Shows this help message.")

	fmt.Println("  -api-listen string")
	fmt.Println("\t IP[:Port] to use for REST API (default port is 8080, address is 127.0.0.1).")

	fmt.Println("  -configuration string")
	fmt.Println("\t Path to configuration file.")

	os.Exit(1)
}
