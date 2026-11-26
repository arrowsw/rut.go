package rut

import (
	"strings"
	"testing"
)

func TestGetBodyExtractsTheBody(t *testing.T) {
	cases := []struct{ in, want string }{
		{"18.972.631-7", "18972631"},
		{"9.068.826-K", "9068826"},
		{"14.625.621-k", "14625621"},
		{"189726317", "18972631"},
		{"9068826K", "9068826"},
		{"009.068.826-K", "9068826"},
		{"  18.972.631-7  ", "18972631"},
		{"(18.972.631-7)", "18972631"},
		{"18#972#631-7", "18972631"},
		{"18.264.159-0", "18264159"},
		{"1.000.000-2", "1000000"},
	}
	for _, c := range cases {
		got, err := GetBody(c.in)
		if err != nil || got != c.want {
			t.Errorf("GetBody(%q) = %q, %v; want %q, nil", c.in, got, err, c.want)
		}
	}
}

func TestGetBodyNeverContainsK(t *testing.T) {
	for _, in := range []string{"9.068.826-K", "14.625.621-k", "9068826K", "14625621K"} {
		got, err := GetBody(in)
		if err != nil || strings.ContainsAny(got, "kK") {
			t.Errorf("GetBody(%q) = %q, %v; want a body without K", in, got, err)
		}
	}
}

func TestGetBodyRejectsInvalidInput(t *testing.T) {
	for _, in := range []string{"", "123", "invalid", "1234567", "12345678901"} {
		got, err := GetBody(in)
		if got != "" {
			t.Errorf("GetBody(%q) = %q, want \"\"", in, got)
		}
		wantInvalid(t, "GetBody("+in+")", err)
	}
}

func TestGetBodyAgreesWithCalculateVerifier(t *testing.T) {
	for _, in := range []string{"18.972.631-7", "12.345.678-5", "99.999.999-9"} {
		body, err := GetBody(in)
		if err != nil {
			t.Fatalf("GetBody(%q): %v", in, err)
		}
		if v, err := CalculateVerifier(body); err != nil || v != in[len(in)-1:] {
			t.Errorf("CalculateVerifier(GetBody(%q)) = %q, %v; want %q", in, v, err, in[len(in)-1:])
		}
	}
}
