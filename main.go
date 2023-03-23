package main

import (
	"flag"
	"fmt"

	"github.com/flohansen/port-scanner/scan"
)

func main() {
	host := flag.String("host", "localhost", "The target host")
	start := flag.Int("start", 1, "The start port")
	end := flag.Int("end", 1000, "The end port")
	flag.Parse()

	tcpScannerPool := scan.CreateScannerPool(16, "tcp", *host, *start, *end)

	fmt.Printf("Target: %s:(%d-%d)\n", *host, *start, *end)
	fmt.Printf("Open ports:\n")
	for port := range tcpScannerPool {
		if port.State == scan.Open {
			fmt.Printf("  - %d/%s\t%s\n", port.Number, port.Protocol, port.Service)
		}
	}
}
