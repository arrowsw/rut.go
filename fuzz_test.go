package rut

import (
	"regexp"
	"strings"
	"testing"
)

// The golden fixture pins the answers of rut.ts; these fuzz tests check the
// rules that tie the functions together on any input, invalid UTF-8 included,
// which rut.ts cannot produce. The seeds run on every `go test`; explore with
//
//	go test -fuzz=FuzzInvariants -fuzztime=1m

var (
	cleanShape  = regexp.MustCompile(`^[1-9][0-9]{6,7}[0-9K]$`)
	dottedOut   = regexp.MustCompile(`^[1-9][0-9]?\.[0-9]{3}\.[0-9]{3}-[0-9K]$`)
	maskedShape = regexp.MustCompile(`^[1-9][0-9]?\.\*\*\*\.\*\*\*-[0-9K]$`)
	anyPattern  = regexp.MustCompile(PatternAny)
)

func fuzzSeeds(f *testing.F) {
	for _, s := range []string{
		"", " ", "12.345.678-5", "12345678-5", "123456785", "9.068.826-k", "009.068.826-K",
		"11.111.111-1", "12.345.678-9", "RUT: 12.345.678-5", "\ufeff12.345.678-5\u3000",
		"\u008512345678-5", "12\u00a0345\u00a0678", "1234\xff5678", "\xed\xa0\x80123456785",
		strings.Repeat("\u00a0", 55) + "123456785", strings.Repeat("0", 56) + "123456785",
		"１２３４５６７８-５", "12345678--5", "1-9", "100000000-7",
	} {
		f.Add(s)
	}
}

func FuzzInvariants(f *testing.F) {
	fuzzSeeds(f)
	f.Fuzz(func(t *testing.T, s string) {
		cleaned, cleanErr := Clean(s)
		body, bodyErr := GetBody(s)
		verifier, verifierErr := GetVerifier(s)
		d, decomposeErr := Decompose(s)
		masked, maskErr := Mask(s)
		formatted, formatErr := Format(s)
		plain, plainErr := Format(s, WithDots(false))
		cv, cvErr := CalculateVerifier(s)
		valid := Validate(s)
		like := IsRutLike(s)

		// Every failure is the bare sentinel, so no input reaches an error.
		for _, err := range []error{cleanErr, bodyErr, verifierErr, decomposeErr, maskErr, formatErr, plainErr, cvErr} {
			if err != nil && err != ErrInvalidRut {
				t.Fatalf("unexpected error %v", err)
			}
		}

		// The lenient helpers succeed and fail together and agree.
		if (cleanErr == nil) != (bodyErr == nil) || (cleanErr == nil) != (verifierErr == nil) ||
			(cleanErr == nil) != (decomposeErr == nil) || (cleanErr == nil) != (maskErr == nil) {
			t.Fatalf("lenient helpers disagree on %q", s)
		}
		if cleanErr == nil {
			if !cleanShape.MatchString(cleaned) {
				t.Fatalf("Clean(%q) = %q, not a compact RUT shape", s, cleaned)
			}
			if again, err := Clean(cleaned); err != nil || again != cleaned {
				t.Fatalf("Clean is not idempotent on %q", s)
			}
			if body+verifier != cleaned || d.Body != body || d.Verifier != verifier {
				t.Fatalf("GetBody/GetVerifier/Decompose disagree with Clean on %q", s)
			}
			if !maskedShape.MatchString(masked) {
				t.Fatalf("Mask(%q) = %q", s, masked)
			}
		}

		// Format succeeds exactly when the cleaned value passes Modulo 11, except
		// for the empty string, which gives "".
		cleanValid := cleanErr == nil && calculateVerifierForBody(body) == verifier
		if s == "" {
			if formatted != "" || formatErr != nil {
				t.Fatalf("Format(\"\") = %q, %v", formatted, formatErr)
			}
		} else if (formatErr == nil) != cleanValid || (plainErr == nil) != cleanValid {
			t.Fatalf("Format(%q) = %q, %v; clean valid = %v", s, formatted, formatErr, cleanValid)
		}
		if formatErr == nil && s != "" {
			if !dottedOut.MatchString(formatted) || !Validate(formatted) || strings.ReplaceAll(formatted, ".", "") != plain {
				t.Fatalf("Format(%q) = %q / %q", s, formatted, plain)
			}
		}

		// Validate is IsRutLike plus Modulo 11, and implies every lenient helper.
		if valid && (!like || !cleanValid) {
			t.Fatalf("Validate(%q) = true but IsRutLike = %v, clean valid = %v", s, like, cleanValid)
		}
		if like && cleanErr != nil {
			t.Fatalf("IsRutLike(%q) = true but Clean fails", s)
		}
		if Validate(s, WithStrict(true)) && !valid {
			t.Fatalf("strict accepts %q but the default rejects it", s)
		}

		// PatternAny matches exactly what IsRutLike accepts, within the cap.
		if isBounded(s) && anyPattern.MatchString(strings.TrimFunc(s, isJSSpace)) != like {
			t.Fatalf("PatternAny and IsRutLike disagree on %q", s)
		}

		// Equals(s, s) holds exactly for a cleanable value that passes Modulo 11.
		if Equals(s, s) != cleanValid || Equals(s, s, WithRequireValid(false)) != (cleanErr == nil) {
			t.Fatalf("Equals(%q, %q) is inconsistent with Clean", s, s)
		}

		// CalculateVerifier agrees with the verifier of any valid compact RUT.
		if cvErr == nil && len(cv) != 1 {
			t.Fatalf("CalculateVerifier(%q) = %q", s, cv)
		}
		if valid {
			if v, err := CalculateVerifier(body); err != nil || v != verifier {
				t.Fatalf("CalculateVerifier(GetBody(%q)) = %q, %v; want %q", s, v, err, verifier)
			}
		}
	})
}

func FuzzEqualsIsSymmetric(f *testing.F) {
	f.Add("12.345.678-5", "123456785")
	f.Add("012345678-5", "12.345.678-5")
	f.Add("12345678-9", "123456789")
	f.Add("nope", "12.345.678-5")
	f.Fuzz(func(t *testing.T, a, b string) {
		if Equals(a, b) != Equals(b, a) || Equals(a, b, WithRequireValid(false)) != Equals(b, a, WithRequireValid(false)) {
			t.Fatalf("Equals(%q, %q) is not symmetric", a, b)
		}
		if Equals(a, b) && !Equals(a, b, WithRequireValid(false)) {
			t.Fatalf("Equals(%q, %q) holds only with the stricter default", a, b)
		}
	})
}
