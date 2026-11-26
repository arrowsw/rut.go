package rut

// IsRutLike reports whether rut has the shape of a RUT, without checking its
// verifier digit. It accepts three shapes: compact (123456785), compact with
// hyphen (12345678-5) and dotted (12.345.678-5), with optional surrounding
// whitespace and a k or K verifier. Leading zeros are rejected.
func IsRutLike(rut string) bool {
	_, _, ok := parseRutLike(rut)
	return ok
}
