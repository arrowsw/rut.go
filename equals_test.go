package rut

import "testing"

type pair struct{ name, a, b string }

func TestEqualsMatchesTheSameValidRutAcrossSpellings(t *testing.T) {
	cases := []pair{
		{"dotted vs compact", "12.345.678-5", "123456785"},
		{"dotted vs hyphen", "12.345.678-5", "12345678-5"},
		{"compact vs hyphen", "123456785", "12345678-5"},
		{"whitespace", "12.345.678-5", "  12.345.678-5  "},
		{"verifier case", "14.625.621-K", "14.625.621-k"},
		{"leading zeros", "9.068.826-K", "0009068826K"},
		{"zero-padded vs canonical", "012345678-5", "12.345.678-5"},
		{"identical", "12.345.678-5", "12.345.678-5"},
	}
	for _, c := range cases {
		for _, opts := range [][]EqualsOption{nil, {WithRequireValid(true)}, {WithRequireValid(false)}} {
			if !Equals(c.a, c.b, opts...) {
				t.Errorf("%s: Equals(%q, %q, %d options) = false, want true", c.name, c.a, c.b, len(opts))
			}
		}
	}
}

func TestEqualsRequiresAValidRutByDefault(t *testing.T) {
	cases := []pair{
		{"identical strings, wrong verifier", "12345678-9", "12345678-9"},
		{"same wrong value across shapes", "12.345.678-9", "123456789"},
		{"zero-padded wrong verifier", "0012345678-9", "12345678-9"},
	}
	for _, c := range cases {
		if Equals(c.a, c.b) || Equals(c.a, c.b, WithRequireValid(true)) {
			t.Errorf("%s: Equals(%q, %q) = true, want false", c.name, c.a, c.b)
		}
		// The dedup use case: the same typo twice is still one entity.
		if !Equals(c.a, c.b, WithRequireValid(false)) {
			t.Errorf("%s: Equals(%q, %q, WithRequireValid(false)) = false, want true", c.name, c.a, c.b)
		}
	}
}

func TestEqualsRejectsDifferentOrUnparseableRuts(t *testing.T) {
	cases := []pair{
		{"different bodies", "12.345.678-5", "1.234.567-4"},
		{"same body, different verifier", "12345678-5", "12345678-9"},
		{"unrelated", "12.345.678-5", "98.765.432-1"},
		{"second invalid", "12.345.678-5", "nope"},
		{"first invalid", "nope", "12.345.678-5"},
		{"both invalid", "nope", "also-nope"},
		{"both empty", "", ""},
	}
	for _, c := range cases {
		if Equals(c.a, c.b) || Equals(c.a, c.b, WithRequireValid(false)) {
			t.Errorf("%s: Equals(%q, %q) = true, want false", c.name, c.a, c.b)
		}
	}
}

func TestEqualsChecksValidityOnTheNormalizedValue(t *testing.T) {
	// Validate rejects the zero-padded shape, Equals accepts it: Equals is a
	// normalizing comparison and only adds the Modulo 11 check.
	if Validate("012345678-5") {
		t.Fatal("Validate(\"012345678-5\") = true, want false")
	}
	if !Equals("012345678-5", "12.345.678-5") {
		t.Error("Equals(\"012345678-5\", \"12.345.678-5\") = false, want true")
	}
}

func TestEqualsIsSymmetric(t *testing.T) {
	for _, c := range []pair{
		{"", "12.345.678-5", "123456785"},
		{"", "12345678-9", "123456789"},
		{"", "012345678-5", "12.345.678-5"},
		{"", "nope", "12.345.678-5"},
	} {
		for _, opts := range [][]EqualsOption{nil, {WithRequireValid(false)}} {
			if Equals(c.a, c.b, opts...) != Equals(c.b, c.a, opts...) {
				t.Errorf("Equals(%q, %q) != Equals(%q, %q)", c.a, c.b, c.b, c.a)
			}
		}
	}
}

func TestEqualsIgnoresNilOptions(t *testing.T) {
	if Equals("12345678-9", "12345678-9", nil) {
		t.Error("a nil option must keep the default WithRequireValid(true)")
	}
}
