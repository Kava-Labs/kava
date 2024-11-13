module github.com/kava-labs/kava

go 1.21.0

toolchain go1.21.9

require (
	cosmossdk.io/api v0.7.5 // indirect
	cosmossdk.io/core v0.11.1
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/log v1.4.1
	cosmossdk.io/math v1.3.0
	cosmossdk.io/simapp v0.0.0-20231127212628-044ff4d8c015
	cosmossdk.io/store v1.1.1
	cosmossdk.io/x/evidence v0.1.1
	cosmossdk.io/x/tx v0.13.5
	//github.com/Kava-Labs/opendb v0.0.0-20241024062216-dd5bc73e28f2
	github.com/Kava-Labs/opendb v0.0.0-20241024062216-dd5bc73e28f2
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cometbft/cometbft v1.0.0-rc1.0.20240908111210-ab0be101882f
	github.com/cometbft/cometbft-db v0.11.0
	github.com/cosmos/cosmos-db v1.0.2
	github.com/cosmos/cosmos-proto v1.0.0-beta.5
	github.com/cosmos/cosmos-sdk v0.50.10
	github.com/cosmos/cosmos-sdk/x/upgrade v0.1.4
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/gogoproto v1.7.0
	github.com/cosmos/iavl v1.2.0
	github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8 v8.0.2
	github.com/cosmos/ibc-go/modules/capability v1.0.1
	github.com/cosmos/ibc-go/v8 v8.5.1
	github.com/ethereum/go-ethereum v1.10.26
	github.com/evmos/ethermint v0.0.0-00010101000000-000000000000
	github.com/go-kit/kit v0.13.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.5.4
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/linxGnu/grocksdb v1.9.3
	github.com/pelletier/go-toml/v2 v2.2.3
	github.com/prometheus/client_golang v1.20.4
	github.com/spf13/cast v1.7.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.9.0
	github.com/subosito/gotenv v1.6.0
	golang.org/x/crypto v0.28.0
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8
	google.golang.org/genproto/googleapis/api v0.0.0-20240814211410-ddb44dafa142
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
	sigs.k8s.io/yaml v1.4.0
)

require (
	cosmossdk.io/client/v2 v2.0.0-beta.3
	cosmossdk.io/x/upgrade v0.1.4
)

require (
	cloud.google.com/go v0.115.1 // indirect
	cloud.google.com/go/auth v0.8.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.3 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/iam v1.1.13 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	cosmossdk.io/collections v0.4.0 // indirect
	cosmossdk.io/depinject v1.0.0 // indirect
	cosmossdk.io/x/feegrant v0.1.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.2 // indirect
	github.com/DataDog/datadog-go v3.2.0+incompatible // indirect
	github.com/DataDog/zstd v1.5.5 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.1 // indirect
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d // indirect
	github.com/bgentry/speakeasy v0.2.0 // indirect
	github.com/bits-and-blooms/bitset v1.8.0 // indirect
	github.com/btcsuite/btcd v0.24.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.6 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cockroachdb/apd/v2 v2.0.2 // indirect
	github.com/cockroachdb/errors v1.11.3 // indirect
	github.com/cockroachdb/fifo v0.0.0-20240606204812-0bbfbd93a7ce // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/pebble v1.1.1 // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/coinbase/rosetta-sdk-go/types v1.0.0 // indirect
	github.com/cosmos/btcutil v1.0.5 // indirect
	github.com/cosmos/gogogateway v1.2.0 // indirect
	github.com/cosmos/ics23/go v0.11.0 // indirect
	github.com/cosmos/ledger-cosmos-go v0.13.3 // indirect
	github.com/cosmos/rosetta v0.50.6 // indirect
	github.com/cosmos/rosetta-sdk-go v0.10.0 // indirect
	github.com/danieljoos/wincred v1.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.1.2-0.20240116140435-c67e07994f91 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/dop251/goja v0.0.0-20230806174421-c933cf95e127 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/emicklei/dot v1.6.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gballet/go-libpcsclite v0.0.0-20190607065134-2772fd86a8ff // indirect
	github.com/getsentry/sentry-go v0.27.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-sourcemap/sourcemap v2.1.3+incompatible // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/orderedcode v0.0.1 // indirect
	github.com/google/pprof v0.0.0-20230228050547-1710fef4ab10 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/goware/urlx v0.3.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-getter v1.7.6 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.3 // indirect
	github.com/hashicorp/go-plugin v1.6.0 // indirect
	github.com/hashicorp/go-safetemp v1.0.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/hdevalence/ed25519consensus v0.2.0 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.2.4 // indirect
	github.com/huandu/skiplist v1.2.1 // indirect
	github.com/huin/goupnp v1.3.0 // indirect
	github.com/iancoleman/orderedmap v0.3.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/improbable-eng/grpc-web v0.15.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20230904125328-1f23a7beb09a // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.60.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/prometheus/tsdb v0.7.1 // indirect
	github.com/rakyll/statik v0.1.7 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sasha-s/go-deadlock v0.3.5 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/status-im/keycard-go v0.2.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d // indirect
	github.com/tendermint/go-amino v0.16.0 // indirect
	github.com/tidwall/btree v1.7.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/zondax/hid v0.9.2 // indirect
	github.com/zondax/ledger-go v0.14.3 // indirect
	go.etcd.io/bbolt v1.3.10 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/term v0.25.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	google.golang.org/api v0.192.0 // indirect
	google.golang.org/genproto v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240930140551-af27646dc61f // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.1 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
	pgregory.net/rapid v1.1.0 // indirect
)

