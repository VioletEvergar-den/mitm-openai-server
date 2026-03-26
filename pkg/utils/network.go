package utils

import (
	"net"
	"strings"
)

// GetLocalIPs 获取本机所有非回环IP地址
// 返回内网IP地址列表
func GetLocalIPs() []string {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
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

			ips = append(ips, ip.String())
		}
	}

	return ips
}

// IsPrivateIP 判断是否为内网IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	parsedIP = parsedIP.To4()
	if parsedIP == nil {
		return false
	}

	if parsedIP[0] == 10 {
		return true
	}

	if parsedIP[0] == 172 && parsedIP[1] >= 16 && parsedIP[1] <= 31 {
		return true
	}

	if parsedIP[0] == 192 && parsedIP[1] == 168 {
		return true
	}

	if parsedIP[0] == 127 {
		return true
	}

	return false
}

// GetPrimaryIP 获取首选IP地址（优先返回非内网IP）
func GetPrimaryIP() string {
	ips := GetLocalIPs()
	if len(ips) == 0 {
		return "127.0.0.1"
	}

	for _, ip := range ips {
		if !IsPrivateIP(ip) {
			return ip
		}
	}

	return ips[0]
}

// GetPrivateIPs 获取所有内网IP
func GetPrivateIPs() []string {
	var privateIPs []string
	ips := GetLocalIPs()

	for _, ip := range ips {
		if IsPrivateIP(ip) && !strings.HasPrefix(ip, "127.") {
			privateIPs = append(privateIPs, ip)
		}
	}

	return privateIPs
}

// GetPublicIPs 获取所有公网IP（非内网IP）
func GetPublicIPs() []string {
	var publicIPs []string
	ips := GetLocalIPs()

	for _, ip := range ips {
		if !IsPrivateIP(ip) {
			publicIPs = append(publicIPs, ip)
		}
	}

	return publicIPs
}
