package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func getIP(ctx context.Context) (string, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		log.Printf("cannot get client ip\n")
		return "", status.Errorf(codes.Aborted, "cannot get client ip")
	}
	ip, _, err := net.SplitHostPort(peer.Addr.String())
	if err != nil {
		log.Printf("cannot retrive client ip from %s\n", peer.Addr.String())
		return "", status.Errorf(codes.Aborted, "cannot get client ip")
	}
	return ip, nil 
}
