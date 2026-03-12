package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func isIPWithingNet(ipStr string, subnetStr *net.IPNet) (bool, error) {
	clientIP := net.ParseIP(ipStr)
	if clientIP == nil {
		return false, fmt.Errorf("client IP form is incorrect: %s", ipStr)
	}

	return subnetStr.Contains(clientIP), nil
}

func SubnetChecker(subnet *net.IPNet) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Println("subnetchecking is starting")
			if subnet != nil {
				ipStr := r.Header.Get("X-Real-IP")
				if ipStr == "" {
					log.Printf("no X-Real-IP header in request from %s", r.RemoteAddr)
					http.Error(w, "no X-Real-IP header", http.StatusForbidden)
					return
				}
				clientIP := net.ParseIP(ipStr)
				if clientIP == nil {
					log.Printf("incorrect client ip address: %s", ipStr)
					http.Error(w, "ip is incorrect", http.StatusBadRequest)
					return
				}
				if !subnet.Contains(clientIP) {
					log.Printf("ip %s is not within the trusted network %s", ipStr, subnet)
					http.Error(w, "ip is untrusted", http.StatusForbidden)
					return
				}
			}
			h(w, r)
		}
	}
}
