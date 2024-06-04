package main

import (
	"log"
	"net"
	"sync"
	"time"
)

var (
	whitelist = map[string]struct{}{}
	wlMutex   = sync.RWMutex{}
)

func checkHostIPs(wg *sync.WaitGroup, host string) {
	defer wg.Done()

	// Resolve the host to get the IPs
	ips, err := net.LookupHost(host)
	if err != nil {
		log.Printf("Failed to resolve host %s: %v", host, err)
		return
	}

	// No IPs found
	if len(ips) == 0 {
		log.Println("No IPs found for host", host)
		return
	}

	// Update the whitelist
	hostWhitelist := make([]net.IP, 0, len(ips))
	for _, ipStr := range ips {
		ip, err := net.ResolveIPAddr("ip", ipStr)
		if err != nil {
			log.Println("Failed to resolve IP:", err)
			continue
		}
		hostWhitelist = append(hostWhitelist, ip.IP)

		wlMutex.Lock()
		whitelist[ip.String()] = struct{}{}
		wlMutex.Unlock()
	}

	log.Println("Resolved", host, "to", hostWhitelist)
}

func checkIPs() {
	wg := &sync.WaitGroup{}
	for _, host := range cfg.Whitelist {
		wg.Add(1)
		go checkHostIPs(wg, host)
	}
	wg.Wait()
}

func checkIPsLoop() {
	checkIPs()

	ticker := time.NewTicker(cfg.UpdateInterval)
	for range ticker.C {
		checkIPs()
	}
}
