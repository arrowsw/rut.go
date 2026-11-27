package rut

import (
	"strings"
	"testing"
)

func TestIsRutLikeAcceptsTheThreeShapes(t *testing.T) {
	cases := []struct{ name, in string }{
		{"dotted", "12.345.678-9"},
		{"compact with hyphen", "12345678-9"},
		{"compact", "123456789"},
		{"dotted 7-digit body, K verifier", "1.234.567-K"},
		{"surrounding whitespace", "  12.345.678-5  "},
		{"lowercase k", "12.345.678-k"},
		{"smallest 7-digit body", "1000000-9"},
		{"largest 8-digit body", "99999999-9"},
		{"BOM U+FEFF is trimmed", "\ufeff12.345.678-5"},
		{"ideographic space U+3000 is trimmed", "12.345.678-5\u3000"},
		{"no-break spaces U+00A0 are trimmed", "\u00a012345678-5\u00a0"},
		{"line separator U+2028 is trimmed", "\u202812345678-5"},
	}
	for _, c := range cases {
		if !IsRutLike(c.in) {
			t.Errorf("%s: IsRutLike(%q) = false, want true", c.name, c.in)
		}
	}
}

func TestIsRutLikeRejectsOtherShapes(t *testing.T) {
	cases := []struct{ name, in string }{
		{"letters", "abcdefghi"},
		{"empty", ""},
		{"whitespace only", " \t\n"},
		{"wrong dot grouping", "12.34.56-7"},
		{"non-canonical grouping", "12.345678-5"},
		{"non-canonical grouping", "12345.678-5"},
		{"dotted without hyphen", "12.345.6785"},
		{"over the 64-unit cap", strings.Repeat("1", 128)},
		{"internal whitespace", "12 345 678 5"},
		{"K not at the end", "K2345678-5"},
		{"leading zeros, dotted", "009.068.826-K"},
		{"leading zeros, compact", "00012345678"},
		{"single leading zero on a 7-digit body", "01234567-4"},
		{"6-digit body, hyphen", "999999-K"},
		{"6-digit body, dotted", "999.999-K"},
		{"6-digit body, compact", "999999K"},
		{"1-digit body", "1-9"},
		{"9-digit body, hyphen", "100000000-7"},
		{"9-digit body, dotted", "123.456.789-5"},
		{"NEL U+0085 is not trimmed (unicode.IsSpace would trim it)", "\u008512345678-5"},
		{"zero-width space is not trimmed", "12345678-5\u200b"},
		{"Mongolian vowel separator is not trimmed", "\u180e12345678-5"},
		{"an inner newline is not trimmed", "12345678\n-5"},
		{"invalid UTF-8 is not trimmed", "\xff12345678-5"},
	}
	for _, c := range cases {
		if IsRutLike(c.in) {
			t.Errorf("%s: IsRutLike(%q) = true, want false", c.name, c.in)
		}
	}
}

func TestIsRutLikeDoesNotCheckTheVerifier(t *testing.T) {
	if !IsRutLike("12.345.678-0") {
		t.Error("IsRutLike(\"12.345.678-0\") = false, want true: the shape is right")
	}
}

func TestIsRutLikeCapsInputAt64UTF16Units(t *testing.T) {
	for _, pad := range []string{" ", "\u00a0", "\u3000", "\ufeff"} {
		atCap := strings.Repeat(pad, 55) + "123456785"
		if !IsRutLike(atCap) {
			t.Errorf("64 units padded with %q: IsRutLike = false, want true", pad)
		}
		if IsRutLike(pad + atCap) {
			t.Errorf("65 units padded with %q: IsRutLike = true, want false", pad)
		}
	}
}

