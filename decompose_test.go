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

func TestGetVerifierExtractsTheVerifier(t *testing.T) {
	cases := []struct{ in, want string }{
		{"23.579.222-2", "2"},
		{"18.972.631-7", "7"},
		{"9.068.826-K", "K"},
		{"18972631-7", "7"},
		{"189726317", "7"},
		{"009.068.826-K", "K"},
		{"  18.972.631-7  ", "7"},
		{"(18.972.631-7)", "7"},
		{"18#972#631-7", "7"},
		{"9.068.826-k", "K"},
		{"9068826k", "K"},
		{"12.345.678-9", "9"}, // a wrong verifier is still extracted
	}
	for _, c := range cases {
		got, err := GetVerifier(c.in)
		if err != nil || got != c.want {
			t.Errorf("GetVerifier(%q) = %q, %v; want %q, nil", c.in, got, err, c.want)
		}
	}
}

func TestGetVerifierReachesEveryPossibleOutput(t *testing.T) {
	cases := []struct{ body, dv string }{
		{"10000004", "0"}, {"10000009", "1"}, {"10000003", "2"}, {"10000008", "3"},
		{"10000002", "4"}, {"10000007", "5"}, {"10000001", "6"}, {"10000006", "7"},
		{"10000000", "8"}, {"10000005", "9"}, {"10000013", "K"},
	}
	for _, c := range cases {
		if got, err := GetVerifier(c.body + c.dv); err != nil || got != c.dv {
			t.Errorf("GetVerifier(%q) = %q, %v; want %q, nil", c.body+c.dv, got, err, c.dv)
		}
	}
}

func TestGetVerifierRejectsInvalidInput(t *testing.T) {
	for _, in := range []string{"", "123", "invalid", "1234567", "12345678901"} {
		got, err := GetVerifier(in)
		if got != "" {
			t.Errorf("GetVerifier(%q) = %q, want \"\"", in, got)
		}
		wantInvalid(t, "GetVerifier("+in+")", err)
	}
}

func TestGetBodyPlusGetVerifierIsClean(t *testing.T) {
	for _, in := range []string{"18.972.631-7", "9.068.826-K", "14.625.621-k", "009.068.826-K", " 18972631-7 "} {
		body, _ := GetBody(in)
		verifier, _ := GetVerifier(in)
		if cleaned, err := Clean(in); err != nil || body+verifier != cleaned {
			t.Errorf("GetBody(%q) + GetVerifier(%q) = %q, Clean = %q, %v", in, in, body+verifier, cleaned, err)
		}
	}
}
