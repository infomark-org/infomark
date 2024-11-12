module github.com/infomark-org/infomark

require (
	github.com/DATA-DOG/go-txdb v0.1.4
	github.com/alexedwards/scs v1.4.1
	github.com/coreos/go-semver v0.3.1
	github.com/creasty/defaults v1.6.0
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v24.0.9+incompatible
	github.com/franela/goblin v0.0.0-20211003143422-0a4f594942bf
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/cors v1.2.1
	github.com/go-chi/jwtauth/v5 v5.1.0
	github.com/go-chi/render v1.0.2
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/uuid v1.3.0
	// Cannot be changed to v1.3 as this breaks the JOIN
	// https://github.com/jmoiron/sqlx/issues/755
	// https://github.com/jmoiron/sqlx/pull/754
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.10.7
	github.com/markbates/pkger v0.17.1
	github.com/mattn/go-sqlite3 v1.14.10
	github.com/prometheus/client_golang v1.14.0
	github.com/redis/go-redis/v9 v9.0.2
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/streadway/amqp v1.0.0
	github.com/ulule/limiter/v3 v3.11.0
	golang.org/x/crypto v0.21.0
	gopkg.in/guregu/null.v3 v3.5.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dhui/dktest v0.3.10 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gobuffalo/here v0.6.7 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.5 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.0.21 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.40.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.1.0 // indirect
)

go 1.19
