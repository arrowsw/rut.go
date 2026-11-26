package rut

// Clean normalizes a RUT to its compact form: "12.345.678-k" becomes
// "12345678K".
//
// It removes every character outside [0-9kK] wherever it sits, including
// embedded text ("RUT: 12.345.678-5" gives "123456785"), drops leading zeros
// and uppercases k. The result must have a 7- or 8-digit body followed by one
// verifier character, otherwise Clean returns ErrInvalidRut.
//
// Clean does not check the Modulo 11 verifier.
func Clean(rut string) (string, error) {
	body, verifier, ok := parseLenient(rut)
	if !ok {
		return "", ErrInvalidRut
	}
	return body + verifier, nil
}
