package vault_external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	logf "github.com/anzx/fabric-cards/pkg/middleware/log"

	"github.com/anzx/fabric-cards/pkg/util/apic"
	"github.com/anzx/pkg/errors"
	"github.com/anzx/pkg/errors/errcodes"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iam/v1"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
	"google.golang.org/grpc/codes"
	jwt2 "gopkg.in/square/go-jose.v2/jwt"
)

type Credentials struct {
	ClientEmail string `json:"client_email"`
}

type AuthRequest struct {
	Role string `json:"role"`
	JWT  string `json:"jwt"`
}

const (
	metadataEmailPath = "/computeMetadata/v1/instance/service-accounts/default/email"
)

// login performs a Vault login
func (c *client) login(ctx context.Context) (*SecretAuth, error) {
	email, err := c.getServiceEmail(ctx)
	// err is already a well behaved ANZ error
	if err != nil {
		return nil, err
	}
	logf.Info(ctx, "got default service account email for vault login")

	jwt, err := c.getJwt(ctx, email)
	// err is already a well behaved ANZ error
	if err != nil {
		return nil, err
	}
	logf.Info(ctx, "got signed jwt for vault login")

	authResponse, err := c.api.vaultLogin(ctx, c.config.AuthRole, jwt)
	// err is already a well behaved ANZ error
	if err != nil {
		logf.Error(ctx, err, "failed to login with Vault")
		return nil, err
	}

	if authResponse.Auth == nil {
		return nil, errors.New(
			codes.Internal,
			"vault login failed",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, "login response has no auth data"))
	}

	return authResponse.Auth, nil
}

// getServiceEmail is used to retrieve the service account email for Vault login
func (c *client) getServiceEmail(ctx context.Context) (string, error) {
	// Support a hard-coded override for testing and stubbed environments
	if c.config.OverrideServiceEmail != "" {
		logf.Info(ctx, "using email from config as default")
		return c.config.OverrideServiceEmail, nil
	}

	googleDefaultEmail, hasEmail, err := getGoogleDefaultEmail(ctx)
	if err != nil {
		return "", err
	}

	if hasEmail {
		return googleDefaultEmail, nil
	}

	logf.Info(ctx, "could not get google default email, falling through to default")

	// Report but do not throw error, we have fallbacks
	// Fallthrough to default address
	mail, err := c.defaultEmail(ctx)
	if err != nil {
		// Don't wrap this error, defaultEmail already provides enough info in errors
		return "", err
	}
	return mail, nil
}

func getGoogleDefaultEmail(ctx context.Context) (string, bool, error) {
	logf.Info(ctx, "looking up default credentials from Google")
	defaultCredentials, err := google.FindDefaultCredentials(ctx, iam.CloudPlatformScope)
	if err != nil {
		// We have fallbacks here, log this error then try some other things
		logf.Error(ctx, err, "failed to get default Google credentials")
		return "", false, nil
	}
	// If the credentials contains a JSON payload, this payload contains our email
	if defaultCredentials.JSON != nil {
		var data Credentials
		err = json.Unmarshal(defaultCredentials.JSON, &data)
		if err != nil {
			// Wrap Unmarshal error to be more descriptive and give context
			return "", false, errors.Wrap(
				err,
				codes.Internal,
				"failed to get email from default credentials",
				errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("failed to unmarshal Google credentials: %s", err.Error())),
			)
		}

		if data.ClientEmail != "" {
			return data.ClientEmail, true, nil
		}

		logf.Info(ctx, "no client email in Google default credentials")
	} else {
		logf.Info(ctx, "no JSON body present in Google default credentials response")
	}

	return "", false, nil
}

// defaultEmail retrieves the default email address from google
func (c *client) defaultEmail(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s%s", c.config.MetadataAddress, metadataEmailPath)

	logf.Info(ctx, "requesting default email from %s", url)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", errors.Wrap(
			err,
			codes.Unknown,
			"failed to request default email",
			errors.NewErrorInfo(ctx, errcodes.Unknown, "failed to create HTTP request"),
		)
	}

	request.Header.Add("Metadata-Flavor", "Google")
	response, err := c.metadataHttpClient.Do(request)
	if err != nil {
		logf.Error(ctx, err, "failed to get default email")
		return "", errors.Wrap(
			err,
			codes.Internal,
			"failed to request default email",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("get default email metadata request failed: %s", err.Error())),
		)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	statusOK := response.StatusCode >= http.StatusOK && response.StatusCode < 300
	if !statusOK {
		return "", errors.Wrap(
			err,
			apic.CodeFromHTTPStatus(response.StatusCode),
			"failed to request default email",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("vault login response (%d) from login call: %s %s", response.StatusCode, request.Method, request.URL.String())),
		)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(
			err,
			codes.Internal,
			"failed to get default email",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, "failed to read response body of default email metadata"),
		)
	}
	// The email address is the body of the response
	return strings.TrimSpace(string(responseBody)), nil
}

// getJwt creates a signed JWT
func (c *client) getJwt(ctx context.Context, email string) (string, error) {
	logf.Info(ctx, "obtaining signed JWT")
	claims := jwt2.Claims{
		Subject: email,
		Audience: jwt2.Audience{
			fmt.Sprintf("vault/%s", c.config.AuthRole),
		},
		Expiry: jwt2.NewNumericDate(time.Now().UTC().Add(c.config.TokenLifetime)),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", errors.Wrap(
			err,
			codes.Internal,
			"failed to get signed JWT",
			errors.NewErrorInfo(ctx, errcodes.Unknown, fmt.Sprintf("failed to marshal jwt request: %s", err.Error())),
		)
	}

	request := &credentialspb.SignJwtRequest{
		Name:    fmt.Sprintf("projects/-/serviceAccounts/%s", email),
		Payload: string(payload),
	}

	logf.Debug(ctx, "sign JWT from request: %+v", request)

	signedJwtResponse, err := c.jwtSigner.SignJwt(ctx, request)
	if err != nil {
		logf.Error(ctx, err, "vault login failed to sign JWT")
		return "", errors.Wrap(
			err,
			codes.Internal,
			"failed to get signed JWT",
			errors.NewErrorInfo(ctx, errcodes.DownstreamFailure, fmt.Sprintf("error occurred signing JWT: %s", err.Error())),
		)
	}

	return signedJwtResponse.SignedJwt, nil
}
