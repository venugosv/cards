package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/pflag"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigMarshalJSON(t *testing.T) {
	c := Config{}
	rc := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	require.NotPanics(t, func() { c.ServeJSON(rc, req) })
	resp := rc.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("content-type"))
}

func TestConfigMarshalYAML(t *testing.T) {
	c := Config{}
	rc := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err)
	require.NotPanics(t, func() { c.ServeYAML(rc, req) })
	resp := rc.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "application/yaml", resp.Header.Get("content-type"))
}

func TestConfigString(t *testing.T) {
	test := struct {
		name   string
		fields Config
		want   string
	}{
		name:   "successful test of config to string",
		fields: Config{},
		want:   "spec:\n  appName: \"\"\n  port: 0\n  log:\n    level: \"\"\n    payloadDecider:\n      server: {}\n      client: {}\n  auth:\n    issuers: []\n    staticKeys: []\n  featureToggles: {}\n",
	}
	t.Run(test.name, func(t *testing.T) {
		got := test.fields.String()
		assert.Equal(t, test.want, got)
	})
}

func TestLoadAndValidateConfig(t *testing.T) {
	t.Run("success cards load config", func(t *testing.T) {
		args := []string{"/cards", "-c", "../../../config/cards/config/app/local.yaml"}
		if len(args) > 0 {
			t.Cleanup(func() {
				pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
			})
			os.Args = args
		}

		want := "spec:\n  appName: Cards\n  port: 8080\n  log:\n    level: debug\n    payloadDecider:\n      server:\n        /fabric.service.card.v1beta1.cardapi/activate: true\n        /fabric.service.card.v1beta1.cardapi/audittrail: true\n        /fabric.service.card.v1beta1.cardapi/changepin: true\n        /fabric.service.card.v1beta1.cardapi/getdetails: false\n        /fabric.service.card.v1beta1.cardapi/getwrappingkey: false\n        /fabric.service.card.v1beta1.cardapi/list: true\n        /fabric.service.card.v1beta1.cardapi/replace: true\n        /fabric.service.card.v1beta1.cardapi/resetpin: true\n        /fabric.service.card.v1beta1.cardapi/setpin: true\n        /fabric.service.card.v1beta1.cardapi/verifypin: true\n        /fabric.service.eligibility.v1beta1.cardeligibilityapi/can: true\n      client:\n        /fabric.service.accounts.v1alpha6.accountapi/getaccountlist: true\n        /fabric.service.eligibility.v1beta1.cardeligibilityapi/can: false\n        /fabric.service.entitlements.v1beta1.cardentitlementsapi/getentitledcard: true\n        /fabric.service.entitlements.v1beta1.cardentitlementsapi/listentitledcards: true\n        /fabric.service.entitlements.v1beta1.entitlementscontrolapi/forcepartytolatest: true\n        /fabric.service.entitlements.v1beta1.entitlementscontrolapi/registercardtopersona: true\n        /fabric.service.selfservice.v1beta2.partyapi/getparty: true\n  entitlements:\n    baseURL: http://localhost:9060\n  eligibility:\n    baseURL: http://localhost:8070\n  auth:\n    issuers:\n    - name: fakerock.sit.fabric.gcpnp.anz\n      jwksUrl: http://localhost:9080/.well-known/jwks.json\n      cacheTTL: 30m0s\n      cacheRefresh: 0s\n    staticKeys: []\n    insecure: true\n  ctm:\n    baseURL: http://localhost:9070/ctm\n    clientIDEnvKey: apic-corp-client-id-np\n    maxRetries: 3\n  echidna:\n    baseURL: http://localhost:9070/ca\n    clientIDEnvKey: apic-ecom-client-id-np\n    maxRetries: 3\n  rateLimit:\n    redis:\n      addr: localhost:6379\n      db: 0\n      secretId: testSecretId\n    limits:\n      activate:\n        rate: 5\n        period: 1m0s\n  selfService:\n    baseURL: http://localhost:9060\n  vault:\n    vaultAddress: http://localhost:9070/vault\n    authRole: gcpiamrole-fabric-encdec.common\n    localToken: \"\"\n    authPath: v1/auth/gcp-fabric\n    namespace: eaas-test\n    zone: corp\n    metadataAddress: \"\"\n    overrideServiceEmail: fabric@anz.com\n    noGoogleCredentialsClient: true\n    tokenLifetime: 5m0s\n    tokenRenewBuffer: 2m0s\n    blockForTokenTime: 0s\n    tokenErrorRetryTime: 0s\n    tokenErrorRetryMaxTime: 5m0s\n  featureToggles:\n    rpc:\n      /fabric.service.card.v1beta1.cardapi/activate: true\n      /fabric.service.card.v1beta1.cardapi/audittrail: true\n      /fabric.service.card.v1beta1.cardapi/changepin: true\n      /fabric.service.card.v1beta1.cardapi/getdetails: true\n      /fabric.service.card.v1beta1.cardapi/getwrappingkey: true\n      /fabric.service.card.v1beta1.cardapi/list: true\n      /fabric.service.card.v1beta1.cardapi/replace: true\n      /fabric.service.card.v1beta1.cardapi/resetpin: true\n      /fabric.service.card.v1beta1.cardapi/setpin: true\n      /fabric.service.card.v1beta1.cardapi/verifypin: true\n      /fabric.service.card.v1beta1.walletapi/createapplepaymenttoken: true\n      /fabric.service.card.v1beta1.walletapi/creategooglepaymenttoken: true\n      /fabric.service.eligibility.v1beta1.cardeligibilityapi/can: true\n    features:\n      DCVV2: true\n      FORGEROCK_SYSTEM_LOGIN: true\n      PIN_CHANGE_COUNT: true\n      REASON_DAMAGED: true\n      REASON_LOST: true\n      REASON_STOLEN: true\n  auditlog:\n    name: fabric-cards\n    domain: fabric.gcp.anz\n    provider: fabric\n    pubsub:\n      projectID: auditlog\n      topicID: auditlog\n      emulatorHost: localhost:8086\n  ocv:\n    baseURL: http://localhost:9070/ocv\n    clientIDEnvKey: apic-corp-client-id-np\n    maxRetries: 3\n    enableLogging: false\n  visaGateway:\n    baseURL: http://localhost:7080\n    clientID: c5934653-ff6a-46cb-81aa-850f50e6f95b\n  cardcontrols:\n    baseURL: http://localhost:8080\n  apcam:\n    baseURL: http://localhost:9070/apcam\n    clientIDEnvKey: apic-ecom-client-id-np\n    maxRetries: 3\n  forgerock:\n    baseURL: http://localhost:9070/forgerock/\n    clientID: fabric-cards\n    clientSecretKey: cards-forgerock-secret-np\n  gpay:\n    apikeykey: wallet-visa-api-key-np\n    sharedsecretkey: wallet-visa-shared-secret-np\nops:\n  port: 8072\n  opentelemetry:\n    trace:\n      exporter: jaeger\n      type: \"\"\n      sampleProbability: 0\n    metrics:\n      exporter: prometheus\n      pushPeriod: 0s\n    exporters:\n      jaeger:\n        collectorEndpoint: http://localhost:14268/api/traces\n"

		got, err := Load()
		require.NoError(t, err)
		assert.Equal(t, want, got.String())
		assert.Nil(t, err)
	})
}
