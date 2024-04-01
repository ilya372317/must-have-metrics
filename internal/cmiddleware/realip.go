package cmiddleware

import (
	"fmt"
	"net"

	"github.com/go-resty/resty/v2"
)

func WithRealIP() resty.RequestMiddleware {
	return func(c *resty.Client, r *resty.Request) error {
		ipStr, err := externalIP()
		if err != nil {
			return fmt.Errorf("failed resolve client ip address: %w", err)
		}
		r.Header.Set("X-Real-IP", ipStr)

		return nil
	}
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("are you connected to the network?")
}
