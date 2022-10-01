package vault_external

import "time"

type Config struct {
	// Address is the vault URL we connect to
	Address string `json:"vaultAddress" yaml:"vaultAddress" mapstructure:"vaultAddress"`
	// AuthRole is role we request from vault on login (and forms the audience of our JWT)
	// There is also a vault "role" we use when doing a transformation. This is a different role!
	AuthRole string `json:"authRole" yaml:"authRole" mapstructure:"authRole"`
	// LocalToken hard-codes a constant token instead of requesting one from vault, for stubbed environments
	LocalToken string `json:"localToken" yaml:"localToken" mapstrucutre:"localToken"`
	// AuthPath is the base of the URL we use to login
	AuthPath string `json:"authPath" yaml:"authPath" mapstructure:"authPath"`
	// NameSpace is the vault namespace we are working in
	NameSpace string `json:"namespace" yaml:"namespace" mapstructure:"namespace"`
	// Zone is the vault zone we are working in
	Zone string `json:"zone" yaml:"zone" mapstructure:"zone"`
	// MetadataAddress is an address we attempt to get a service email from if google default credentials fails
	MetadataAddress string `json:"metadataAddress" yaml:"metadataAddress" mapstructure:"metadataAddress"`
	// OverrideServiceEmail uses a hardcoded service email, for local testing and stubs
	OverrideServiceEmail string `json:"overrideServiceEmail" yaml:"overrideServiceEmail" mapstructure:"overrideServiceEmail"`
	// NoGoogleCredentialsClient uses a hard-coded JWT signer for testing rather than using the google credentials client
	NoGoogleCredentialsClient bool `json:"noGoogleCredentialsClient" yaml:"noGoogleCredentialsClient" mapstructure:"noGoogleCredentialsClient"`
	// TokenLifetime is the expected lifetime of a vault token
	TokenLifetime time.Duration `json:"tokenLifetime" yaml:"tokenLifetime" mapstructure:"tokenLifetime"`
	// TokenRenewBuffer is what we subtract from TokenLifetime to get our timer delay eg. (TokenLifetime - TokenRenewBuffer)
	TokenRenewBuffer time.Duration `json:"tokenRenewBuffer" yaml:"tokenRenewBuffer" mapstructure:"tokenRenewBuffer"`
	// BlockForTokenTime is how long a REST call blocks waiting for a new vault token
	BlockForTokenTime time.Duration `json:"blockForTokenTime" yaml:"blockForTokenTime" mapstructure:"blockForTokenTime"`
	// TokenErrorRetryFirstTime is the initial value for our retry-backoff timer if token renew fails
	TokenErrorRetryFirstTime time.Duration `json:"tokenErrorRetryTime" yaml:"tokenErrorRetryTime" mapstructure:"tokenErrorRetryTime"`
	// TokenErrorRetryMaxTime is the maximum value for our retry-backoff timer if our token renew continues to fail
	TokenErrorRetryMaxTime time.Duration `json:"tokenErrorRetryMaxTime" yaml:"tokenErrorRetryMaxTime" mapstructure:"tokenErrorRetryMaxTime"`
}
