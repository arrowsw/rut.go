package rut

// Mask hides the middle of a RUT for logs and screens, keeping only the
// leading group and the verifier: "12.345.678-5" gives "12.***.***-5".
//
// The input is normalized like Clean and returns ErrInvalidRut in the same
// cases; the Modulo 11 verifier is not checked.
//
// Masking is not anonymization: with the leading digits and the verifier
// kept, about 90,000 RUTs share each masked value, so treat masked RUTs as
// personal data.
func Mask(rut string) (string, error) {
	body, verifier, ok := parseLenient(rut)
	if !ok {
		return "", ErrInvalidRut
	}
	head := body[:len(body)-6] // 1 digit for a 7-digit body, 2 for an 8-digit one
	return head + ".***.***-" + verifier, nil
}
