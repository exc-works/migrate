package script

import "github.com/exc-works/migrate/internal/dialect"

const (
	VerificationVerified     = "verified"
	VerificationExperimental = "experimental"
	TransactionFramework     = "framework"
	TransactionNone          = "none"
)

// DialectCapability separates driver verification from transaction behavior.
// Framework transactions cannot undo implicitly committing DDL.
type DialectCapability struct {
	Name            string `json:"name"`
	Verification    string `json:"verification"`
	TransactionMode string `json:"transaction_mode"`
	Notes           string `json:"notes,omitempty"`
}

// Dialects returns a fresh list of executable script dialects. Only full builds
// expose these capabilities to users; SQL-only execution still fails preflight.
func Dialects() []DialectCapability {
	return []DialectCapability{
		{
			Name:            dialect.NamePostgres,
			Verification:    VerificationVerified,
			TransactionMode: TransactionFramework,
		},
		{
			Name:            dialect.NameMySQL,
			Verification:    VerificationVerified,
			TransactionMode: TransactionFramework,
			Notes:           "requires transactional tables; DDL may implicitly commit",
		},
		{
			Name:            dialect.NameMariaDB,
			Verification:    VerificationVerified,
			TransactionMode: TransactionFramework,
			Notes:           "requires transactional tables; DDL may implicitly commit",
		},
		{
			Name:            dialect.NameSQLite,
			Verification:    VerificationVerified,
			TransactionMode: TransactionFramework,
		},
		{
			Name:            dialect.NameTiDB,
			Verification:    VerificationExperimental,
			TransactionMode: TransactionFramework,
			Notes:           "not verified against a real database; DDL may implicitly commit",
		},
		{
			Name:            dialect.NameRedshift,
			Verification:    VerificationExperimental,
			TransactionMode: TransactionFramework,
			Notes:           "not verified against a real database; transaction behavior depends on the statement",
		},
		{
			Name:            dialect.NameOracle,
			Verification:    VerificationExperimental,
			TransactionMode: TransactionFramework,
			Notes:           "not verified against a real database; DDL implicitly commits; PL/SQL blocks and procedure calls require SQL migrations",
		},
	}
}

func Capability(name string) (DialectCapability, bool) {
	for _, capability := range Dialects() {
		if capability.Name == name {
			return capability, true
		}
	}
	return DialectCapability{}, false
}

func SupportedDialect(name string) bool {
	_, ok := Capability(name)
	return ok
}
