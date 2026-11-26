package rut

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
