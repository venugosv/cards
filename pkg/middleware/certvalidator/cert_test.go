package certvalidator

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/anzx/pkg/errors"
	"google.golang.org/grpc/codes"

	"golang.org/x/sync/errgroup"

	"github.com/anzx/fabric-cards/pkg/servers"
	ecpb "github.com/anzx/fabricapis/pkg/visa/service/enrollmentcallback"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/anzx/pkg/monitoring/names"
)

var (
	once  sync.Once
	certs testCerts
)

type testCerts struct {
	bytes  []byte
	base64 string
	cert   *x509.Certificate
}

func getCerts(t *testing.T) testCerts {
	once.Do(func() {
		bytes := encodePublicCert(t, generatePrivateKey(t))
		base64 := base64.StdEncoding.EncodeToString(bytes)

		cert, err := x509.ParseCertificate(bytes)
		if err != nil {
			t.Fatalf("unable to parse cert block %v", err)
		}

		certs = testCerts{
			bytes:  bytes,
			base64: base64,
			cert:   cert,
		}
	})
	return certs
}

func TestCertMatcher(t *testing.T) {
	tests := map[string]bool{
		xClientCertificate:                   true,
		xClientCertificateCommonName:         true,
		xClientCertificateFingerprint:        true,
		xClientCertificateSerialNumber:       true,
		"X-CLIENT-CERTIFICATE":               true,
		"X-CLIENT-CERTIFICATE-COMMON-NAME":   true,
		"X-CLIENT-CERTIFICATE-FINGERPRINT":   true,
		"X-CLIENT-CERTIFICATE-SERIAL-NUMBER": true,
	}
	for key, val := range tests {
		got, ok := IncomingHeaderMatcher(key)
		assert.Equal(t, val, ok)
		assert.Equal(t, key, got)
	}
}

func TestCertValidator_interceptor(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
		wantErr  string
	}{
		{
			name: "success",
			metadata: map[string]string{
				xClientCertificate: getCerts(t).base64,
			},
		},
		{
			name:    "unable to fetch metadata from incoming context",
			wantErr: "fabric error: status_code=Unauthenticated, error_code=4, message=unable to validate client certificate, reason=unable to fetch metadata from incoming context",
		},
		{
			name:     "no certificate in header",
			metadata: map[string]string{},
			wantErr:  "fabric error: status_code=Unauthenticated, error_code=4, message=unable to validate client certificate, reason=no certificate in header",
		},
		{
			name: "unable to read incoming pem block",
			metadata: map[string]string{
				xClientCertificate: "qwerty",
			},
			wantErr: "fabric error: status_code=Unauthenticated, error_code=4, message=unable to validate client certificate, reason=unable to read incoming pem block",
		},
		{
			name: "unable to verify incoming certificate",
			metadata: map[string]string{
				xClientCertificate: base64.StdEncoding.EncodeToString(encodePublicCert(t, generatePrivateKey(t))),
			},
			wantErr: "fabric error: status_code=Unauthenticated, error_code=4, message=unable to validate client certificate, reason=unable to verify incoming certificate",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, metadata.New(test.metadata))
			}

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			}

			config := &Config{
				Root:         getCerts(t).base64,
				Intermediate: getCerts(t).base64,
			}
			interceptor, _ := UnaryServerInterceptor(context.Background(), config)
			got, err := interceptor(ctx, nil, nil, handler)
			if test.wantErr != "" {
				assert.EqualError(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, got)
			}
		})
	}
}

