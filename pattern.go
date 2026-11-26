package rut

// The canonical shapes accepted by Validate and IsRutLike, as regular
// expression source strings, for the places where code cannot run: a JSON
// Schema or OpenAPI pattern, a struct-tag validator, a database CHECK
// constraint. They are the same strings as rut.ts's RUT_PATTERN, byte for
// byte, and part of the frozen contract.
//
// A pattern proves shape only, never the verifier digit: pair it with
// Validate wherever code can run. PatternAny and PatternCompact match any
// 9-digit number, such as a phone number, so prefer PatternHyphen or
// PatternDotted in API schemas.
//
// The strings are anchored, flagless and backslash-free, so they can be
// pasted into JSON, YAML or SQL as they are, and they compile with Go's
// regexp package. For inputs up to the 64-unit cap,
// IsRutLike(s) == regexp.MustCompile(PatternAny).MatchString(trimmed s),
// where trimming removes the JavaScript whitespace described in the package
// documentation.
const (
	// PatternAny matches any of the three shapes below.
	PatternAny = "^([1-9][0-9]{6,7}-?[0-9kK]|[1-9][0-9]?[.][0-9]{3}[.][0-9]{3}-[0-9kK])$"
	// PatternCompact matches 123456785.
	PatternCompact = "^[1-9][0-9]{6,7}[0-9kK]$"
	// PatternHyphen matches 12345678-5.
	PatternHyphen = "^[1-9][0-9]{6,7}-[0-9kK]$"
	// PatternDotted matches 12.345.678-5.
	PatternDotted = "^[1-9][0-9]?[.][0-9]{3}[.][0-9]{3}-[0-9kK]$"
)
