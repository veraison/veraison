module github.com/veraison/veraison/provisioning/plugins/corim-psa

replace (
	github.com/veraison/common => ../../../common
	github.com/veraison/endorsement => ../../../endorsement
	github.com/veraison/veraison/provisioning/decoder => ../../decoder
	github.com/veraison/veraison/provisioning/plugins/common => ../common
)

go 1.17

require (
	github.com/hashicorp/go-plugin v1.4.3
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	github.com/veraison/corim v0.0.0-20211208184228-467df6558f2c
	github.com/veraison/eat v0.0.0-20210331113810-3da8a4dd42ff
	github.com/veraison/endorsement v0.0.0-00010101000000-000000000000
	github.com/veraison/veraison/provisioning/decoder v0.0.0-00010101000000-000000000000
	github.com/veraison/veraison/provisioning/plugins/common v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.0-20210816181553-5444fa50b93d // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/fxamacker/cbor/v2 v2.3.0 // indirect
	github.com/goccy/go-json v0.7.6 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.0 // indirect
	github.com/lestrrat-go/iter v1.0.1 // indirect
	github.com/lestrrat-go/jwx v1.2.6 // indirect
	github.com/lestrrat-go/option v1.0.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/ohler55/ojg v1.2.1 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.9.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/veraison/go-cose v0.0.0-20201125131510-de93f6091ed4 // indirect
	github.com/veraison/swid v0.0.1-beta.6 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420 // indirect
	golang.org/x/sys v0.0.0-20210927094055-39ccf1dd6fa6 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
