package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"syscall"

	"github.com/c-robinson/iplib"
	"tailscale.com/net/socks5"
)

func main() {
	if err := loadConfig(); err != nil {
		panic(err)
	}
	go checkIPsLoop()

	subnet := iplib.NewNet6(net.ParseIP(cfg.Subnet), cfg.SubnetMask, 0)

	server := &socks5.Server{
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Split host and port
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("failed to split host and port: %w", err)
			}

			ip, err := net.ResolveIPAddr("ip", host)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve IP: %w", err)
			}

			_, ok := whitelist[ip.String()]
			if !ok {
				for _, whost := range cfg.Whitelist {
					if strings.EqualFold(host, whost) {
						ok = true
						break
					}
				}
			}

			// Valid dest
			if ok {
				newAddr := subnet.RandomIP()
				log.Println("Dialing", network, addr, "from", newAddr)

				dialer := &net.Dialer{
					Control: func(network, address string, c syscall.RawConn) error {
						var operr error
						if err := c.Control(func(fd uintptr) {
							operr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_FREEBIND, 1)
						}); err != nil {
							return err
						}
						return operr
					},
					LocalAddr: &net.TCPAddr{
						IP: newAddr,
					},
				}

				conn, err := dialer.DialContext(ctx, network, addr)
				if err != nil {
					log.Println("Failed to dial:", err)
				}
				return conn, err
			}

			return nil, fmt.Errorf("ip %s is not in the whitelist", ip.IP.String())
		},
	}

	ln, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(ln); err != nil {
		panic(err)
	}
}
