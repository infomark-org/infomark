module github.com/infomark-org/infomark

require (
	github.com/DATA-DOG/go-txdb v0.1.4
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/alexedwards/scs v1.4.1
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/coreos/go-semver v0.3.1
	github.com/creasty/defaults v1.6.0
	github.com/davecgh/go-spew v1.1.1
	github.com/dhui/dktest v0.3.10 // indirect
	github.com/docker/docker v23.0.1+incompatible
	github.com/docker/go-units v0.5.0 // indirect
	github.com/franela/goblin v0.0.0-20211003143422-0a4f594942bf
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/cors v1.2.1
	github.com/go-chi/jwtauth/v5 v5.1.0
	github.com/go-chi/render v1.0.2
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/gobuffalo/here v0.6.7 // indirect
	github.com/goccy/go-json v0.10.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/uuid v1.3.0
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	// Cannot be changed to v1.3 as this breaks the JOIN
	// https://github.com/jmoiron/sqlx/issues/755
	// https://github.com/jmoiron/sqlx/pull/754
	github.com/jmoiron/sqlx v1.2.0
	github.com/lestrrat-go/jwx/v2 v2.0.8 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lib/pq v1.10.7
	github.com/markbates/pkger v0.17.1
	github.com/mattn/go-sqlite3 v1.14.10
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.40.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/redis/go-redis/v9 v9.0.2
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.6.1
	github.com/streadway/amqp v1.0.0
	github.com/ulule/limiter/v3 v3.11.0
	golang.org/x/crypto v0.6.0
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	gopkg.in/guregu/null.v3 v3.5.0
	gopkg.in/yaml.v2 v2.4.0
)

go 1.13
