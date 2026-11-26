package rut

import (
	"regexp"
	"strings"
)

const (
	minRutLength = 8 // 7-digit body + verifier
	maxRutLength = 9 // 8-digit body + verifier
)

// parseLenient is the normalization-first parser shared by the recovery
// helpers (Clean, GetBody, GetVerifier, Decompose, Mask, Format). It strips
// separators, embedded garbage, leading zeros and case before splitting, and
// deliberately does not check the Modulo 11 verifier: these helpers recover
// RUT-shaped values from messy input, they do not assert validity.
func parseLenient(s string) (body, verifier string, ok bool) {
	if !isBounded(s) {
		return "", "", false
	}
	return splitCleanRut(normalizeRutValue(s))
}

// normalizeRutValue keeps only [0-9kK], drops the leading zeros of what is
// left and uppercases k. Working byte by byte is safe: no byte of a multi-byte
// UTF-8 sequence is ASCII, so non-ASCII text is removed whole, just as
// JavaScript removes it unit by unit.
func normalizeRutValue(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '0' && len(out) == 0:
			// leading zero
		case c >= '0' && c <= '9':
			out = append(out, c)
		case c == 'k' || c == 'K':
			out = append(out, 'K')
		}
	}
	return string(out)
}

// splitCleanRut checks the length of a normalized value and splits it into
// body and verifier. A total of 8–9 characters already implies a 7–8 digit
// body.
func splitCleanRut(cleaned string) (body, verifier string, ok bool) {
	if len(cleaned) < minRutLength || len(cleaned) > maxRutLength {
		return "", "", false
	}
	body, verifier = cleaned[:len(cleaned)-1], cleaned[len(cleaned)-1:]
	if !isDigits(body) {
		return "", "", false
	}
	// normalizeRutValue only keeps digits and K, so the verifier is always
	// one of "0"–"9" or "K" here.
	return body, verifier, true
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

// rutShape is the union of the three canonical shapes: compact (123456785),
// compact with hyphen (12345678-5) and dotted (12.345.678-5). It allows a
// leading zero on purpose; parseRutLike rejects it explicitly.
var rutShape = regexp.MustCompile(`^(?:[0-9]{7,8}-?[0-9kK]|[0-9]{1,3}\.[0-9]{3}\.[0-9]{3}-[0-9kK])$`)

// parseRutLike is the canonical parser used by IsRutLike. After trimming
// JavaScript whitespace the input must be one of the three canonical shapes,
// without leading zeros.
func parseRutLike(s string) (body, verifier string, ok bool) {
	if !isBounded(s) {
		return "", "", false
	}
	input := strings.TrimFunc(s, isJSSpace)
	if input == "" || !rutShape.MatchString(input) {
		return "", "", false
	}
	// Reject leading-zero padding: a canonical RUT has none, and padding forms
	// an unbounded family of strings for one RUT (12345678 = 012345678 = …)
	// that must not slip past an identity check. Every accepted shape starts
	// with the body's top digit, so a leading '0' is always padding.
	if input[0] == '0' {
		return "", "", false
	}
	return splitCleanRut(normalizeRutValue(input))
}
