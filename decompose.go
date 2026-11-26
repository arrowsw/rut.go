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
