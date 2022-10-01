package wallet

import (
	"testing"

	"github.com/anzx/fabric-cards/pkg/integration/selfservice"

	"github.com/anzx/fabric-cards/pkg/integration/auditlogger"

	"github.com/anzx/fabricapis/pkg/fabric/type/audit"
	"github.com/anzx/fabricapis/pkg/fabric/type/audit/servicedata"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/anzx/fabric-cards/pkg/integration/eligibility"
	"github.com/anzx/fabric-cards/pkg/integration/entitlements"

	cpb "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"

	"github.com/anzx/fabric-cards/test/data"
	"github.com/anzx/fabric-cards/test/fixtures"
	"github.com/stretchr/testify/assert"
)

const (
	nonce          = "ZWZhZHNmZjIzNDEyMzQxMjQzMTIzcg=="
	nonceSignature = "WldaaFpITm1aakl6TkRFeU16UXhNalF6TVRJemNnPT0="
	certificate    = "MIICYDCCAgigAwIBAgIBATAJBgcqhkjOPQQBME0xCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANWSUMxDDAKBgNVBAcMA01FTDEVMBMGA1UECgwMQU5aLCBMaW1pdGVkMQswCQYDVQQDDAJDQTAgFw0xNzA1MzAwMDU4MjBaGA8yMDY3MDUxODAwNTgyMFowTzELMAkGA1UEBhMCQVUxDDAKBgNVBAgMA1ZJQzEMMAoGA1UEBwwDTUVMMRUwEwYDVQQKDAxBTlosIExpbWl0ZWQxDTALBgNVBAMMBExFQUYwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATi9v+FoWFJ7kaW7znWFmKannAyGCL99B5Gnsg+RRSho5054bNX9O3VcXe8/NbVgFGoI/0gGGLFitNCP8OFWFlWo4HVMIHSMB0GA1UdDgQWBBTnMcsNS3qeIABiVs71OoxbqopjYTAfBgNVHSMEGDAWgBSP/I+jB807+ZXcfzac/Kogwft1kjAJBgNVHRMEAjAAMAsGA1UdDwQEAwIFoDBKBgNVHREEQzBBggtleGFtcGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tghBtYWlsLmV4YW1wbGUuY29tgg9mdHAuZXhhbXBsZS5jb20wLAYJYIZIAYb4QgENBB8WHU9wZW5TU0wgR2VuZXJhdGVkIENlcnRpZmljYXRlMAkGByqGSM49BAEDRwAwRAIhAJvuc9LobSfBKan/CzBjAGIuPVhQKb95Q/6twil72gU5Ah9USslVcYrIJDTJoN8h0chdJNuj9hpztSl9JNu+GlbQ, MIICqDCCAk+gAwIBAgIBATAJBgcqhkjOPQQBMIGSMQswCQYDVQQGEwJBVTEMMAoGA1UECAwDVklDMQwwCgYDVQQHDANNRUwxFTATBgNVBAoMDEFOWiwgTGltaXRlZDEMMAoGA1UECwwDQ0FNMSEwHwYDVQQDDBh3d3cucm9vdC1jZXJ0aWZpY2F0ZS5jb20xHzAdBgkqhkiG9w0BCQEWEHRlc3RAZXhhbXBsZS5jb20wIBcNMTcwNTMwMDA1NjQzWhgPMjA2NzA1MTgwMDU2NDNaME0xCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANWSUMxDDAKBgNVBAcMA01FTDEVMBMGA1UECgwMQU5aLCBMaW1pdGVkMQswCQYDVQQDDAJDQTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABMbd8c5xiBWoBPjFMxraRov/u2Ex140NCOTUajXizo+VvW2p+VDz3VMSYggZb8afBHCospeopuK4KUN7m4cZILKjgdgwgdUwHQYDVR0OBBYEFI/8j6MHzTv5ldx/Npz8qiDB+3WSMB8GA1UdIwQYMBaAFL1KnPWJHaHxbcXyGZ4VzR1t60QQMAwGA1UdEwQFMAMBAf8wCwYDVR0PBAQDAgKkMEoGA1UdEQRDMEGCC2V4YW1wbGUuY29tgg93d3cuZXhhbXBsZS5jb22CEG1haWwuZXhhbXBsZS5jb22CD2Z0cC5leGFtcGxlLmNvbTAsBglghkgBhvhCAQ0EHxYdT3BlblNTTCBHZW5lcmF0ZWQgQ2VydGlmaWNhdGUwCQYHKoZIzj0EAQNIADBFAiAvEmjURPLZJIog8OzklpuUzJ3pQZpXAraApe0rehEM1gIhAMh1/dDD4e+cmF25wGkYOmdmnHOPT4SCWpvaHVo/nWCg"
)

func TestNewService(t *testing.T) {
	c := fixtures.AServer().WithData(data.AUserWithACard())
	eligibility := &eligibility.Client{
		CardEligibilityAPIClient: c.CardEligibilityAPIClient,
	}
	entitlements := &entitlements.Client{
		CardEntitlementsAPIClient: c.CardEntitlementsAPIClient,
	}
	auditlog := &auditlogger.Client{
		Publisher: c.AuditLogPublisher,
	}
	selfService := &selfservice.Client{
		PartyAPIClient: c.SelfServiceClient,
	}
	got := NewServer(c.CTMClient, c.VaultClient, c.APCAMClient, eligibility, entitlements, auditlog, c.GPayClient, selfService)
	assert.NotNil(t, got)
	assert.IsType(t, &server{}, got)
}

func buildCardServer(c *fixtures.ServerBuilder) cpb.WalletAPIServer {
	eligibility := &eligibility.Client{
		CardEligibilityAPIClient: c.CardEligibilityAPIClient,
	}
	entitlements := &entitlements.Client{
		CardEntitlementsAPIClient: c.CardEntitlementsAPIClient,
	}
	auditlog := &auditlogger.Client{
		Publisher: c.AuditLogPublisher,
	}
	selfService := &selfservice.Client{
		PartyAPIClient: c.SelfServiceClient,
	}
	return NewServer(c.CTMClient, c.VaultClient, c.APCAMClient, eligibility, entitlements, auditlog, c.GPayClient, selfService)
}

func TestServer_CreateApplePaymentTokenAuditLog(t *testing.T) {
	t.Run("audit log send expected service data", func(t *testing.T) {
		sd := servicedata.CreatePaymentToken{}

		hook := func(buf []byte) {
			p := &audit.AuditLog{}
			_ = protojson.Unmarshal(buf, p)
			_ = p.GetServiceData()[0].UnmarshalTo(&sd)
		}
		builder := fixtures.AServer().WithData(data.AUserWithACard()).WithAuditLogHook(hook)
		request := &cpb.CreateApplePaymentTokenRequest{
			TokenizedCardNumber: data.AUserWithACard().Token(),
		}
		ctx, _ := fixtures.GetTestContextWithLogger(nil)
		s := buildCardServer(builder)

		_, _ = s.CreateApplePaymentToken(ctx, request)
		require.NoError(t, sd.Validate())
		assert.Equal(t, data.AUserWithACard().Token(), sd.GetTokenizedCardNumber())
		assert.Equal(t, data.AUserWithACard().CardNumber()[12:], sd.GetLast_4Digits())
		assert.Equal(t, apple, sd.GetProvider())
	})
}