replace (
	// need to fix the comet.BlockInfo error, new changed to Info
	//cosmossdk.io/core/comet => cosmossdk.io/core/comet v0.11.0
	// Use the cosmos keyring code
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	// Use cometbft fork of tendermint
	//github.com/cometbft/cometbft => github.com/kava-labs/cometbft cometbft-patch-test // 8345af773eb9d2fe95656815662f28b0f3b64b46
	//github.com/cometbft/cometbft => github.com/kava-labs/cometbft v0.0.0-20241007151334-8345af773eb9
	github.com/cometbft/cometbft => github.com/kava-labs/cometbft v0.0.0-20241024200036-527d8df9ff12
	//github.com/cometbft/cometbft-db => github.com/kava-labs/cometbft-db kava-patched-test b2740b2e4bed1112feb4157b154ce969759b52ce
	github.com/cometbft/cometbft-db => github.com/kava-labs/cometbft-db v0.0.0-20241007145430-b2740b2e4bed
	// Use cosmos-sdk fork with backported fix for unsafe-reset-all, staking transfer events, and custom tally handler support
	//github.com/cosmos/cosmos-sdk => github.com/kava-labs/cosmos-sdk v0.47.10-iavl-v1-kava.1
	//github.com/cosmos/cosmos-sdk => github.com/kava-labs/cosmos-sdk v0.50.10-test-patch 5f9239e3147358ef034bfc4d19aacb34e5ea2064
	github.com/cosmos/cosmos-sdk => github.com/kava-labs/cosmos-sdk v0.0.0-20241112215550-dfd34308fe09
	//github.com/cosmos/cosmos-sdk => ../cosmos-sdk

	//github.com/cosmos/cosmos-sdk/store => cosmossdk.io/store v1.1.1
	github.com/cosmos/cosmos-sdk/x/evidence => cosmossdk.io/x/evidence v0.1.1
	github.com/cosmos/cosmos-sdk/x/feegrant => cosmossdk.io/x/feegrant v0.1.1
	github.com/cosmos/cosmos-sdk/x/nft => cosmossdk.io/x/nft v0.1.1
	github.com/cosmos/cosmos-sdk/x/upgrade => github.com/Kava-Labs/cosmos-sdk/x/upgrade v0.0.0-20241024201445-a8e1a4abf85b
	//cosmossdk.io/x/upgrade => github.com/Kava-Labs/cosmos-sdk/x/upgrade v0.0.0-20241024201445-a8e1a4abf85b
	// See https://github.com/cosmos/cosmos-sdk/pull/13093
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2
	// Tracking kava-labs/go-ethereum kava/release/v1.10 branch
	// TODO: Tag before release
	github.com/ethereum/go-ethereum => github.com/Kava-Labs/go-ethereum v1.10.27-0.20240513233504-6e038346780b
	// We can also to use wrap over EVM instead of changing this
	//github.com/ethereum/go-ethereum => github.com/Kava-Labs/go-ethereum v0.0.0-20241015215028-50a0b049fdc5
	// Use ethermint fork that respects min-gas-price with NoBaseFee true and london enabled, and includes eip712 support
	// Tracking kava-labs/etheremint master branch
	// TODO: Tag before release
	//github.com/evmos/ethermint => github.com/kava-labs/ethermint v0.21.1-0.20240802224012-586960857184
	github.com/evmos/ethermint => github.com/kava-labs/ethermint v0.0.0-20241112214305-986ef503477d
	//github.com/evmos/ethermint => ../ethermint
	// See https://github.com/cosmos/cosmos-sdk/pull/10401, https://github.com/cosmos/cosmos-sdk/commit/0592ba6158cd0bf49d894be1cef4faeec59e8320
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.0
	// Downgraded to avoid bugs in following commits which causes "version does not exist" errors
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
// Avoid change in slices.SortFunc, see https://github.com/cosmos/cosmos-sdk/issues/20159
// Was fixed
//golang.org/x/exp => golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb
)
