package rut

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
)

// golden is testdata/golden.json.gz: what the pinned rut.ts release answers
// for 48,537 inputs and 11,000 pairs, its RUT_PATTERN, and what its generate()
// returns in 3,000 calls fed a fixed stream of random draws.
// scripts/golden/generate.mjs writes it; CI checks that it is up to date.
type golden struct {
	RutTs    string            `json:"rutTs"`
	Cases    [][]any           `json:"cases"` // input, clean, calculateVerifier, format, formatNoDots, mask, flags
	Pairs    [][]any           `json:"pairs"` // a, b, flags
	Patterns map[string]string `json:"patterns"`
	Generate struct {
		Draws []uint32 `json:"draws"`
		Calls [][]any  `json:"calls"` // bodyLength, format, count, result
	} `json:"generate"`
}

var (
	goldenOnce sync.Once
	goldenData golden
	goldenErr  error
)

func loadGolden(t *testing.T) golden {
	t.Helper()
	goldenOnce.Do(func() { goldenData, goldenErr = readGolden("testdata/golden.json.gz") })
	if goldenErr != nil {
		t.Fatal(goldenErr)
	}
	return goldenData
}

func readGolden(path string) (golden, error) {
	var g golden
	f, err := os.Open(path)
	if err != nil {
		return g, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return g, err
	}
	if err := json.NewDecoder(zr).Decode(&g); err != nil {
		return g, err
	}
	if len(g.Cases) == 0 || len(g.Pairs) == 0 || len(g.Patterns) == 0 ||
		len(g.Generate.Draws) == 0 || len(g.Generate.Calls) == 0 {
		return g, fmt.Errorf("%s is missing a section", path)
	}
	return g, nil
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

func TestGoldenPatternsMatchRutTs(t *testing.T) {
	g := loadGolden(t)
	want := map[string]string{
		"any":     PatternAny,
		"compact": PatternCompact,
		"hyphen":  PatternHyphen,
		"dotted":  PatternDotted,
	}
	if !reflect.DeepEqual(g.Patterns, want) {
		t.Errorf("RUT_PATTERN in rut.ts %s is %q; the Pattern constants are %q", g.RutTs, g.Patterns, want)
	}
}

func TestGoldenGenerateMatchesRutTs(t *testing.T) {
	g := loadGolden(t)
	src := draws(g.Generate.Draws...)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Go needed more than the %d random draws rut.ts used: %v", len(g.Generate.Draws), r)
		}
	}()

	failures := 0
	for _, c := range g.Generate.Calls {
		var opts []GenerateOption
		if c[0] != nil {
			opts = append(opts, WithBodyLength(int(c[0].(float64))))
		}
		if c[1] != nil {
			opts = append(opts, WithFormat(GenerateFormat(c[1].(string))))
		}
		o := collectGenerateOptions(opts)

		var got any
		if c[2] == nil {
			got = generateOne(src, o)
		} else {
			list := []any{}
			for _, rut := range generateN(src, int(c[2].(float64)), o) {
				list = append(list, rut)
			}
			got = list
		}
		if !reflect.DeepEqual(got, c[3]) {
			if failures < 5 {
				t.Errorf("generate(bodyLength %v, format %v, count %v) = %v, rut.ts gives %v", c[0], c[1], c[2], got, c[3])
			}
			failures++
		}
	}
	if failures > 0 {
		t.Errorf("generate: %d of %d calls differ from rut.ts", failures, len(g.Generate.Calls))
	}
	if src.Len() != 0 {
		t.Errorf("Go used %d fewer random draws than rut.ts", src.Len()/4)
	}
}
