package rut

// maxInputLength is a security bound, not a format rule: oversized input is
// rejected before any parsing runs. As in rut.ts it is 64 UTF-16 code units,
// generously above any real formatted RUT.
const maxInputLength = 64

// isBounded reports whether s is non-empty and at most maxInputLength UTF-16
// code units long, which is how JavaScript measures a string's length. A rune
// outside the Basic Multilingual Plane counts as two units and each byte that
// is not valid UTF-8 counts as one.
func isBounded(s string) bool {
	if len(s) == 0 {
		return false
	}
	// Every UTF-16 unit needs at least one byte of UTF-8, so a string this
	// short cannot exceed the cap.
	if len(s) <= maxInputLength {
		return true
	}
	units := 0
	for _, r := range s {
		units++
		if r > 0xFFFF {
			units++
		}
		if units > maxInputLength {
			return false
		}
	}
	return true
}

// isJSSpace reports whether r is one of the 25 characters that JavaScript's
// String.prototype.trim() removes and its \s class matches: ASCII tab, line
// feed, vertical tab, form feed, carriage return and space, plus U+00A0,
// U+1680, U+2000–U+200A, U+2028, U+2029, U+202F, U+205F, U+3000 and U+FEFF.
//
// It differs from unicode.IsSpace, which also accepts U+0085 and rejects
// U+FEFF, and from the \s class of Go's regexp, which is ASCII only.
func isJSSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ',
		0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}
