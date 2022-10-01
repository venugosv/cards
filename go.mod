module github.com/anzx/fabric-cards

go 1.17

require (
	cloud.google.com/go/iam v0.3.0
	cloud.google.com/go/pubsub v1.25.1
	github.com/alicebob/miniredis/v2 v2.21.0
	github.com/anz-bank/equals v0.0.0-20210608071158-94d986b192f6
	github.com/anz-bank/pkg v0.0.43
	github.com/anzx/anzdata v0.20.0
	github.com/anzx/fabric-commandcentre-sdk v1.19.1
	github.com/anzx/fabric-pnv v0.7.0
	github.com/anzx/fabric-visa-gateway v1.2.3
	github.com/anzx/fabricapis/pkg/fabric/service/accounts/v1alpha6 v0.7.4
	github.com/anzx/fabricapis/pkg/fabric/service/card/v1beta1 v0.7.4
	github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta1 v0.4.3
	github.com/anzx/fabricapis/pkg/fabric/service/cardcontrols/v1beta2 v0.0.7
	github.com/anzx/fabricapis/pkg/fabric/service/commandcentre/v1beta1 v1.2.3
	github.com/anzx/fabricapis/pkg/fabric/service/eligibility/v1beta1 v0.2.4
	github.com/anzx/fabricapis/pkg/fabric/service/entitlements/v1beta1 v0.0.26
	github.com/anzx/fabricapis/pkg/fabric/service/fakerock/v1alpha1 v0.1.11
	github.com/anzx/fabricapis/pkg/fabric/service/selfservice/v1beta2 v0.3.0
	github.com/anzx/fabricapis/pkg/fabric/type v0.9.0
	github.com/anzx/fabricapis/pkg/fabric/type/audit v0.11.0
	github.com/anzx/fabricapis/pkg/gateway/visa/service/cardonfile v0.0.3
	github.com/anzx/fabricapis/pkg/gateway/visa/service/customerrules v0.0.13
	github.com/anzx/fabricapis/pkg/gateway/visa/service/dcvv2 v0.0.1
	github.com/anzx/fabricapis/pkg/visa/service/enrollmentcallback v0.0.7
	github.com/anzx/fabricapis/pkg/visa/service/notificationcallback v0.0.8
	github.com/anzx/pkg/accountformatter v1.0.0
	github.com/anzx/pkg/auditlog v0.8.0
	github.com/anzx/pkg/errors v0.8.0
	github.com/anzx/pkg/gsm v0.2.0
	github.com/anzx/pkg/jsontime v0.4.0
	github.com/anzx/pkg/jwtauth v0.14.0
	github.com/anzx/pkg/log v0.1.9
	github.com/anzx/pkg/log/fabriclog v0.1.11
	github.com/anzx/pkg/monitoring v0.3.19
	github.com/anzx/pkg/opentelemetry v0.23.7
	github.com/anzx/pkg/validator v0.1.2
	github.com/anzx/pkg/xcontext v0.1.0
	github.com/anzx/utils/forgejwt/v2 v2.1.8
	github.com/brehv/r v0.0.0-20210715230501-94fcba6f5df7
	github.com/brianvoe/gofakeit/v6 v6.18.0
	github.com/cenkalti/backoff/v4 v4.1.3
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redis_rate/v9 v9.1.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3
	github.com/hashicorp/go-retryablehttp v0.7.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/otel/trace v1.7.0
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e
	golang.org/x/oauth2 v0.0.0-20220822191816-0ebed06d0094
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	google.golang.org/api v0.93.0
	google.golang.org/genproto v0.0.0-20220822174746-9e6da59bd2fc
	google.golang.org/grpc v1.49.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go v0.104.0 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/secretmanager v1.4.0 // indirect
	cloud.google.com/go/spanner v1.24.0 // indirect
	cloud.google.com/go/trace v1.2.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v0.32.7 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.3.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/alecthomas/repr v0.0.0-20210801044451-80ca428c5142 // indirect
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/anz-bank/boomer v0.0.0-20220517015306-c4e48f260cce // indirect
	github.com/anzx/fabricapis/pkg/fabric/api v0.5.0 // indirect
	github.com/anzx/pkg/guid v0.1.4 // indirect
	github.com/anzx/pkg/log/monitoring v0.1.1 // indirect
	github.com/anzx/pkg/logging v0.4.4 // indirect
	github.com/anzx/pkg/opencensus v1.3.1 // indirect
	github.com/anzx/pkg/uuid v0.2.0 // indirect
	github.com/arr-ai/frozen v0.20.0 // indirect
	github.com/arr-ai/hash v0.8.0 // indirect
	github.com/asaskevich/EventBus v0.0.0-20200907212545-49d423059eef // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.3 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.10.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jarcoal/httpmock v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.12.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/rs/zerolog v1.27.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/subosito/gotenv v1.3.0 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.5.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	github.com/zeromq/goczmq v4.1.0+incompatible // indirect
	github.com/zeromq/gomq v0.0.0-20201031135124-cef4e507bb8e // indirect
	github.com/zeromq/gomq/zmtp v0.0.0-20201031135124-cef4e507bb8e // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.4.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/bridge/opencensus v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.4.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.4.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.4.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.4.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.4.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.27.0 // indirect
	go.opentelemetry.io/otel/metric v0.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.4.1 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.27.0 // indirect
	go.opentelemetry.io/proto/otlp v0.12.0 // indirect
	gocloud.dev v0.26.0 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
)

replace (
	github.com/anzx/pkg/opentelemetry => github.com/anzx/pkg/opentelemetry v0.23.1
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
)

exclude (
	github.com/gogo/protobuf v1.1.0
	github.com/gogo/protobuf v1.1.1
	github.com/gogo/protobuf v1.2.0
	github.com/gogo/protobuf v1.2.1
	github.com/gogo/protobuf v1.3.1
)
