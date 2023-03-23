package scan

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type PortState int

const (
	Open   PortState = 0
	Closed PortState = 1
)

type Protocol string

const (
	Tcp Protocol = "tcp"
	Udp Protocol = "udp"
)

type Port struct {
	Number   int
	Service  string
	Protocol Protocol
	State    PortState
}

func isHttp(conn net.Conn) bool {
	client := http.Client{Transport: &http.Transport{Dial: func(network, addr string) (net.Conn, error) { return conn, nil }}}
	req, _ := http.NewRequest("GET", "http://"+conn.LocalAddr().String(), nil)
	_, err := client.Do(req)
	return err == nil
}

func scanPort(protocol string, host string, port int) Port {
	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.DialTimeout(protocol, address, 60*time.Second)
	if err != nil {
		return Port{
			Number:   port,
			Protocol: Protocol(protocol),
			State:    Closed,
		}
	}
	defer conn.Close()

	if isHttp(conn) {
		return Port{
			Number:   port,
			Service:  "http",
			Protocol: Protocol(protocol),
			State:    Open,
		}
	}

	return Port{
		Number:   port,
		Protocol: Protocol(protocol),
		State:    Open,
	}
}

func mergeScannerChannels(cs ...<-chan Port) <-chan Port {
	channel := make(chan Port)
	var wg sync.WaitGroup

	wg.Add(len(cs))
	handler := func(c <-chan Port) {
		defer wg.Done()

		for n := range c {
			channel <- n
		}
	}

	for _, c := range cs {
		go handler(c)
	}

	go func() {
		wg.Wait()
		close(channel)
	}()

	return channel
}

func CreateScannerPool(n int, protocol string, host string, start int, end int) <-chan Port {
	portChannel := make(chan int, n)
	go func() {
		defer close(portChannel)
		for port := start; port <= end; port++ {
			portChannel <- port
		}
	}()

	var scannerChannels []<-chan Port
	for i := 0; i < n; i++ {
		scannerChannel := make(chan Port)
		scannerChannels = append(scannerChannels, scannerChannel)

		go func() {
			defer close(scannerChannel)
			for port := range portChannel {
				scannerChannel <- scanPort(protocol, host, port)
			}
		}()
	}

	return mergeScannerChannels(scannerChannels...)
}
