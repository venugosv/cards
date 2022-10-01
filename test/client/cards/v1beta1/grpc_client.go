package v1beta1

import (
	"context"
	"fmt"
	"testing"

	"github.com/anzx/fabric-cards/test/client/common"

	"github.com/anzx/fabric-cards/pkg/integration/vault"
	cpbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1"
	epbv1beta1 "github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1"
	"google.golang.org/grpc"
)

type GRPCTestClient struct {
	cardAPIClient            cpbv1beta1.CardAPIClient
	cardEligibilityAPIClient epbv1beta1.CardEligibilityAPIClient
	walletAPIClient          cpbv1beta1.WalletAPIClient
	ctx                      context.Context
	state                    common.ConnectionState
}

func NewGRPCClient(ctx context.Context, cc *grpc.ClientConn, vault vault.Client) *GRPCTestClient {
	return &GRPCTestClient{
		ctx:                      ctx,
		cardAPIClient:            cpbv1beta1.NewCardAPIClient(cc),
		cardEligibilityAPIClient: epbv1beta1.NewCardEligibilityAPIClient(cc),
		walletAPIClient:          cpbv1beta1.NewWalletAPIClient(cc),
		state: common.ConnectionState{
			Vault: vault,
		},
	}
}

func (c *GRPCTestClient) GetCurrentCard() *cpbv1beta1.Card {
	return c.state.CurrentCard
}

func (c *GRPCTestClient) SetTokenizedCardNumber(tokenizedCardNumber string) {
	if c.state.CurrentCard == nil {
		c.state.CurrentCard = &cpbv1beta1.Card{}
	}
	c.state.CurrentCard.TokenizedCardNumber = tokenizedCardNumber
}

func (c *GRPCTestClient) LoadCard(t *testing.T) {
	listResponse, err := c.List()
	if err != nil {
		t.Fatalf("LoadCard: failed to ListCards: %v", err)
	}

	c.state.GetCard(listResponse.GetCards(), c.state.CurrentCard.GetTokenizedCardNumber())
}

func (c *GRPCTestClient) List() (*cpbv1beta1.ListResponse, error) {
	return c.cardAPIClient.List(c.ctx, &cpbv1beta1.ListRequest{})
}

func (c *GRPCTestClient) Can(eligibility epbv1beta1.Eligibility) bool {
	return c.state.Can(eligibility)
}

func (c *GRPCTestClient) GetDetails() (*cpbv1beta1.GetDetailsResponse, error) {
	return c.cardAPIClient.GetDetails(c.ctx, &cpbv1beta1.GetDetailsRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
	})
}

func (c *GRPCTestClient) Activate() (*cpbv1beta1.ActivateResponse, error) {
	plaintext := c.state.CurrentCard.GetTokenizedCardNumber()
	if c.state.Vault != nil {
		var err error
		plaintext, err = c.state.Vault.DecodeCardNumber(c.ctx, c.state.CurrentCard.TokenizedCardNumber)
		if err != nil {
			return nil, fmt.Errorf("unable to getRequest last 6 digits %v", err)
		}
	}
	return c.cardAPIClient.Activate(c.ctx, &cpbv1beta1.ActivateRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
		Last_6Digits:        plaintext[len(plaintext)-6:],
	})
}

func (c *GRPCTestClient) GetWrappingKey() (*cpbv1beta1.GetWrappingKeyResponse, error) {
	return c.cardAPIClient.GetWrappingKey(c.ctx, &cpbv1beta1.GetWrappingKeyRequest{})
}

func (c *GRPCTestClient) ResetPIN() (*cpbv1beta1.ResetPINResponse, error) {
	return c.cardAPIClient.ResetPIN(c.ctx, &cpbv1beta1.ResetPINRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
		EncryptedPinBlock:   pinBlock1,
	})
}

func (c *GRPCTestClient) SetPIN() (*cpbv1beta1.SetPINResponse, error) {
	return c.cardAPIClient.SetPIN(c.ctx, &cpbv1beta1.SetPINRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
		EncryptedPinBlock:   pinBlock1,
	})
}

func (c *GRPCTestClient) Replace(reason cpbv1beta1.ReplaceRequest_Reason) (*cpbv1beta1.ReplaceResponse, error) {
	return c.cardAPIClient.Replace(c.ctx, &cpbv1beta1.ReplaceRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
		Reason:              reason,
	})
}

func (c *GRPCTestClient) AuditTrail() (*cpbv1beta1.AuditTrailResponse, error) {
	return c.cardAPIClient.AuditTrail(c.ctx, &cpbv1beta1.AuditTrailRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
	})
}

