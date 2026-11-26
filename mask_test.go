package rut

import (
	"strings"
	"testing"
)

func TestMaskKeepsTheHeadAndTheVerifier(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"8-digit body", "12.345.678-5", "12.***.***-5"},
		{"7-digit body", "1.234.567-4", "1.***.***-4"},
		{"compact input", "123456785", "12.***.***-5"},
		{"compact with hyphen", "12345678-5", "12.***.***-5"},
		{"lowercase k", "14.625.621-k", "14.***.***-K"},
		{"leading zeros", "0009068826K", "9.***.***-K"},
		{"wrong verifier is still masked", "12.345.678-9", "12.***.***-9"},
	}
	for _, c := range cases {
		got, err := Mask(c.in)
		if err != nil || got != c.want {
			t.Errorf("%s: Mask(%q) = %q, %v; want %q, nil", c.name, c.in, got, err, c.want)
		}
	}
}

func TestMaskNeverRevealsTheMaskedDigits(t *testing.T) {
	got, _ := Mask("12.345.678-5")
	if strings.Contains(got, "345") || strings.Contains(got, "678") {
		t.Errorf("Mask(\"12.345.678-5\") = %q reveals masked digits", got)
	}
}

func TestMaskRejectsInvalidInput(t *testing.T) {
	for _, in := range []string{"", "123", "invalid", "not-a-rut", strings.Repeat("1", 128)} {
		got, err := Mask(in)
		if got != "" {
			t.Errorf("Mask(%q) = %q, want \"\"", truncate(in), got)
		}
		wantInvalid(t, "Mask("+truncate(in)+")", err)
	}
}
