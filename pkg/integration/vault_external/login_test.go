package vault_external

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/require"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

type ErrorSignedJwt struct {
	jwt   string
	key   string
	error string
}

func (e *ErrorSignedJwt) SignJwt(_ context.Context, _ *credentialspb.SignJwtRequest, _ ...gax.CallOption) (*credentialspb.SignJwtResponse, error) {
	if e.error != "" {
		return nil, errors.New(e.error)
	}
	return &credentialspb.SignJwtResponse{
		SignedJwt: e.jwt,
		KeyId:     e.key,
	}, nil
}

func TestClient_DefaultEmail(t *testing.T) {
	tests := []struct {
		name          string
		httpResponse  string
		httpError     string
		expectedError string
		path          string
		statusCode    int
	}{
		{
			name:         "happy path",
			httpResponse: "ok@anz.com",
			statusCode:   200,
		},
		{
			name:          "fail to create request",
			statusCode:    200,
			path:          "/%20%",
			expectedError: "fabric error: status_code=Unknown, error_code=0, message=failed to request default email, reason=failed to create HTTP request",
		},
		{
			name:          "http 400 error",
			statusCode:    400,
			expectedError: "fabric error: status_code=InvalidArgument, error_code=2, message=failed to request default email, reason=vault login response (400) from login call: GET",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.WriteHeader(test.statusCode)
				_, _ = response.Write([]byte(test.httpResponse))
			}))

			generator := client{
				metadataHttpClient: server.Client(),
				config: &Config{
					MetadataAddress: fmt.Sprintf("%s%s", server.URL, test.path),
				},
			}

			email, err := generator.defaultEmail(context.Background())
			if test.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, test.httpResponse, email)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			}
		})
	}
}

func TestClient_GetJwt(t *testing.T) {
	tests := []struct {
		name          string
		jwtResponse   string
		jwtError      string
		expectedError string
	}{
		{
			name:        "happy path",
			jwtResponse: "foobar",
		},
		{
			name:          "jwt sign fails",
			jwtError:      "bad sign",
			expectedError: "fabric error: status_code=Internal, error_code=2, message=failed to get signed JWT, reason=error occurred signing JWT: bad sign",
		},
	}

	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			jwter := ErrorSignedJwt{
				jwt:   test.jwtResponse,
				error: test.jwtError,
			}
			generator := client{
				jwtSigner: &jwter,
				config: &Config{
					AuthRole: "foobar",
				},
			}

			jwt, err := generator.getJwt(context.Background(), "fabric@anz.com")
			if test.jwtError == "" {
				require.NoError(t, err)
				require.Equal(t, test.jwtResponse, jwt)
			} else {
				require.Error(t, err)
				require.Equal(t, test.expectedError, err.Error())
			}
		})
	}
}

func TestClient_Login(t *testing.T) {
	tests := []struct {
		name          string
		emailError    bool
		jwtError      string
		apiError      string
		response      Secret
		expectedError string
	}{
		{
			name: "happy path",
			response: Secret{
				Auth: &SecretAuth{
					ClientToken:   "foo",
					LeaseDuration: 1234,
				},
			},
		},
		{
			name:          "no auth in response",
			response:      Secret{},
			expectedError: "fabric error: status_code=Internal, error_code=2, message=vault login failed, reason=login response has no auth data",
		},
		{
			name:          "email error",
			emailError:    true,
			expectedError: "fabric error: status_code=Internal, error_code=2, message=failed to request default email, reason=get default email metadata request failed: Get \"/computeMetadata/v1/instance/service-accounts/default/email\": unsupported protocol scheme \"\"",
		},
		{
			name:          "jwt error",
			jwtError:      "foo",
			expectedError: "fabric error: status_code=Internal, error_code=2, message=failed to get signed JWT, reason=error occurred signing JWT: foo",
		},
		{
			name:          "error calling api",
			expectedError: "fabric error: status_code=Internal, error_code=2, message=vault login failed, reason=login response has no auth data",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			c := client{
				config: &Config{
					AuthRole: "some-role",
				},
				jwtSigner: &ErrorSignedJwt{
					jwt:   "foobar",
					key:   "key",
					error: test.jwtError,
				},

				api: &FakeVaultAPI{
					loginResponse: &test.response,
					runError:      test.apiError,
				},
				metadataHttpClient: &http.Client{},
			}

			if !test.emailError {
				_ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/zzz/invalid/path")
				c.config.OverrideServiceEmail = "foo@anz.local"
			}

			_, err := c.login(context.Background())
			if test.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClient_GetServiceEmail(t *testing.T) {
	tests := []struct {
		name               string
		credentialsEmail   string
		defaultEmail       string
		overrideEmail      string
		createMetadataFile bool
		defaultEmailError  bool
		expectedError      string
		expectedEmail      string
	}{
		{
			name:               "get email from metadata",
			createMetadataFile: true,
			credentialsEmail:   "fabric@anz.com",
			expectedEmail:      "fabric@anz.com",
		},
		{
			name:               "failed to get metadata",
			createMetadataFile: false,
			credentialsEmail:   "fail_test@anz.com",
			defaultEmail:       "fabric@anz.com",
			expectedEmail:      "fabric@anz.com",
		},
		{
			name:               "no email in metadata",
			createMetadataFile: true,
			credentialsEmail:   "",
			defaultEmail:       "foo@test.anz",
			expectedEmail:      "foo@test.anz",
		},
		{
			name:               "override email",
			createMetadataFile: true,
			credentialsEmail:   "fabric@anz.com",
			defaultEmail:       "foo@anz.local",
			overrideEmail:      "abc@localhost",
			expectedEmail:      "abc@localhost",
		},
		{
			name:               "default fallthrough fails",
			createMetadataFile: true,
			defaultEmailError:  true,
			expectedError:      "fabric error: status_code=Internal, error_code=2, message=failed to request default email, reason=vault login response (500) from login call: GET ",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			if test.createMetadataFile {
				creds := map[string]interface{}{
					"web": map[string]interface{}{
						"client_id":     "xxxx",
						"redirect_uris": []string{"http://anz.local"},
					},
					"client_email": test.credentialsEmail,
				}
				b, err := json.Marshal(creds)
				require.NoError(t, err, "failed to marshal google metadata json")

				file, err := ioutil.TempFile("/tmp", ".metadata")
				require.NoError(t, err, "failed to create tmp testing metadata file")

				_, err = file.Write(b)
				require.NoError(t, err, "failed to write google metadata to file")

				err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", file.Name())
				require.NoError(t, err, "failed to set google metadata env var")
			} else {
				err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/.this_file_does_not_exist")
				require.NoError(t, err, "failed to set google metadata env var")
			}

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				if test.defaultEmailError {
					rw.WriteHeader(500)
					_, _ = rw.Write([]byte("bad"))
					return
				}
				_, _ = rw.Write([]byte(test.defaultEmail))
			}))

			client := client{
				config: &Config{
					OverrideServiceEmail: test.overrideEmail,
					MetadataAddress:      server.URL,
				},
				metadataHttpClient: server.Client(),
			}
			email, err := client.getServiceEmail(context.Background())
			if test.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedEmail, email)
			}
		})
	}
}