func (c *GRPCTestClient) CreateApplePaymentToken() (*cpbv1beta1.CreateApplePaymentTokenResponse, error) {
	var (
		nonce          = "ZWZhZHNmZjIzNDEyMzQxMjQzMTIzcg=="
		nonceSignature = "WldaaFpITm1aakl6TkRFeU16UXhNalF6TVRJemNnPT0="
		certificate    = []string{
			"MIICYDCCAgigAwIBAgIBATAJBgcqhkjOPQQBME0xCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANWSUMxDDAKBgNVBAcMA01FTDEVMBMGA1UECgwMQU5aLCBMaW1pdGVkMQswCQYDVQQDDAJDQTAgFw0xNzA1MzAwMDU4MjBaGA8yMDY3MDUxODAwNTgyMFowTzELMAkGA1UEBhMCQVUxDDAKBgNVBAgMA1ZJQzEMMAoGA1UEBwwDTUVMMRUwEwYDVQQKDAxBTlosIExpbWl0ZWQxDTALBgNVBAMMBExFQUYwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAATi9v+FoWFJ7kaW7znWFmKannAyGCL99B5Gnsg+RRSho5054bNX9O3VcXe8/NbVgFGoI/0gGGLFitNCP8OFWFlWo4HVMIHSMB0GA1UdDgQWBBTnMcsNS3qeIABiVs71OoxbqopjYTAfBgNVHSMEGDAWgBSP/I+jB807+ZXcfzac/Kogwft1kjAJBgNVHRMEAjAAMAsGA1UdDwQEAwIFoDBKBgNVHREEQzBBggtleGFtcGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tghBtYWlsLmV4YW1wbGUuY29tgg9mdHAuZXhhbXBsZS5jb20wLAYJYIZIAYb4QgENBB8WHU9wZW5TU0wgR2VuZXJhdGVkIENlcnRpZmljYXRlMAkGByqGSM49BAEDRwAwRAIhAJvuc9LobSfBKan/CzBjAGIuPVhQKb95Q/6twil72gU5Ah9USslVcYrIJDTJoN8h0chdJNuj9hpztSl9JNu+GlbQ",
			"MIICqDCCAk+gAwIBAgIBATAJBgcqhkjOPQQBMIGSMQswCQYDVQQGEwJBVTEMMAoGA1UECAwDVklDMQwwCgYDVQQHDANNRUwxFTATBgNVBAoMDEFOWiwgTGltaXRlZDEMMAoGA1UECwwDQ0FNMSEwHwYDVQQDDBh3d3cucm9vdC1jZXJ0aWZpY2F0ZS5jb20xHzAdBgkqhkiG9w0BCQEWEHRlc3RAZXhhbXBsZS5jb20wIBcNMTcwNTMwMDA1NjQzWhgPMjA2NzA1MTgwMDU2NDNaME0xCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANWSUMxDDAKBgNVBAcMA01FTDEVMBMGA1UECgwMQU5aLCBMaW1pdGVkMQswCQYDVQQDDAJDQTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABMbd8c5xiBWoBPjFMxraRov/u2Ex140NCOTUajXizo+VvW2p+VDz3VMSYggZb8afBHCospeopuK4KUN7m4cZILKjgdgwgdUwHQYDVR0OBBYEFI/8j6MHzTv5ldx/Npz8qiDB+3WSMB8GA1UdIwQYMBaAFL1KnPWJHaHxbcXyGZ4VzR1t60QQMAwGA1UdEwQFMAMBAf8wCwYDVR0PBAQDAgKkMEoGA1UdEQRDMEGCC2V4YW1wbGUuY29tgg93d3cuZXhhbXBsZS5jb22CEG1haWwuZXhhbXBsZS5jb22CD2Z0cC5leGFtcGxlLmNvbTAsBglghkgBhvhCAQ0EHxYdT3BlblNTTCBHZW5lcmF0ZWQgQ2VydGlmaWNhdGUwCQYHKoZIzj0EAQNIADBFAiAvEmjURPLZJIog8OzklpuUzJ3pQZpXAraApe0rehEM1gIhAMh1/dDD4e+cmF25wGkYOmdmnHOPT4SCWpvaHVo/nWCg",
		}
	)
	return c.walletAPIClient.CreateApplePaymentToken(c.ctx, &cpbv1beta1.CreateApplePaymentTokenRequest{
		TokenizedCardNumber: c.state.CurrentCard.TokenizedCardNumber,
		Nonce:               nonce,
		NonceSignature:      nonceSignature,
		Certificates:        certificate,
	})
}

func (c *GRPCTestClient) CreateGooglePaymentToken() (*cpbv1beta1.CreateGooglePaymentTokenResponse, error) {
	var (
		stableHardwareId = "IGoXgEyHZIK9g1EcQT62LkWr"
		activeWalletId   = "b710ZoYlVe9A-VcLjK7X_UmR"
	)
	return c.walletAPIClient.CreateGooglePaymentToken(c.ctx, &cpbv1beta1.CreateGooglePaymentTokenRequest{
		TokenizedCardNumber: c.state.CurrentCard.GetTokenizedCardNumber(),
		CardNetwork:         cpbv1beta1.CardNetwork_CARD_NETWORK_VISA,
		StableHardwareId:    stableHardwareId,
		ActiveWalletId:      activeWalletId,
	})
}
