package rut

const (
	minBodyLength = 7
	maxBodyLength = 8
)

// CalculateVerifier returns the verifier digit ("0"–"9" or "K") for a RUT
// body.
//
// The body may contain dots, hyphens and whitespace anywhere, and leading
// zeros; they are removed first. What remains must be 7 or 8 digits
// (1000000–99999999), otherwise it returns ErrInvalidRut.
func CalculateVerifier(body string) (string, error) {
	cleaned, ok := normalizeRutBody(body)
	if !ok {
		return "", ErrInvalidRut
	}
	return calculateVerifierForBody(cleaned), nil
}

// normalizeRutBody removes dots, hyphens and JavaScript whitespace from body,
// then its leading zeros, and checks that 7 or 8 digits remain.
func normalizeRutBody(body string) (string, bool) {
	if !isBounded(body) {
		return "", false
	}
	digits := make([]byte, 0, len(body))
	for _, r := range body {
		switch {
		case r == '.' || r == '-' || isJSSpace(r):
			continue
		case r >= '0' && r <= '9':
			if len(digits) == 0 && r == '0' {
				continue // leading zero
			}
			digits = append(digits, byte(r))
		default:
			return "", false
		}
	}
	if len(digits) < minBodyLength || len(digits) > maxBodyLength {
		return "", false
	}
	return string(digits), true
}

// verifierByCheckDigit maps 11 - (sum % 11), always in 1..11, to its verifier.
const verifierByCheckDigit = "?123456789K0"

// calculateVerifierForBody runs Modulo 11 over a string of ASCII digits.
func calculateVerifierForBody(body string) string {
	sum, multiplier := 0, 2
	for i := len(body) - 1; i >= 0; i-- {
		sum += int(body[i]-'0') * multiplier
		if multiplier == 7 {
			multiplier = 2
		} else {
			multiplier++
		}
	}
	check := 11 - sum%11
	return verifierByCheckDigit[check : check+1]
}
