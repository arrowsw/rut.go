package rut

import (
	"bytes"
	"encoding/binary"
	"regexp"
	"testing"
)

var (
	dottedShape = regexp.MustCompile(`^[0-9]{1,2}\.[0-9]{3}\.[0-9]{3}-[0-9K]$`)
	compact8    = regexp.MustCompile(`^[0-9]{8}[0-9K]$`)
	hyphen8     = regexp.MustCompile(`^[0-9]{8}-[0-9K]$`)
	compact7    = regexp.MustCompile(`^[0-9]{7}[0-9K]$`)
	hyphen7     = regexp.MustCompile(`^[0-9]{7}-[0-9K]$`)
)

func TestGenerateEmitsADottedRutWithAn8DigitBody(t *testing.T) {
	for i := 0; i < 200; i++ {
		rut := Generate()
		if !dottedShape.MatchString(rut) {
			t.Fatalf("Generate() = %q, want the dotted shape", rut)
		}
		d, err := Decompose(rut)
		if err != nil || len(d.Body) != 8 || d.Body[0] == '0' {
			t.Fatalf("Generate() = %q: body %q, %v; want 8 digits", rut, d.Body, err)
		}
	}
}

func TestGenerateEmitsValidNonSuspiciousRuts(t *testing.T) {
	for i := 0; i < 1000; i++ {
		rut := Generate()
		if !Validate(rut, WithStrict(true)) {
			t.Fatalf("Generate() = %q, which fails Validate(WithStrict(true))", rut)
		}
		if f, err := Format(rut); err != nil || f != rut {
			t.Fatalf("Format(%q) = %q, %v; want it unchanged", rut, f, err)
		}
	}
}

func TestGenerateIsRandom(t *testing.T) {
	a, b, c := Generate(), Generate(), Generate()
	if a == b || b == c {
		t.Errorf("consecutive RUTs repeat: %q %q %q", a, b, c)
	}
	verifiers := map[string]bool{}
	for _, rut := range GenerateN(300) {
		verifiers[rut[len(rut)-1:]] = true
	}
	// P(no K in 300 draws) = (10/11)^300, about 4e-13.
	if len(verifiers) < 2 || !verifiers["K"] {
		t.Errorf("verifiers seen in 300 RUTs: %v; want several, K included", verifiers)
	}
}

func TestGenerateOptions(t *testing.T) {
	cases := []struct {
		name  string
		opts  []GenerateOption
		shape *regexp.Regexp
	}{
		{"default", nil, dottedShape},
		{"8-digit body", []GenerateOption{WithBodyLength(8)}, dottedShape},
		{"dotted", []GenerateOption{WithFormat(Dotted)}, dottedShape},
		{"compact", []GenerateOption{WithFormat(Compact)}, compact8},
		{"hyphen", []GenerateOption{WithFormat(Hyphen)}, hyphen8},
		{"7-digit compact", []GenerateOption{WithBodyLength(7), WithFormat(Compact)}, compact7},
		{"7-digit hyphen", []GenerateOption{WithBodyLength(7), WithFormat(Hyphen)}, hyphen7},
		// As in rut.ts, a body length other than 7 means 8 and an unknown
		// format means dotted.
		{"other body length", []GenerateOption{WithBodyLength(5), WithFormat(Hyphen)}, hyphen8},
		{"unknown format", []GenerateOption{WithFormat("spaced")}, dottedShape},
		{"nil option", []GenerateOption{nil}, dottedShape},
	}
	for _, c := range cases {
		for i := 0; i < 20; i++ {
			rut := Generate(c.opts...)
			if !c.shape.MatchString(rut) || !Validate(rut, WithStrict(true)) {
				t.Fatalf("%s: Generate() = %q, want shape %s and a valid RUT", c.name, rut, c.shape)
			}
		}
	}
	for i := 0; i < 50; i++ {
		if d, _ := Decompose(Generate(WithBodyLength(7))); len(d.Body) != 7 {
			t.Fatalf("WithBodyLength(7): body %q, want 7 digits", d.Body)
		}
	}
}

func TestGenerateN(t *testing.T) {
	ruts := GenerateN(25, WithBodyLength(7), WithFormat(Hyphen))
	if len(ruts) != 25 {
		t.Fatalf("GenerateN(25) returned %d RUTs", len(ruts))
	}
	for _, rut := range ruts {
		if !hyphen7.MatchString(rut) || !Validate(rut, WithStrict(true)) {
			t.Errorf("GenerateN item %q, want a valid 7-digit hyphen RUT", rut)
		}
	}
	for _, n := range []int{0, -1} {
		if got := GenerateN(n); got == nil || len(got) != 0 {
			t.Errorf("GenerateN(%d) = %#v, want an empty, non-nil slice", n, got)
		}
	}
}

