module github.com/manabie-com/backend

go 1.20

require (
	cloud.google.com/go/cloudsqlconn v1.1.1
	cloud.google.com/go/profiler v0.1.2
	cloud.google.com/go/storage v1.29.0
	cloud.google.com/go/texttospeech v1.6.0
	cloud.google.com/go/vision v1.2.0
	code.cloudfoundry.org/bytefmt v0.0.0-20200131002437-cf55d5288a48
	contrib.go.opencensus.io/exporter/prometheus v0.4.2
	firebase.google.com/go v3.13.0+incompatible
	firebase.google.com/go/v4 v4.6.0
	github.com/GoogleContainerTools/skaffold v1.39.14
	github.com/GoogleContainerTools/skaffold/v2 v2.5.1
	github.com/Nerzal/gocloak/v10 v10.0.1
	github.com/PuerkitoBio/goquery v1.7.1
	github.com/Shopify/sarama v1.37.2
	github.com/abadojack/whatlanggo v1.0.1
	github.com/awa/go-iap v1.3.5
	github.com/aws/aws-sdk-go v1.44.50
	github.com/bradleyfalzon/ghinstallation/v2 v2.1.0
	github.com/buger/jsonparser v1.1.1
	github.com/bxcodec/faker/v3 v3.8.0
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/cucumber/godog v0.12.6
	github.com/cucumber/messages-go/v16 v16.0.1
	github.com/dgraph-io/ristretto v0.0.3
	github.com/docker/docker v23.0.3+incompatible
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/ernesto-jimenez/httplogger v0.0.0-20220128121225-117514c3f345
	github.com/ettle/strcase v0.1.1
	github.com/gin-contrib/zap v0.0.1
	github.com/gin-gonic/gin v1.9.1
	github.com/go-kafka/connect v0.9.0
	github.com/go-pg/pg v8.0.7+incompatible
	github.com/go-playground/validator/v10 v10.14.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/gobeam/stringy v0.0.5
	github.com/gocarina/gocsv v0.0.0-20220823132111-71f3a5cb2654
	github.com/gogo/protobuf v1.3.2
	github.com/golang-migrate/migrate/v4 v4.16.2
	github.com/golang/protobuf v1.5.3
	github.com/google/go-cmp v0.5.9
	github.com/google/go-github/v41 v41.0.0
	github.com/google/go-jsonnet v0.18.0
	github.com/google/oauth2l v1.2.2
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.8.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/hashicorp/golang-lru/v2 v2.0.2
	github.com/hasura/go-graphql-client v0.7.1
	github.com/ianlopshire/go-fixedwidth v0.9.3
	github.com/imdario/mergo v0.3.13
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733
	github.com/jackc/pgconn v1.14.0
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgproto3/v2 v2.3.2
	github.com/jackc/pgtype v1.14.0
	github.com/jackc/pgx/v4 v4.18.1
	github.com/jackc/puddle v1.3.0
	github.com/jarcoal/httpmock v1.1.0
	github.com/jedib0t/go-pretty/v6 v6.3.2
	github.com/joho/godotenv v1.5.1
	github.com/ktr0731/grpc-web-go-client v0.2.7
	github.com/lestrrat-go/jwx v1.2.26
	github.com/lestrrat/go-jwx v0.0.0-20180221005942-b7d4802280ae
	github.com/lib/pq v1.10.6
	github.com/magiconair/properties v1.8.7
	github.com/mailru/easyjson v0.7.7
	github.com/manabie-com/j4 v0.0.0-20221101070859-588b9a5aa26c
	github.com/manifoldco/promptui v0.9.0
	github.com/minio/minio-go/v7 v7.0.12
	github.com/nats-io/nats.go v1.21.0
	github.com/nleeper/goment v1.4.5-0.20221117170701-54447c7bdcf9
	github.com/nyaruka/phonenumbers v1.0.75
	github.com/oklog/ulid/v2 v2.0.2
	github.com/olivere/elastic/v7 v7.0.29
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/pganalyze/pg_query_go/v2 v2.1.0
	github.com/pkg/errors v0.9.1
	github.com/processout/grpc-go-pool v1.2.1
	github.com/prometheus/client_golang v1.15.0
	github.com/r3labs/diff/v3 v3.0.0
	github.com/redis/go-redis/v9 v9.1.0
	github.com/robfig/cron/v3 v3.0.0
	github.com/segmentio/ksuid v1.0.4
	github.com/sendgrid/rest v2.6.9+incompatible
	github.com/sendgrid/sendgrid-go v3.12.0+incompatible
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/square/go-jose/v3 v3.0.0-20200630053402-0a67ce9b0693
	github.com/stretchr/testify v1.8.4
	github.com/thmeitz/ksqldb-go v0.0.4
	github.com/tidwall/gjson v1.14.0
	github.com/tkuchiki/faketime v0.1.1
	github.com/vektra/mockery/v2 v2.14.0
	github.com/vmihailenco/taskq/v3 v3.2.8
	github.com/y-bash/go-gaga v0.0.2
	github.com/yeqown/go-qrcode/v2 v2.0.2
	github.com/yeqown/go-qrcode/writer/standard v1.1.1
	github.com/yoheimuta/go-protoparser v3.4.0+incompatible
	github.com/yudai/gojsondiff v1.0.0
	go.mongodb.org/mongo-driver v1.11.3
	go.mozilla.org/sops/v3 v3.7.1
	go.opencensus.io v0.24.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.40.0
	go.opentelemetry.io/contrib/propagators/b3 v1.9.0
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/exporters/jaeger v1.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.14.0
	go.opentelemetry.io/otel/sdk v1.14.0
	go.opentelemetry.io/otel/trace v1.14.0
	go.uber.org/automaxprocs v1.5.2
	go.uber.org/multierr v1.10.0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.9.0
	golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0
	golang.org/x/mod v0.10.0
	golang.org/x/net v0.10.0
	golang.org/x/sync v0.2.0
	golang.org/x/tools v0.9.1
	google.golang.org/api v0.119.0
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1
	google.golang.org/grpc v1.54.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools v2.2.0+incompatible
	istio.io/client-go v1.17.1
	k8s.io/api v0.27.4
	k8s.io/apimachinery v0.27.4
	k8s.io/autoscaler/vertical-pod-autoscaler v0.12.0
	k8s.io/client-go v0.27.4

)

