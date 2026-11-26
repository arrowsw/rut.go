package rut

// EqualsOption configures Equals.
type EqualsOption func(*equalsOptions)

type equalsOptions struct {
	requireValid bool
}

// WithRequireValid chooses whether Equals also requires the shared value to
// pass Modulo 11. It is true by default. Pass false for the pure shape
// comparison, for example to deduplicate a dirty dataset where the same typo
// twice is one entity.
func WithRequireValid(requireValid bool) EqualsOption {
	return func(o *equalsOptions) { o.requireValid = requireValid }
}

// Equals reports whether a and b are the same RUT: both must normalize to the
// same value and, by default, that value must pass Modulo 11.
//
// Dots, hyphens, leading zeros, embedded garbage and verifier case are ignored:
// Equals("12.345.678-5", "123456785") is true. Validity is checked on the
// normalized value, not on Validate's canonical shape, so
// Equals("012345678-5", "12.345.678-5") is true even though
// Validate("012345678-5") is false. Two strings with a wrong verifier are not
// the same RUT, because they are not RUTs: Equals("12345678-9", "12345678-9")
// is false unless WithRequireValid(false) is given.
//
// It returns false when either argument cannot be normalized to a RUT shape.
func Equals(a, b string, opts ...EqualsOption) bool {
	o := equalsOptions{requireValid: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	bodyA, verifierA, ok := parseLenient(a)
	if !ok {
		return false
	}
	bodyB, verifierB, ok := parseLenient(b)
	if !ok || bodyA != bodyB || verifierA != verifierB {
		return false
	}
	return !o.requireValid || calculateVerifierForBody(bodyA) == verifierA
}
