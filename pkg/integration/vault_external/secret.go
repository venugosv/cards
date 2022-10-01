package vault_external

type Secret struct {
	RequestID     string `json:"request_id"`
	LeaseID       string `json:"lease_id"`
	LeaseDuration int    `json:"lease_duration"`
	Renewable     bool   `json:"renewable"`

	Auth *SecretAuth `json:"auth,omitempty"`
	Data map[string]interface{}
}

type SecretAuth struct {
	ClientToken      string            `json:"client_token"`
	Accessor         string            `json:"accessor"`
	Policies         []string          `json:"policies"`
	TokenPolicies    []string          `json:"token_policies"`
	IdentityPolicies []string          `json:"identity_policies"`
	Metadata         map[string]string `json:"metadata"`
	Orphan           bool              `json:"orphan"`
	EntityID         string            `json:"entity_id"`

	// LeaseDuration is in seconds
	LeaseDuration int  `json:"lease_duration"`
	Renewable     bool `json:"renewable"`
}
