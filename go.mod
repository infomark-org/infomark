module github.com/infomark-org/infomark

require (
	github.com/DATA-DOG/go-txdb v0.1.4
	github.com/alexedwards/scs v1.4.1
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/creasty/defaults v1.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v20.10.8+incompatible
	github.com/franela/goblin v0.0.0-20181003173013-ead4ad1d2727
	github.com/go-chi/chi/v5 v5.0.3
	github.com/go-chi/cors v1.2.0
	github.com/go-chi/jwtauth/v5 v5.0.1
	github.com/go-chi/render v1.0.1
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/go-redis/redis/v8 v8.4.2
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/uuid v1.1.2
	// Cannot be changed to v1.3 as this breaks the JOIN
	// https://github.com/jmoiron/sqlx/issues/755
	// https://github.com/jmoiron/sqlx/pull/754
	github.com/jmoiron/sqlx v1.2.0
	github.com/lib/pq v1.8.0
	github.com/markbates/pkger v0.15.1
	github.com/mattn/go-sqlite3 v1.14.8
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/streadway/amqp v1.0.0
	github.com/ulule/limiter/v3 v3.8.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	gopkg.in/guregu/null.v3 v3.5.0
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/v3 v3.0.3 // indirect
)

go 1.13
