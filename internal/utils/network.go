package utils

import (
	"log"
	"net"

	defaultroute "github.com/nixigaj/go-default-route"
)

func GetDefaultIface() (string, error) {
	details, err := defaultroute.DefaultRoute()
	if err != nil {
		return "", err
	}
	return details.InterfaceName, nil
}

func GetLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("cannot get interface info: %v\nn", err)
	}
	for _, iface := range interfaces {
		defaultIface, err := GetDefaultIface()
		if err != nil {
			log.Printf("cannot get default gw info: %v\n", err)
		}
		if iface.Name == defaultIface {
			addrs, err := iface.Addrs()
			if err != nil {
				log.Printf("cannot get addresses info: %v\n ", err)
			}
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if ok {
					if !ipNet.IP.IsLoopback() {
						return ipNet.IP.String()
					}
				}
			}
		}
	}
	log.Printf("set 127.0.0.1\n")
	return "127.0.0.1"
}
