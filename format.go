package rut

// FormatOption configures Format.
type FormatOption func(*formatOptions)

type formatOptions struct {
	dots bool
}

// WithDots chooses whether Format groups the body with dots
// (12.345.678-5, the default) or not (12345678-5).
func WithDots(dots bool) FormatOption {
	return func(o *formatOptions) { o.dots = dots }
}

// Format renders a valid RUT in canonical form: "123456785" gives
// "12.345.678-5", or "12345678-5" with WithDots(false).
//
// The input is normalized like Clean (separators, garbage, leading zeros and
// case are ignored), and then the Modulo 11 verifier must be right: a wrong
// verifier returns ErrInvalidRut instead of being repaired.
//
// The empty string gives "" and no error, as in rut.ts: check for "" before
// treating the result as a RUT.
//
// rut.ts's format-as-you-type mode (incremental: true) is not ported.
func Format(rut string, opts ...FormatOption) (string, error) {
	o := formatOptions{dots: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if rut == "" {
		return "", nil
	}
	body, verifier, ok := parseLenient(rut)
	if !ok || calculateVerifierForBody(body) != verifier {
		return "", ErrInvalidRut
	}
	return formatBody(body, o.dots) + "-" + verifier, nil
}

// formatBody groups a digit-only body in thousands from the right
// (12345678 → 12.345.678), or returns it untouched when dots is false.
func formatBody(body string, dots bool) string {
	if !dots {
		return body
	}
	out := make([]byte, 0, len(body)+len(body)/3)
	for i := 0; i < len(body); i++ {
		if i > 0 && (len(body)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, body[i])
	}
	return string(out)
}
