package rut

// GetBody returns the body of a RUT, the digits before the verifier:
// "12.345.678-5" gives "12345678".
//
// It normalizes the input exactly like Clean and returns ErrInvalidRut in the
// same cases. It does not check the Modulo 11 verifier.
func GetBody(rut string) (string, error) {
	body, _, ok := parseLenient(rut)
	if !ok {
		return "", ErrInvalidRut
	}
	return body, nil
}

// GetVerifier returns the verifier of a RUT, its last character, always
// uppercase: "9.068.826-k" gives "K".
//
// It normalizes the input exactly like Clean and returns ErrInvalidRut in the
// same cases. It does not check that the verifier is the right one.
func GetVerifier(rut string) (string, error) {
	_, verifier, ok := parseLenient(rut)
	if !ok {
		return "", ErrInvalidRut
	}
	return verifier, nil
}

// DecomposedRut is a RUT split into its body and its verifier.
type DecomposedRut struct {
	Body     string // 7 or 8 digits, without leading zeros
	Verifier string // "0"–"9" or "K"
}

// Decompose splits a RUT into body and verifier: "12.345.678-5" gives
// DecomposedRut{Body: "12345678", Verifier: "5"}.
//
// It normalizes the input exactly like Clean and returns ErrInvalidRut, with
// the zero DecomposedRut, in the same cases. It does not check the Modulo 11
// verifier.
func Decompose(rut string) (DecomposedRut, error) {
	body, verifier, ok := parseLenient(rut)
	if !ok {
		return DecomposedRut{}, ErrInvalidRut
	}
	return DecomposedRut{Body: body, Verifier: verifier}, nil
}