func TestCertValidator(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		secretKey string
		redirect  bool
		wantErr   string
	}{
		{
			name: "success",
			config: &Config{
				Root:         getCerts(t).base64,
				Intermediate: getCerts(t).base64,
			},
		}, {
			name: "unable to read pem block",
			config: &Config{
				Root: getCerts(t).base64,
			},
			wantErr: "fabric error: status_code=Unauthenticated, error_code=4, message=unable to validate client certificate, reason=unable to parse cert block",
		}, {
			name:    "unable to read pem block",
			config:  &Config{},
			wantErr: "fabric error: status_code=Unauthenticated, error_code=4, message=unable to validate client certificate, reason=unable to parse cert block",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			got, err := UnaryServerInterceptor(ctx, test.config)
			if test.wantErr != "" {
				assert.EqualError(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

type testConfig struct {
	AppName      string
	Port         int
	Certificates *Config
}

func TestCertValidator_InterceptorMatcher(t *testing.T) {
	if !testing.Short() {
		t.Skip()
	}
	cfg := testConfig{
		AppName: "test-app",
		Port:    6789,
		Certificates: &Config{
			Root:         getCerts(t).base64,
			Intermediate: getCerts(t).base64,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(setupAPIServer(gCtx, cfg))
	g.Go(servers.SignalListener(gCtx))

	host := fmt.Sprintf(":%v", cfg.Port)

	// prepare grpc client
	cc, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	e := ecpb.NewEnrollmentCallbackAPIClient(cc)

	t.Run("grpc unable to validate client certificate", func(t *testing.T) {
		_, err = e.Enroll(context.Background(), &ecpb.Request{})
		ferr, ok := errors.FromStatusError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, ferr.GetStatusCode())
	})
	t.Run("grpc successfully validated client certificate", func(t *testing.T) {
		md := metadata.New(map[string]string{
			xClientCertificate: getCerts(t).base64,
		})
		ctx = metadata.NewOutgoingContext(context.Background(), md)
		_, err = e.Enroll(ctx, &ecpb.Request{})
		ferr, ok := errors.FromStatusError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unimplemented, ferr.GetStatusCode())
	})

	// prepare rest client
	restTarget := fmt.Sprintf("http://%s/webhook/Visa/AccountServices/v3/Enrollment/Notification", host)

	t.Run("rest unable to validate client certificate", func(t *testing.T) {
		httpRequest, err := http.NewRequest(http.MethodPost, restTarget, nil)
		require.NoError(t, err)
		httpClient := http.DefaultClient

		response, err := httpClient.Do(httpRequest)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	})
	t.Run("rest successfully validated client certificate", func(t *testing.T) {
		httpRequest, err := http.NewRequest(http.MethodPost, restTarget, nil)
		require.NoError(t, err)
		httpRequest.Header.Set(xClientCertificate, getCerts(t).base64)
		httpClient := http.DefaultClient

		response, err := httpClient.Do(httpRequest)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotImplemented, response.StatusCode)
	})
}

func setupAPIServer(ctx context.Context, cfg testConfig) func() error {
	return func() error {
		certValidator, err := UnaryServerInterceptor(ctx, cfg.Certificates)
		if err != nil {
			return err
		}

		interceptors := []grpc.UnaryServerInterceptor{
			certValidator,
		}

		grpcRegistrations := []servers.GRPCRegistration{
			func(server *grpc.Server) {
				ecpb.RegisterEnrollmentCallbackAPIServer(server, &ecpb.UnimplementedEnrollmentCallbackAPIServer{})
			},
		}

		grpcServer := servers.GRPCServer(grpcRegistrations, interceptors)

		serveMux := runtime.NewServeMux(
			runtime.WithIncomingHeaderMatcher(IncomingHeaderMatcher),
		)

		restRegistrations := []servers.RestRegistration{
			ecpb.RegisterEnrollmentCallbackAPIHandlerFromEndpoint,
		}

		restServer := servers.CreateRestServer(ctx, cfg.Port, serveMux, names.Unknown, restRegistrations...)

		return servers.Serve(ctx, cfg.AppName, cfg.Port, grpcServer, restServer)
	}
}

// encodePublicCert take a rsa.PublicKey and return a x509 cert bytes
func encodePublicCert(t *testing.T, privateKey *rsa.PrivateKey) []byte {
	serial := big.NewInt(2019)
	ca := genCert(serial)

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Log("failed to create x509 cert")
		t.Fail()
	}

	// Private key in PEM format
	return caBytes
}

func genCert(serial *big.Int) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

// generatePrivateKey creates a RSA Private Key of specified byte size.
func generatePrivateKey(t *testing.T) *rsa.PrivateKey {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Log("failed to create private key")
		t.Fail()
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		t.Log("failed to validate private key")
		t.Fail()
	}

	t.Log("Private Key generated")

	return privateKey
}
