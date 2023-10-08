package grpcx

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc/peer"
)

type Peer struct {
	Host     string
	Port     int
	Protocol string
	Security string
}

func (p Peer) IsSecure() bool {
	return p.Security != ""
}

func (p Peer) Addr() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

func (p Peer) String() string {
	if p.IsSecure() {
		return fmt.Sprintf("%s (with %s) %s", p.Protocol, p.Security, p.Addr())
	}
	return fmt.Sprintf("%s %s", p.Protocol, p.Addr())
}

func extractPeerInfo(ctx context.Context) Peer {
	var rv Peer
	p, ok := peer.FromContext(ctx)
	if ok {
		rv.Protocol = p.Addr.Network()
		if p.AuthInfo != nil {
			rv.Security = p.AuthInfo.AuthType()
		}
		var host string
		var port int
		switch p.Addr.(type) {

		case *net.TCPAddr:
			addr := p.Addr.(*net.TCPAddr)
			host = addr.IP.String()
			port = addr.Port

		case *net.UDPAddr:
			addr := p.Addr.(*net.UDPAddr)
			host = addr.IP.String()
			port = addr.Port

		case *net.IPAddr:
			addr := p.Addr.(*net.IPAddr)
			host = addr.IP.String()
		}

		rv.Host = host
		rv.Port = port
	}
	return rv
}
