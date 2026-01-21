package avtransport

import (
	"fmt"
	"net"
	"renderctl/internal/models"
	"renderctl/logger"
)

func expandCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ip = ip.To4()
	if ip == nil {
		return nil, fmt.Errorf("only IPv4 CIDR supported")
	}

	var ips []string

	// Convert IP to uint32
	start := ipToUint32(ip)
	mask := netmaskToUint32(ipnet.Mask)

	network := start & mask
	broadcast := network | ^mask

	// Skip network and broadcast
	for i := network + 1; i < broadcast; i++ {
		ips = append(ips, uint32ToIP(i))
	}

	return ips, nil
}

func ipToUint32(ip net.IP) uint32 {
	return uint32(ip[0])<<24 |
		uint32(ip[1])<<16 |
		uint32(ip[2])<<8 |
		uint32(ip[3])
}

func uint32ToIP(n uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}

func netmaskToUint32(mask net.IPMask) uint32 {
	return uint32(mask[0])<<24 |
		uint32(mask[1])<<16 |
		uint32(mask[2])<<8 |
		uint32(mask[3])
}

func ScanSubnet(cfg *models.Config) {
	logger.Notify("Running subnet scan")
	ips, err := expandCIDR(cfg.Subnet)
	if err != nil {
		logger.Error("Invalid subnet: %v", err)
	}

	logger.Notify("Scanning subnet %s (%d hosts)", cfg.Subnet, len(ips))

	for _, ip := range ips {
		cfg.TIP = ip

		ok, err := probeAVTransport(cfg)
		if err != nil || !ok {
			continue
		}

		logger.Success("AVTransport found at %s", ip)
	}
	logger.Done("Subnet scan completed")
}
