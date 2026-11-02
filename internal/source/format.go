package source

// Migration formats and extensions shared by discovery, runtimes and CLI generation.
const (
	FormatSQL = "sql"
	FormatJS  = "js"
	FormatTS  = "ts"
	FormatMJS = "mjs"

	ExtensionSQL = "." + FormatSQL
	ExtensionJS  = "." + FormatJS
	ExtensionTS  = "." + FormatTS
	ExtensionMJS = "." + FormatMJS
)
