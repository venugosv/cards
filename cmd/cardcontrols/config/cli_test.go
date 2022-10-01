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
		want:   "spec:\n  appName: \"\"\n  port: 0\n  log:\n    level: \"\"\n    payloadDecider:\n      server: {}\n      client: {}\n  auth:\n    issuers: []\n    staticKeys: []\n  featureToggles: {}\n  forgerock: null\n  fakerock: null\n",
	}
	t.Run(test.name, func(t *testing.T) {
		got := test.fields.String()
		assert.Equal(t, test.want, got)
	})
}

func TestLoadAndValidateConfig(t *testing.T) {
	t.Run("success load config", func(t *testing.T) {
		args := []string{"/cardcontrols", "-c", "../../../config/cardcontrols/config/app/local.yaml"}
		if len(args) > 0 {
			t.Cleanup(func() {
				pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
			})
			os.Args = args
		}
		want := "spec:\n  appName: CardControls\n  port: 8070\n  log:\n    level: debug\n    payloadDecider:\n      server:\n        /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/block: true\n        /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/list: true\n        /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/query: true\n        /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/remove: true\n        /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/set: true\n        /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/blockcard: true\n        /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/listcontrols: true\n        /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/querycontrols: true\n        /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/removecontrols: true\n        /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/setcontrols: true\n        /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/transfercontrols: true\n      client:\n        /fabric.service.eligibility.v1beta1.cardeligibilityapi/can: false\n        /fabric.service.entitlements.v1beta1.cardentitlementsapi/getentitledcard: true\n        /fabric.service.entitlements.v1beta1.cardentitlementsapi/listentitledcards: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/createcontrols: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/deletecontrols: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/getcontroldocument: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/listcontroldocuments: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/register: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/updateaccount: true\n        /gateway.visa.service.customerrules.v1.customerrulesapi/updatecontrols: true\n  entitlements:\n    baseURL: http://localhost:9060\n  eligibility:\n    baseURL: http://localhost:8070\n  auth:\n    issuers:\n    - name: fakerock.sit.fabric.gcpnp.anz\n      jwksUrl: http://localhost:9080/.well-known/jwks.json\n      cacheTTL: 30m0s\n      cacheRefresh: 0s\n    staticKeys: []\n    insecure: true\n  visa:\n    baseURL: http://localhost:9070/vctc\n    clientIDEnvKey: apic-ecom-client-id-np\n    maxRetries: 3\n  visaGateway:\n    baseURL: http://localhost:7080\n    clientID: \"\"\n  ctm:\n    baseURL: http://localhost:9070/ctm\n    clientIDEnvKey: apic-corp-client-id-np\n    maxRetries: 3\n  commandCentre:\n    pubsubEmulatorHost: localhost:8185\n    env: local\n  vault:\n    vaultAddress: http://localhost:9070/vault\n    authRole: gcpiamrole-fabric-encdec.common\n    localToken: \"\"\n    authPath: v1/auth/gcp-fabric\n    namespace: eaas-test\n    zone: corp\n    metadataAddress: \"\"\n    overrideServiceEmail: fabric@anz.com\n    noGoogleCredentialsClient: true\n    tokenLifetime: 5m0s\n    tokenRenewBuffer: 2m0s\n    blockForTokenTime: 0s\n    tokenErrorRetryTime: 0s\n    tokenErrorRetryMaxTime: 5m0s\n  featureToggles:\n    rpc:\n      /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/block: true\n      /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/list: true\n      /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/query: true\n      /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/remove: true\n      /fabric.service.cardcontrols.v1beta1.cardcontrolsapi/set: true\n      /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/blockcard: true\n      /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/listcontrols: true\n      /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/querycontrols: true\n      /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/removecontrols: true\n      /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/setcontrols: true\n      /fabric.service.cardcontrols.v1beta2.cardcontrolsapi/transfercontrols: true\n    features:\n      DCVV2: true\n      FORGEROCK_SYSTEM_LOGIN: true\n      MCT_GAMBLING: true\n      TCT_ATM_WITHDRAW: true\n      TCT_CONTACTLESS: true\n      TCT_CROSS_BORDER: true\n      TCT_E_COMMERCE: true\n  auditlog:\n    name: fabric-cardcontrols\n    domain: fabric.gcp.anz\n    provider: fabric\n    pubsub:\n      projectID: auditlog\n      topicID: auditlog\n      emulatorHost: localhost:8086\n  ocv:\n    baseURL: http://localhost:9070/ocv\n    clientIDEnvKey: apic-corp-client-id-np\n    maxRetries: 3\n    enableLogging: true\n  forgerock:\n    baseURL: http://localhost:9070/forgerock/\n    clientID: fabric-cardcontrols\n    clientSecretKey: cardcontrols-forgerock-secret-np\n  fakerock: null\nops:\n  port: 8082\n  opentelemetry:\n    trace:\n      exporter: jaeger\n      type: \"\"\n      sampleProbability: 0\n    metrics:\n      exporter: prometheus\n      pushPeriod: 0s\n    exporters:\n      jaeger:\n        collectorEndpoint: http://localhost:14268/api/traces\n"

		got, err := Load()
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, want, got.String())
	})
}