func TestValidateAcceptsValidRuts(t *testing.T) {
	cases := []struct{ name, in string }{
		{"dotted", "12.345.678-5"},
		{"compact with hyphen", "12345678-5"},
		{"compact", "123456785"},
		{"dotted, 7-digit body", "1.234.567-4"},
		{"compact with hyphen, 7-digit body", "1234567-4"},
		{"compact, 7-digit body", "12345674"},
		{"uppercase K", "14.625.621-K"},
		{"lowercase k", "14.625.621-k"},
		{"compact uppercase K", "14625621K"},
		{"compact lowercase k", "14625621k"},
		{"verifier 0", "18.264.950-3"},
		{"verifier 0 (check digit 11)", "10.000.004-0"},
		{"minimum 8-digit body", "10.000.000-8"},
		{"floor of the range", "1000000-9"},
		{"ceiling of the range", "99.999.999-9"},
		{"surrounding spaces", "  12.345.678-5  "},
		{"tab and newline", "\t12.345.678-5\n"},
		{"compact with surrounding spaces", "  123456785  "},
		{"K with surrounding whitespace", "\n14.625.621-K\t"},
		{"BOM U+FEFF", "\ufeff12.345.678-5"},
		{"ideographic space U+3000", "12.345.678-5\u3000"},
		{"no-break spaces U+00A0", "\u00a012345678-5\u00a0"},
		{"whitespace padding up to 64 units", strings.Repeat(" ", 55) + "123456785"},
	}
	for _, c := range cases {
		if !Validate(c.in) {
			t.Errorf("%s: Validate(%q) = false, want true", c.name, c.in)
		}
	}
}

func TestValidateRejectsInvalidRuts(t *testing.T) {
	cases := []struct{ name, in string }{
		{"wrong verifier", "18.972.631-8"},
		{"wrong verifier (K)", "23.478.522-K"},
		{"wrong verifier (0 instead of 5)", "12.345.678-0"},
		{"letters", "invalid"},
		{"letters only", "abcdefghi"},
		{"too few digits with dots", "1.1.1-1"},
		{"empty", ""},
		{"special characters", "12#34%56&789K"},
		{"comma grouping", "12,345,678-5"},
		{"non-canonical grouping", "12.345678-5"},
		{"non-canonical grouping", "12345.678-5"},
		{"dotted without hyphen", "12.345.6785"},
		{"dotted without hyphen, 7-digit body", "1.234.5674"},
		{"every digit dotted", "1.2.3.4.5.6.7.8-5"},
		{"internal whitespace", "12 345 678 5"},
		{"leading K", "K2345678-5"},
		{"K in the middle", "1234K678-5"},
		{"two K", "1234567K-K"},
		{"two leading K", "KK345678-5"},
		{"too short", "123"},
		{"7 characters", "1234567"},
		{"too long", "12345678901"},
		{"one above the ceiling", "100000000-7"},
		{"leading zeros, dotted", "009.068.826-K"},
		{"leading zeros, compact", "0009068826K"},
		{"leading zero, dotted 8-digit body", "018.972.631-7"},
		{"leading zeros, compact 7-digit body", "00012345674"},
		{"single leading zero on a 7-digit body", "01234567-4"},
		{"single leading zero, compact", "0123456785"},
		{"leading zero on a valid RUT", "010.000.002-4"},
		{"leading garbage, compact", "xx123456785"},
		{"trailing garbage, compact", "123456785xx"},
		{"leading garbage, hyphen", "xx12345678-5"},
		{"trailing garbage, hyphen", "12345678-5xx"},
		{"leading garbage, dotted", "xx12.345.678-5"},
		{"trailing garbage, dotted", "12.345.678-5xx"},
		{"65 units", strings.Repeat(" ", 56) + "123456785"},
		{"far over the cap", strings.Repeat("0", 128) + "x"},
		{"far over the cap", strings.Repeat("1", 20000)},
		{"adversarial 100k input", strings.Repeat("0", 100000) + "x"},
		{"whitespace only", "   "},
		{"tab and newline only", "\t\n"},
		{"letters with whitespace", "  invalid  "},
		{"wrong verifier with whitespace", "  12.345.678-0  "},
		{"Devanagari digits", "\u0967\u0968\u0969\u096a\u096b\u096c\u096d\u096e-\u096b"},
		{"full-width digits", "１２３４５６７８-５"},
		{"right-to-left override", "\u202e12345678-5"},
		{"trailing zero-width space", "12345678-5\u200b"},
		{"Cyrillic K", "14.625.621-\u041a"},
		{"NEL U+0085 is not trimmed", "12.345.678-5\u0085"},
	}
	for _, c := range cases {
		if Validate(c.in) {
			t.Errorf("%s: Validate(%q) = true, want false", c.name, truncate(c.in))
		}
	}
}

