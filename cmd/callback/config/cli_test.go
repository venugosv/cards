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
		want:   "spec:\n  appName: \"\"\n  port: 0\n  log:\n    level: \"\"\n    payloadDecider:\n      server: {}\n      client: {}\n  featureToggles: {}\n  forgerock: null\n  fakerock: null\n  certificates: null\n",
	}
	t.Run(test.name, func(t *testing.T) {
		got := test.fields.String()
		assert.Equal(t, test.want, got)
	})
}

func TestLoadAndValidateConfig(t *testing.T) {
	t.Run("success callback load config", func(t *testing.T) {
		args := []string{"/callback", "-c", "../../../config/callback/config/app/local.yaml"}
		if len(args) > 0 {
			t.Cleanup(func() {
				pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
			})
			os.Args = args
		}
		want := "spec:\n  appName: Callback\n  port: 8060\n  log:\n    level: debug\n    payloadDecider:\n      server:\n        /visa.service.enrollmentcallback.v1.enrollmentcallbackapi/disenroll: true\n        /visa.service.enrollmentcallback.v1.enrollmentcallbackapi/enroll: true\n        /visa.service.notificationcallback.v1.notificationcallbackapi/alert: true\n      client: {}\n  ctm:\n    baseURL: http://localhost:9070/ctm\n    clientIDEnvKey: apic-corp-client-id-np\n    maxRetries: 3\n  vault:\n    vaultAddress: http://localhost:9070/vault\n    authRole: gcpiamrole-fabric-encdec.common\n    localToken: \"\"\n    authPath: v1/auth/gcp-fabric\n    namespace: eaas-test\n    zone: corp\n    metadataAddress: \"\"\n    overrideServiceEmail: fabric@anz.com\n    noGoogleCredentialsClient: true\n    tokenLifetime: 5m0s\n    tokenRenewBuffer: 2m0s\n    blockForTokenTime: 0s\n    tokenErrorRetryTime: 0s\n    tokenErrorRetryMaxTime: 5m0s\n  commandCentre:\n    pubsubEmulatorHost: localhost:8185\n    env: local\n  featureToggles:\n    rpc:\n      /visa.service.enrollmentcallback.v1.enrollmentcallbackapi/disenroll: true\n      /visa.service.enrollmentcallback.v1.enrollmentcallbackapi/enroll: true\n      /visa.service.notificationcallback.v1.notificationcallbackapi/alert: true\n    features:\n      ENROLLMENT_CALLBACK_INTEGRATED: true\n      FORGEROCK_SYSTEM_LOGIN: true\n      NOTIFICATION_CALLBACK_DECLINED_EVENT: true\n  forgerock:\n    baseURL: http://localhost:9070/forgerock/\n    clientID: fabric-visa-callback\n    clientSecretKey: callback-forgerock-secret-np\n  fakerock: null\n  certificates: null\nops:\n  port: 8062\n  opentelemetry:\n    trace:\n      exporter: jaeger\n      type: \"\"\n      sampleProbability: 0\n    metrics:\n      exporter: prometheus\n      pushPeriod: 0s\n    exporters:\n      jaeger:\n        collectorEndpoint: http://localhost:14268/api/traces\n"

		got, err := Load()
		require.NoError(t, err)
		assert.Equal(t, want, got.String())
		assert.Nil(t, err)
	})
}
