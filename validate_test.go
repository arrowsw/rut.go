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
