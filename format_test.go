package rut

import (
	"strings"
	"testing"
)

func TestFormatRendersTheCanonicalShape(t *testing.T) {
	cases := []struct {
		in   string
		opts []FormatOption
		want string
	}{
		{"189726317", []FormatOption{WithDots(true)}, "18.972.631-7"},
		{"189726317", []FormatOption{WithDots(false)}, "18972631-7"},
		{"123456785", []FormatOption{WithDots(true)}, "12.345.678-5"},
		{"123456785", []FormatOption{WithDots(false)}, "12345678-5"},
		{"14625621k", nil, "14.625.621-K"},
		{"09068826K", nil, "9.068.826-K"},
		{"0012345674", nil, "1.234.567-4"},
		{"009068826K", nil, "9.068.826-K"},
		{"14.625.621-k", nil, "14.625.621-K"},
		{"12#345#678#5", nil, "12.345.678-5"},
		{" 123 456 785 ", nil, "12.345.678-5"},
		{"12-345-678-5", nil, "12.345.678-5"},
		{"18.972.631-7", nil, "18.972.631-7"},
		{"9.068.826-K", nil, "9.068.826-K"},
		{"10000130", nil, "1.000.013-0"},
		{"\u00a012.345.678-5\ufeff", nil, "12.345.678-5"},
		{"1234567-4", []FormatOption{WithDots(false)}, "1234567-4"},
	}
	for _, c := range cases {
		got, err := Format(c.in, c.opts...)
		if err != nil || got != c.want {
			t.Errorf("Format(%q) = %q, %v; want %q, nil", c.in, got, err, c.want)
		}
	}
}

func TestFormatIsIdempotent(t *testing.T) {
	for _, in := range []string{"18.972.631-7", "14.625.621-K", "9.068.826-K"} {
		once, err := Format(in)
		if err != nil {
			t.Fatalf("Format(%q): %v", in, err)
		}
		if twice, err := Format(once); err != nil || twice != once {
			t.Errorf("Format(Format(%q)) = %q, %v; want %q", in, twice, err, once)
		}
	}
}

func TestFormatReturnsEmptyForEmptyInput(t *testing.T) {
	// As in rut.ts, the empty string is not an error.
	if got, err := Format(""); got != "" || err != nil {
		t.Errorf("Format(\"\") = %q, %v; want \"\", nil", got, err)
	}
}

func TestFormatRejectsWrongVerifiers(t *testing.T) {
	// rut.ts 3.x "repaired" a wrong verifier; since 4.0 it is rejected.
	for _, in := range []string{"123456789", "1234567K", "12345678K"} {
		got, err := Format(in)
		if got != "" {
			t.Errorf("Format(%q) = %q, want \"\"", in, got)
		}
		wantInvalid(t, "Format("+in+")", err)
	}
}

func TestFormatRejectsInvalidInput(t *testing.T) {
	for _, in := range []string{"123", "1234567", "12345678901", "K234567-8", "1234K678-9", "   ", "not-a-rut"} {
		got, err := Format(in, WithDots(false))
		if got != "" {
			t.Errorf("Format(%q) = %q, want \"\"", in, got)
		}
		wantInvalid(t, "Format("+in+")", err)
	}
}

func TestFormatCapsInputAt64UTF16Units(t *testing.T) {
	atCap := strings.Repeat("0", 55) + "123456785"
	if got, err := Format(atCap); err != nil || got != "12.345.678-5" {
		t.Errorf("64 units: Format = %q, %v; want \"12.345.678-5\", nil", got, err)
	}
	_, err := Format("0" + atCap)
	wantInvalid(t, "Format(65 units)", err)
}

func TestFormatLastOptionWinsAndNilIsIgnored(t *testing.T) {
	if got, _ := Format("123456785", WithDots(false), WithDots(true)); got != "12.345.678-5" {
		t.Errorf("Format with WithDots(false), WithDots(true) = %q, want dotted", got)
	}
	if got, _ := Format("123456785", nil); got != "12.345.678-5" {
		t.Errorf("Format with a nil option = %q, want dotted", got)
	}
}
