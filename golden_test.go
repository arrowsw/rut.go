package rut

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// golden is testdata/golden.json.gz: what the pinned rut.ts release answers
// for 44,000 inputs and 8,000 pairs. scripts/golden/generate.mjs writes it;
// CI checks that it is up to date.
type golden struct {
	RutTs string  `json:"rutTs"`
	Cases [][]any `json:"cases"` // input, clean, calculateVerifier, format, formatNoDots, mask, flags
	Pairs [][]any `json:"pairs"` // a, b, flags
}

func loadGolden(t *testing.T) golden {
	t.Helper()
	f, err := os.Open("testdata/golden.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var g golden
	if err := json.NewDecoder(zr).Decode(&g); err != nil {
		t.Fatal(err)
	}
	if len(g.Cases) == 0 || len(g.Pairs) == 0 {
		t.Fatal("golden fixture is empty")
	}
	return g
}

// result renders a Go result the way the fixture stores a rut.ts one: the
// string, or null when rut.ts returned null (an error in Go).
func result(s string, err error) any {
	if err == ErrInvalidRut {
		return nil
	}
	if err != nil {
		return fmt.Sprintf("unexpected error %v", err)
	}
	return s
}

func flag(flags any, bit int) bool {
	return int(flags.(float64))&(1<<bit) != 0
}

func TestGoldenMatchesRutTs(t *testing.T) {
	g := loadGolden(t)
	t.Logf("replaying %d inputs and %d pairs from rut.ts %s", len(g.Cases), len(g.Pairs), g.RutTs)

	failures := map[string]int{}
	check := func(fn, in string, got, want any) {
		if got == want {
			return
		}
		if failures[fn] < 5 {
			t.Errorf("%s(%+q) = %#v, rut.ts gives %#v", fn, in, got, want)
		}
		failures[fn]++
	}

	for _, c := range g.Cases {
		in := c[0].(string)
		clean := c[1]

		check("Clean", in, result(Clean(in)), clean)

		var wantBody, wantVerifier any
		wantDecomposed := any(nil)
		if s, ok := clean.(string); ok {
			wantBody, wantVerifier = s[:len(s)-1], s[len(s)-1:]
			wantDecomposed = DecomposedRut{Body: s[:len(s)-1], Verifier: s[len(s)-1:]}
		}
		check("GetBody", in, result(GetBody(in)), wantBody)
		check("GetVerifier", in, result(GetVerifier(in)), wantVerifier)
		if d, err := Decompose(in); err == nil {
			check("Decompose", in, d, wantDecomposed)
		} else {
			check("Decompose", in, result("", err), wantDecomposed)
		}

		check("CalculateVerifier", in, result(CalculateVerifier(in)), c[2])
		check("Format", in, result(Format(in)), c[3])
		check("Format(WithDots(false))", in, result(Format(in, WithDots(false))), c[4])
		check("Mask", in, result(Mask(in)), c[5])
		check("IsRutLike", in, IsRutLike(in), flag(c[6], 0))
		check("Validate", in, Validate(in), flag(c[6], 1))
		check("Validate(WithStrict(true))", in, Validate(in, WithStrict(true)), flag(c[6], 2))
	}

	for _, p := range g.Pairs {
		a, b := p[0].(string), p[1].(string)
		in := a + `", "` + b
		check("Equals", in, Equals(a, b), flag(p[2], 0))
		check("Equals(WithRequireValid(false))", in, Equals(a, b, WithRequireValid(false)), flag(p[2], 1))
	}

	for fn, n := range failures {
		t.Errorf("%s: %d mismatches with rut.ts", fn, n)
	}
}
