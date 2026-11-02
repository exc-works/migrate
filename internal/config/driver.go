package config

import (
	"errors"
	"net/url"
	"strings"

	"github.com/exc-works/migrate/internal/dialect"
)

// DriverDSN keeps Oracle reconnections on the full authentication path. go-ora
// v3.0.1's default fast-login cookie path fails on subsequent Oracle Free logins.
// An explicit driver option remains the operator's choice.
func DriverDSN(driverName, dsn string) (string, error) {
	if driverName != dialect.DriverOracle {
		return dsn, nil
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", errors.New("invalid Oracle DSN")
	}
	options, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", errors.New("invalid Oracle DSN options")
	}
	for key := range options {
		if strings.EqualFold(key, "FAST LOGIN") {
			return dsn, nil
		}
	}
	options.Set("FAST LOGIN", "false")
	parsed.RawQuery = options.Encode()
	return parsed.String(), nil
}