// draws returns a random source that yields the given uint32 values, in the
// little-endian encoding randomIntInclusive reads.
func draws(values ...uint32) *bytes.Buffer {
	var buf bytes.Buffer
	for _, v := range values {
		binary.Write(&buf, binary.LittleEndian, v)
	}
	return &buf
}

func TestGenerateMapsDrawsToBodies(t *testing.T) {
	// An 8-digit body is 10,000,000 plus a draw modulo 90,000,000, and draws
	// from 4,230,000,000 up are discarded as biased. For 7 digits the numbers
	// are 1,000,000, 9,000,000 and 4,293,000,000. A repeated-digit body, the
	// all-nines top of the range included, is drawn again.
	seven := []GenerateOption{WithBodyLength(7)}
	cases := []struct {
		name  string
		opts  []GenerateOption
		draws []uint32
		want  string
	}{
		{"lowest draw gives the lowest body", nil, []uint32{0}, "10.000.000-8"},
		{"highest draw that is kept", nil, []uint32{89999998}, "99.999.998-0"},
		{"draws wrap around the range", nil, []uint32{90000000}, "10.000.000-8"},
		{"last unbiased draw", nil, []uint32{4229999998}, "99.999.998-0"},
		{"first biased draw is discarded", nil, []uint32{4230000000, 2345678}, "12.345.678-5"},
		{"repeated-digit body is redrawn", nil, []uint32{1111111, 2345678}, "12.345.678-5"},
		{"all-nines body is redrawn", nil, []uint32{89999999, 2345678}, "12.345.678-5"},
		{"7 digits: lowest body", seven, []uint32{0}, "1.000.000-9"},
		{"7 digits: highest draw that is kept", seven, []uint32{8999998}, "9.999.998-5"},
		{"7 digits: draws wrap around the range", seven, []uint32{9000000}, "1.000.000-9"},
		{"7 digits: last unbiased draw", seven, []uint32{4292999998}, "9.999.998-5"},
		{"7 digits: first biased draw is discarded", seven, []uint32{4293000000, 234567}, "1.234.567-4"},
		{"7 digits: repeated-digit body is redrawn", seven, []uint32{7888888, 234567}, "1.234.567-4"},
		{"7 digits: all-nines body is redrawn", seven, []uint32{8999999, 234567}, "1.234.567-4"},
		{"body length 6 means 8", []GenerateOption{WithBodyLength(6)}, []uint32{0}, "10.000.000-8"},
		{"body length 9 means 8", []GenerateOption{WithBodyLength(9)}, []uint32{0}, "10.000.000-8"},
		{"compact", []GenerateOption{WithFormat(Compact)}, []uint32{2345678}, "123456785"},
		{"hyphen", []GenerateOption{WithFormat(Hyphen)}, []uint32{2345678}, "12345678-5"},
		{"unknown format means dotted", []GenerateOption{WithFormat("spaced")}, []uint32{2345678}, "12.345.678-5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := draws(c.draws...)
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("needed more than the %d draws given: %v", len(c.draws), r)
				}
			}()
			if got := generateOne(src, collectGenerateOptions(c.opts)); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
			if src.Len() != 0 {
				t.Errorf("%d of %d draws left unread", src.Len()/4, len(c.draws))
			}
		})
	}
}

func TestGenerateNDrawsInOrder(t *testing.T) {
	src := draws(0, 1111111, 2345678)
	got := generateN(src, 2, collectGenerateOptions(nil))
	if len(got) != 2 || got[0] != "10.000.000-8" || got[1] != "12.345.678-5" || src.Len() != 0 {
		t.Errorf("generateN = %q with %d bytes unread, want [10.000.000-8 12.345.678-5] and none", got, src.Len())
	}
	if got := generateN(draws(), -1, collectGenerateOptions(nil)); got == nil || len(got) != 0 {
		t.Errorf("generateN(-1) = %#v, want an empty, non-nil slice", got)
	}
}

func TestRandomIntInclusiveRejectsBiasedDraws(t *testing.T) {
	// The first draw, 0xFFFFFFFF, is in the biased top of the uint32 range and
	// must be discarded; the second, 123456789, is accepted.
	buf := draws(0xFFFFFFFF, 123456789)
	got := randomIntInclusive(buf, 10000000, 99999999)
	if want := 10000000 + 123456789%90000000; got != want {
		t.Errorf("randomIntInclusive = %d, want %d", got, want)
	}
	if buf.Len() != 0 {
		t.Errorf("%d bytes left unread, want both draws consumed", buf.Len())
	}
}

func TestRandomIntInclusivePanicsWhenTheSourceFails(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("randomIntInclusive did not panic on an empty source")
		}
	}()
	randomIntInclusive(&bytes.Buffer{}, 1, 10)
}
