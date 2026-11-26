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

func TestRandomIntInclusiveRejectsBiasedDraws(t *testing.T) {
	// The first draw, 0xFFFFFFFF, is in the biased top of the uint32 range and
	// must be discarded; the second, 123456789, is accepted.
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint32(0xFFFFFFFF))
	binary.Write(&buf, binary.LittleEndian, uint32(123456789))
	got := randomIntInclusive(&buf, 10000000, 99999999)
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
