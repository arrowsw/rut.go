package rut

// IsRutLike reports whether rut has the shape of a RUT, without checking its
// verifier digit. It accepts the same three shapes as Validate: compact
// (123456785), compact with hyphen (12345678-5) and dotted (12.345.678-5),
// with optional surrounding whitespace and a k or K verifier. Leading zeros
// are rejected.
func IsRutLike(rut string) bool {
	_, _, ok := parseRutLike(rut)
	return ok
}

// ValidateOption configures Validate.
type ValidateOption func(*validateOptions)

type validateOptions struct {
	strict bool
}

// WithStrict, when true, makes Validate also reject repeated-digit
// placeholders such as 11.111.111-1 or 8.888.888-K, which pass Modulo 11 but
// are rarely real. It is false by default.
func WithStrict(strict bool) ValidateOption {
	return func(o *validateOptions) { o.strict = strict }
}

// Validate reports whether rut is a valid RUT: one of the canonical shapes
// accepted by IsRutLike, with the right Modulo 11 verifier.
//
// Leading-zero padding (012.345.678-5) is rejected, because a real RUT has
// none; normalize zero-padded values with Clean first if you need to accept
// them. "Surrounding whitespace" is whatever JavaScript's trim() removes,
// Unicode spaces included (U+00A0, U+FEFF, U+3000…), so one RUT can be
// written in several accepted ways: Validate answers "is this a valid RUT?",
// not "is this the one spelling of it". To store or compare RUTs, use the
// output of Clean.
func Validate(rut string, opts ...ValidateOption) bool {
	var o validateOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	body, verifier, ok := parseRutLike(rut)
	if !ok {
		return false
	}
	if o.strict && isSuspicious(body) {
		return false
	}
	return calculateVerifierForBody(body) == verifier
}

// isSuspicious reports whether every digit of body is the same.
func isSuspicious(body string) bool {
	for i := 1; i < len(body); i++ {
		if body[i] != body[0] {
			return false
		}
	}
	return true
}
