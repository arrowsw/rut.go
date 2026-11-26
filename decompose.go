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
