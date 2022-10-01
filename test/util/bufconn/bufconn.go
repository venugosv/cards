package bufconn

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/anzx/fabric-cards/pkg/servers"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type lis struct {
	*bufconn.Listener
}

func (l lis) BufDialer(context.Context, string) (net.Conn, error) {
	return l.Dial()
}

func GetClientConn(t *testing.T, registration servers.GRPCRegistration) (*grpc.ClientConn, error) {
	l := GetListener(registration)

	cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(l.BufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return cc, err
}

func GetListener(registration servers.GRPCRegistration) lis {
	l := lis{Listener: bufconn.Listen(bufSize)}
	s := grpc.NewServer()
	registration(s)
	go func() {
		_ = s.Serve(l)
		defer s.Stop()
	}()
	return l
}
