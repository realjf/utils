package utils

import (
	"fmt"
	"net"
	"os"
)

func IsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

type IPAddress struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	Name string `json:"name"`
}

func GetIPAddress() (ips []IPAddress, err error) {
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, iface := range interfaces {
		// 只处理启用的接口
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取接口的地址
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		ipAddr := IPAddress{
			Name: iface.Name,
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip.IsLoopback() {
				continue // 忽略回环地址
			}

			if ip.To4() != nil {
				// IPv4 地址
				fmt.Printf("IPv4 地址: %s (网卡: %s)\n", ip.String(), iface.Name)
				ipAddr.IPv4 = ip.String()
			} else if ip.To16() != nil {
				// IPv6 地址
				fmt.Printf("IPv6 地址: %s (网卡: %s)\n", ip.String(), iface.Name)
				ipAddr.IPv6 = ip.String()
			} else {
				fmt.Println("Error:  Unknown ip address")
			}
		}
		if ipAddr.IPv4 != "" || ipAddr.IPv6 != "" {
			ips = append(ips, ipAddr)
		}
	}

	return
}

func IsIPv6Address(ip string) (is bool) {
	addr := net.ParseIP(ip)
	if addr == nil {
		fmt.Printf("%s 是无效的 IP 地址\n", ip)
		return
	}

	if addr.To4() != nil {
		fmt.Printf("%s 是 IPv4 地址\n", ip)
	} else if addr.To16() != nil {
		fmt.Printf("%s 是 IPv6 地址\n", ip)
		return true
	} else {
		fmt.Printf("%s 既不是 IPv4 也不是 IPv6 地址\n", ip)
	}
	return
}
