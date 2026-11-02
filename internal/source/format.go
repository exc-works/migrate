package source

// Migration formats and extensions shared by discovery, runtimes and CLI generation.
const (
	FormatSQL = "sql"
	FormatJS  = "js"
	FormatTS  = "ts"

	ExtensionSQL = "." + FormatSQL
	ExtensionJS  = "." + FormatJS
	ExtensionTS  = "." + FormatTS
)