func TestValidateRejectsBodiesShorterThan7DigitsEvenWithTheRightVerifier(t *testing.T) {
	// Every verifier here is correct for its body: only the body-length rule
	// can reject these.
	for _, in := range []string{"999999-K", "999.999-K", "999999K", "100000-4", "1000-6", "1-9"} {
		if Validate(in) || Validate(in, WithStrict(true)) {
			t.Errorf("Validate(%q) = true, want false", in)
		}
	}
}

func TestValidateStrictRejectsRepeatedDigitPlaceholders(t *testing.T) {
	for _, in := range []string{"11.111.111-1", "22.222.222-2", "33.333.333-3", "8.888.888-K", "8888888K"} {
		if Validate(in, WithStrict(true)) {
			t.Errorf("Validate(%q, WithStrict(true)) = true, want false", in)
		}
		if !Validate(in) || !Validate(in, WithStrict(false)) {
			t.Errorf("Validate(%q) without strict = false, want true", in)
		}
	}
	for _, in := range []string{"12.345.678-5", "18.972.631-7"} {
		if !Validate(in, WithStrict(true)) {
			t.Errorf("Validate(%q, WithStrict(true)) = false, want true", in)
		}
	}
}

func TestValidateStrictAcceptsBodiesOneDigitFromRepeated(t *testing.T) {
	// Strict mode rejects only bodies whose digits are all the same: changing
	// any one digit of such a body, the first and the last included, makes it
	// pass.
	for _, n := range []int{7, 8} {
		for d := byte('1'); d <= '9'; d++ {
			for i := 0; i < n; i++ {
				body := []byte(strings.Repeat(string(d), n))
				body[i] = '1' + (d-'0')%9 // another nonzero digit
				in := string(body) + calculateVerifierForBody(string(body))
				if !Validate(in, WithStrict(true)) {
					t.Errorf("Validate(%q, WithStrict(true)) = false, want true", in)
				}
			}
		}
	}
	for _, in := range []string{"1.011.111-0", "11.111.112-K"} {
		if !Validate(in, WithStrict(true)) {
			t.Errorf("Validate(%q, WithStrict(true)) = false, want true", in)
		}
	}
}

func TestValidateLastOptionWins(t *testing.T) {
	if Validate("11.111.111-1", WithStrict(false), WithStrict(true)) {
		t.Error("the later WithStrict(true) should win")
	}
	if !Validate("11.111.111-1", WithStrict(true), WithStrict(false)) {
		t.Error("the later WithStrict(false) should win")
	}
}

func truncate(s string) string {
	if len(s) > 40 {
		return s[:40] + "…"
	}
	return s
}

func TestValidateIgnoresNilOptions(t *testing.T) {
	if !Validate("12.345.678-5", nil) {
		t.Error("Validate with a nil option = false, want true")
	}
}

func TestCleanThenValidateAcceptsMessyInput(t *testing.T) {
	// The recipe documented on Clean: normalize first, then prove Modulo 11.
	for _, in := range []string{"009.068.826-K", "RUT: 12.345.678-5", "12 345 678 5"} {
		c, err := Clean(in)
		if err != nil || !Validate(c) {
			t.Errorf("Validate(Clean(%q)) = false (%q, %v), want true", in, c, err)
		}
	}
}
