package rut

import (
	"errors"
	"strings"
	"testing"
)

func TestCalculateVerifierCoversEveryPossibleOutput(t *testing.T) {
	// Smallest 8-digit body producing each of the 11 possible verifiers, pinned
	// from rut.ts so a refactor of the Modulo 11 loop is caught at once.
	cases := []struct{ body, want string }{
		{"10000004", "0"},
		{"10000009", "1"},
		{"10000003", "2"},
		{"10000008", "3"},
		{"10000002", "4"},
		{"10000007", "5"},
		{"10000001", "6"},
		{"10000006", "7"},
		{"10000000", "8"},
		{"10000005", "9"},
		{"10000013", "K"},
	}
	for _, c := range cases {
		got, err := CalculateVerifier(c.body)
		if err != nil || got != c.want {
			t.Errorf("CalculateVerifier(%q) = %q, %v; want %q, nil", c.body, got, err, c.want)
		}
	}
}

func TestCalculateVerifierNormalizesItsInput(t *testing.T) {
	cases := []struct{ name, body, want string }{
		{"8-digit body", "18264958", "9"},
		{"8-digit body", "12345678", "5"},
		{"8-digit body", "18972631", "7"},
		{"K output", "24657622", "K"},
		{"K output", "14625621", "K"},
		{"dots", "18.264.958", "9"},
		{"hyphens", "18-264-958", "9"},
		{"leading zero", "018972631", "7"},
		{"leading zeros", "0018264958", "9"},
		{"floor of the range", "1000000", "9"},
		{"minimum 8-digit body", "10000000", "8"},
		{"ceiling of the range", "99999999", "9"},
		{"ASCII spaces", " 12 345 678 ", "5"},
		{"tab and newline", "12\t345\n678", "5"},
		{"vertical tab (in JS \\s, not in Go's \\s)", "12\v345678", "5"},
		{"no-break space U+00A0", "12\u00a0345\u00a0678", "5"},
		{"BOM U+FEFF (in JS \\s, not in unicode.IsSpace)", "\ufeff12345678", "5"},
		{"ideographic space U+3000", "12\u3000345\u3000678", "5"},
		{"line separator U+2028", "12\u2028345678", "5"},
		{"zero is stripped after separators", " 0 12345678", "5"},
	}
	for _, c := range cases {
		got, err := CalculateVerifier(c.body)
		if err != nil || got != c.want {
			t.Errorf("%s: CalculateVerifier(%q) = %q, %v; want %q, nil", c.name, c.body, got, err, c.want)
		}
	}
}

func TestCalculateVerifierRejectsInvalidInput(t *testing.T) {
	cases := []struct{ name, body string }{
		{"empty", ""},
		{"too short", "123"},
		{"too short", "12345"},
		{"too long", "12345678901"},
		{"one above the ceiling", "100000000"},
		{"letters", "abcdefgh"},
		{"embedded letters", "12abc345678"},
		{"other separators", "12#345#678"},
		{"leading-zero strip leaves too few digits", "0000000"},
		{"leading-zero strip leaves too few digits", "000000"},
		{"body below 7 digits", "999999"},
		{"body below 7 digits", "100000"},
		{"body below 7 digits", "1000"},
		{"body below 7 digits", "1"},
		{"K is not a body digit", "1234567K"},
		{"NEL U+0085 is not JS whitespace", "12\u0085345678"},
		{"zero-width space is not whitespace", "12\u200b345678"},
		{"Mongolian vowel separator is not whitespace", "12\u180e345678"},
		{"full-width digits are not digits", "１２３４５６７８"},
		{"invalid UTF-8", "1234\xff5678"},
	}
	for _, c := range cases {
		got, err := CalculateVerifier(c.body)
		if !errors.Is(err, ErrInvalidRut) || got != "" {
			t.Errorf("%s: CalculateVerifier(%q) = %q, %v; want \"\", ErrInvalidRut", c.name, c.body, got, err)
		}
	}
}

func TestCalculateVerifierCapsInputAt64UTF16Units(t *testing.T) {
	// The cap counts UTF-16 code units, as JavaScript's length does: U+00A0
	// is two bytes in Go but one unit in JavaScript.
	for _, pad := range []string{" ", "\u00a0", "\u3000"} {
		atCap := "12345678" + strings.Repeat(pad, 56)
		if got, err := CalculateVerifier(atCap); err != nil || got != "5" {
			t.Errorf("64 units padded with %q: got %q, %v; want \"5\", nil", pad, got, err)
		}
		overCap := atCap + pad
		if _, err := CalculateVerifier(overCap); !errors.Is(err, ErrInvalidRut) {
			t.Errorf("65 units padded with %q: got err %v; want ErrInvalidRut", pad, err)
		}
	}
}
