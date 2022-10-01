package certvalidator

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"strings"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	anzerrors "github.com/anzx/pkg/errors"
	anzcodes "github.com/anzx/pkg/errors/errcodes"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type Config struct {
	Root         string `json:"root,omitempty"`
	Intermediate string `json:"intermediate,omitempty"`
}

const (
	xClientCertificate             = "x-client-certificate"
	xClientCertificateCommonName   = "x-client-certificate-common-name"
	xClientCertificateFingerprint  = "x-client-certificate-fingerprint"
	xClientCertificateSerialNumber = "x-client-certificate-serial-number"
)

func IncomingHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case xClientCertificate,
		xClientCertificateCommonName,
		xClientCertificateFingerprint,
		xClientCertificateSerialNumber:
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func UnaryServerInterceptor(ctx context.Context, cfg *Config) (grpc.UnaryServerInterceptor, error) {
	certificatesRoot, err := loadPublicCert(ctx, cfg.Root)
	if err != nil {
		logf.Error(ctx, err, "unable to read expected pem block")
		return nil, err
	}

	rootPool := x509.NewCertPool()
	rootPool.AddCert(certificatesRoot)

	certificatesIntermediate, err := loadPublicCert(ctx, cfg.Intermediate)
	if err != nil {
		logf.Error(ctx, err, "unable to read expected pem block")
		return nil, err
	}

	intermediatePool := x509.NewCertPool()
	intermediatePool.AddCert(certificatesIntermediate)

	verifyOpts := x509.VerifyOptions{
		Intermediates: intermediatePool,
		Roots:         rootPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logf.Debug(ctx, "unable to fetch metadata from incoming context")
			return nil, anzerrors.New(codes.Unauthenticated, "unable to validate client certificate",
				anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to fetch metadata from incoming context"))
		}

		certificateString := md.Get(xClientCertificate)
		if len(certificateString) < 1 {
			logf.Debug(ctx, "no certificate in header")
			return nil, anzerrors.New(codes.Unauthenticated, "unable to validate client certificate",
				anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "no certificate in header"))
		}

		incomingCert, err := loadPublicCert(ctx, certificateString[0])
		if err != nil {
			logf.Error(ctx, err, "unable to read incoming pem block")
			return nil, anzerrors.New(codes.Unauthenticated, "unable to validate client certificate",
				anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to read incoming pem block"))
		}

		if _, err = incomingCert.Verify(verifyOpts); err != nil {
			logf.Debug(ctx, "unable to verified incoming certificate")
			return nil, anzerrors.New(codes.Unauthenticated, "unable to validate client certificate",
				anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to verify incoming certificate"))
		}

		logf.Debug(ctx, "certificate verified")

		return handler(ctx, req)
	}, nil
}

func loadPublicCert(ctx context.Context, in string) (*x509.Certificate, error) {
	certBytes, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		logf.Error(ctx, err, "unable to decode base64 certificate")
		return nil, anzerrors.Wrap(err, codes.Unauthenticated, "unable to validate client certificate",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to decode base64 cert certificate"))
	}

	out, err := x509.ParseCertificate(certBytes)
	if err != nil {
		logf.Debug(ctx, "unable to read cert block")
		return nil, anzerrors.Wrap(err, codes.Unauthenticated, "unable to validate client certificate",
			anzerrors.NewErrorInfo(ctx, anzcodes.ValidationFailure, "unable to parse cert block"))
	}

	return out, nil
}
