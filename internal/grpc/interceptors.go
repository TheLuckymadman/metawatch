package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func SubnetChecker(subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		agentIP, err := getIP(ctx)
		if err != nil {
			return nil, err
		}
		fmt.Printf("grpc request from %s\n", agentIP)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("cannot get metadata from the context")
		}
		ipStr := md.Get("x-real-ip")
		if len(ipStr) == 0 {
			log.Printf("no X-Real-IP metadata in request from %s", agentIP)
			return nil, status.Errorf(codes.PermissionDenied, "no X-Real-IP metadata in request")
		}
		ip := ipStr[0]
		log.Printf("X-Real-IP metadata in request from %s is %s", agentIP, ip)
		clientIP := net.ParseIP(ip)
		if clientIP == nil {
			log.Printf("incorrect client ip address: %s", ipStr)
			return nil, status.Errorf(codes.PermissionDenied, "ip is incorrect")
		}
		if !subnet.Contains(clientIP) {
			log.Printf("ip %s is not within the trusted network %s", ipStr, subnet)
			return nil, status.Errorf(codes.PermissionDenied, "ip is untrusted")
		}
		return handler(ctx, req)
	}
}
