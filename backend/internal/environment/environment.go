// Package environment provides methods to interact with environment variables.
package environment

import (
	"fmt"

	"github.com/caarlos0/env/v8"
	"github.com/joho/godotenv"
)

// Variables represents the environment variables used by the application. Please, make sure to update the .env.example
// file when modifying this structure.
//
//nolint:tagalign
type Variables struct {
	BaseURL  string `env:"SP_BASE_URL"     envDefault:"http://localhost:9798"`
	HTTPPort int    `env:"SP_HTTP_PORT"    envDefault:"9798"`

	BasicAuthUsername       string `env:"SP_BASIC_AUTH_USERNAME"        envDefault:"admin"`
	BasicAuthHashedPassword string `env:"SP_BASIC_AUTH_HASHED_PASSWORD" envDefault:"$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa"`

	CookieSecretKey string `env:"SP_COOKIE_SECRET_KEY" envDefault:"mflpw6hs4mdzads3s5kxjgtoltlgp3sp"`

	DBDSN      *string `env:"SP_DB_DSN"` // Optional: if set, overrides individual DB settings
	DBHost     string  `env:"SP_DB_HOST"         envDefault:"localhost"`
	DBPort     string  `env:"SP_DB_PORT"         envDefault:"5432"`
	DBName     string  `env:"SP_DB_NAME"         envDefault:"db"`
	DBUser     string  `env:"SP_DB_USER"         envDefault:"user"`
	DBPassword string  `env:"SP_DB_PASSWORD"     envDefault:"pass"`

	JWTSecretKey string `env:"SP_JWT_SECRET_KEY" envDefault:"q54isdosxiujnhjwmxrscqohr2tfm2c7"`
}

// Parse environment variables.
func Parse() *Variables {
	godotenv.Load()

	result := &Variables{}
	if err := env.Parse(result); err != nil {
		panic(fmt.Errorf("could not parse environment variables: %w", err))
	}

	return result
}

// DSN returns the database DSN. If SP_DB_DSN is explicitly set (not the default), it returns that value.
// Otherwise, it constructs the DSN from individual components (SP_DB_USER, SP_DB_PASSWORD, etc.).
func (v *Variables) DSN() string {
	// If DSN is explicitly set to a non-default value, use it
	if v.DBDSN != nil {
		return *v.DBDSN
	}
	// Otherwise, construct DSN from individual components
	return fmt.Sprintf("%s:%s@%s:%s/%s?sslmode=disable", v.DBUser, v.DBPassword, v.DBHost, v.DBPort, v.DBName)
}