require (
	bou.ke/monkey v1.0.2 // indirect
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/compute v1.19.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/firestore v1.9.0 // indirect
	cloud.google.com/go/iam v0.13.0 // indirect
	cloud.google.com/go/longrunning v0.4.1 // indirect
	cloud.google.com/go/vision/v2 v2.7.0 // indirect
	filippo.io/age v1.0.0 // indirect
	github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src v0.0.0-20230727073715-5c800b136f13
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.28 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.22 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.12 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/log-go v0.4.0
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/Unleash/unleash-client-go/v3 v3.7.0
	github.com/andybalholm/cascadia v1.2.0 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20211106181442-e4c1a74c66bd // indirect
	github.com/armon/go-metrics v0.4.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/bsm/redislock v0.7.1 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/capnm/sysinfo v0.0.0-20130621111458-5909a53897f3 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/containerd/containerd v1.7.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.10.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-redis/redis/v8 v8.11.4 // indirect
	github.com/go-redis/redis_rate/v9 v9.1.2 // indirect
	github.com/go-test/deep v1.0.4 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-jwt/jwt/v4 v4.4.3 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-containerregistry v0.14.0 // indirect
	github.com/google/go-github/v45 v45.2.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/google/s2a-go v0.1.2 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/goware/prefixer v0.0.0-20160118172347-395022866408 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.2.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.4 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/raft v1.3.9 // indirect
	github.com/hashicorp/raft-boltdb/v2 v2.2.2 // indirect
	github.com/hashicorp/vault/api v1.7.2 // indirect
	github.com/hashicorp/vault/sdk v0.5.2 // indirect
	github.com/hashicorp/yamux v0.0.0-20211028200310-0bc27b27de87 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/klauspost/compress v1.16.4 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/krishicks/yaml-patch v0.0.10 // indirect
	github.com/ktr0731/grpc-test v0.1.10 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lestrrat/go-pdebug v0.0.0-20180220043741-569c97477ae8 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/md5-simd v1.1.0 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.6.6 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats-server/v2 v2.7.4 // indirect
	github.com/nats-io/nkeys v0.3.0
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oauth2-proxy/mockoidc v0.0.0-20220221072942-e3afe97dec43
	github.com/oklog/run v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/otiai10/copy v1.12.0
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/prometheus/statsd_exporter v0.22.7 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rqlite/go-sqlite3 v1.25.0 // indirect
	github.com/rqlite/gorqlite v0.0.0-20220528150909-c4e99ae96be6 // indirect
	github.com/rqlite/rqlite v0.0.0-20220807131317-697782280fc2 // indirect
	github.com/rs/xid v1.3.0 // indirect
	github.com/rs/zerolog v1.27.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/segmentio/kafka-go v0.4.39
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.2
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.15.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tkuchiki/go-timezone v0.2.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/twmb/murmur3 v1.1.5 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/yeqown/reedsolomon v1.0.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.mozilla.org/gopgagent v0.0.0-20170926210634-4d7ea76ff71a // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.14.0 // indirect
	go.opentelemetry.io/otel/metric v0.37.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/image v0.7.0 // indirect
	golang.org/x/oauth2 v0.7.0
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	golang.org/x/text v0.9.0
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20230217221049-9d422bf48675 // indirect
	k8s.io/klog/v2 v2.90.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/utils v0.0.0-20230220204549-a5ecb0141aa5
	mellium.im/sasl v0.3.1 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/rqlite/rqlite => github.com/rqlite/rqlite v0.0.0-20220807131317-697782280fc2

replace github.com/miekg/dns v1.0.14 => github.com/miekg/dns v1.1.43

// Fix vulnerability https://github.com/advisories/GHSA-77vh-xpmg-72qh (requires v1.0.2+)
replace github.com/opencontainers/image-spec v1.0.0 => github.com/opencontainers/image-spec v1.0.2

// Fix vulnerability https://github.com/advisories/GHSA-77vh-xpmg-72qh (requires v1.0.2+)
replace github.com/opencontainers/image-spec v1.0.1 => github.com/opencontainers/image-spec v1.0.2

// Fix vulnerability https://github.com/advisories/GHSA-5j5w-g665-5m35 (requires v1.5.8+)
// Fix another vulnerability https://github.com/advisories/GHSA-5ffw-gxpp-mxpf (requires v1.5.13+)
replace github.com/containerd/containerd v1.5.7 => github.com/containerd/containerd v1.5.13

// Force using pgx/v4
replace github.com/jackc/pgx v3.6.2+incompatible => github.com/jackc/pgx/v4 v4.14.1

// Use manabie fork instead
replace github.com/ktr0731/grpc-web-go-client v0.2.7 => github.com/manabie-com/grpc-web-go-client v0.2.9

// Fix vulnerability https://github.com/advisories/GHSA-v95c-p5hm-xq8f (requires v1.0.3+)
replace github.com/opencontainers/runc v1.0.2 => github.com/opencontainers/runc v1.0.3

// Fix a file type confusion at https://github.com/distribution/distribution/commit/b59a6f827947f9e0e67df0cfb571046de4733586
replace github.com/docker/distribution v2.7.1+incompatible => github.com/docker/distribution v2.8.0+incompatible

// Fix another vulnerability https://nvd.nist.gov/vuln/detail/CVE-2022-23471 (requires v1.6.12+)
replace github.com/containerd/containerd v1.6.6 => github.com/containerd/containerd v1.6.12
