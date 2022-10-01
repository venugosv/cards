package common

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"

	"github.com/anzx/fabric-cards/pkg/middleware/errors"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
)

func GetConnection(t *testing.T, ctx context.Context, target string, isInsecure bool, headers ...string) *grpc.ClientConn {
	t.Helper()

	opts := []grpc.DialOption{grpc.WithBlock(), grpc.WithUnaryInterceptor(errors.UnaryClientErrorLogInterceptor())}
	if isInsecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			t.Fatalf("GetConnection: failed to prepare cert pool %v", err)
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: rootCAs}))) //nolint:gosec
	}

	t.Logf("GetConnection: Insecure %v, targeting %s", isInsecure, target)

	t.Logf("GetConnection: Headers %v", headers)

	ctx = metadata.AppendToOutgoingContext(ctx, headers...)

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		t.Fatal("GetConnection: failed to dial", err)
	}

	return conn
}

func TearDown(t *testing.T) error {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:15020/quitquitquit", nil)
	if err != nil {
		t.Log("TearDown: failed to create http request to istio")
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Log("TearDown: failed to call istio http://localhost:15020/quitquitquit")
		return err
	}

	t.Logf("TearDown: status %d", resp.StatusCode)
	return nil
}
