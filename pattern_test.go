package rut

import (
	"regexp"
	"strings"
	"testing"
)

var patterns = map[string]*regexp.Regexp{
	"any":     regexp.MustCompile(PatternAny),
	"compact": regexp.MustCompile(PatternCompact),
	"hyphen":  regexp.MustCompile(PatternHyphen),
	"dotted":  regexp.MustCompile(PatternDotted),
}

// shapeOf names the shape an accepted, already trimmed input is written in.
func shapeOf(s string) string {
	switch {
	case strings.Contains(s, "."):
		return "dotted"
	case strings.Contains(s, "-"):
		return "hyphen"
	default:
		return "compact"
	}
}

// patternMismatches lists where the patterns disagree with IsRutLike on raw:
// PatternAny must match exactly what IsRutLike accepts, and an accepted input
// must match exactly one of the three per-shape patterns.
func patternMismatches(raw string) []string {
	input := strings.TrimFunc(raw, isJSSpace)
	accepted := IsRutLike(raw)
	var bad []string
	if patterns["any"].MatchString(input) != accepted {
		bad = append(bad, "any")
	}
	for _, shape := range []string{"compact", "hyphen", "dotted"} {
		if patterns[shape].MatchString(input) != (accepted && shapeOf(input) == shape) {
			bad = append(bad, shape)
		}
	}
	return bad
}

func TestPatternsAreFrozenByteForByte(t *testing.T) {
	// The same strings as rut.ts's RUT_PATTERN: changing one character breaks
	// every schema and database constraint they were pasted into.
	want := map[string]string{
		PatternAny:     "^([1-9][0-9]{6,7}-?[0-9kK]|[1-9][0-9]?[.][0-9]{3}[.][0-9]{3}-[0-9kK])$",
		PatternCompact: "^[1-9][0-9]{6,7}[0-9kK]$",
		PatternHyphen:  "^[1-9][0-9]{6,7}-[0-9kK]$",
		PatternDotted:  "^[1-9][0-9]?[.][0-9]{3}[.][0-9]{3}-[0-9kK]$",
	}
	for got, w := range want {
		if got != w {
			t.Errorf("pattern = %q, want %q", got, w)
		}
	}
	for _, p := range []string{PatternAny, PatternCompact, PatternHyphen, PatternDotted} {
		if !strings.HasPrefix(p, "^") || !strings.HasSuffix(p, "$") || strings.Contains(p, `\`) {
			t.Errorf("%q must be anchored and backslash-free", p)
		}
	}
}

func TestPatternsAgreeWithIsRutLike(t *testing.T) {
	inputs := []string{
		"123456785", "12345678-5", "12.345.678-5", "1234567-4", "1.234.567-4",
		"14625621-k", "14.625.621-K", "  12.345.678-5  ", "12.345.678-5\n",
		"\ufeff12.345.678-5", "12.345.678-5\u3000",
		"01234567-4", "012.345.678-5", "0012345678", "123.456.789-5",
		"100000000-7", "999999-K", "12.345.6785", "12.345678-5", "12345.678-5",
		"12 345 678-5", "12345678--5", "K2345678-5",
		"１２３４５６７８-5", "\u0661\u0662\u0663\u0664\u0665\u0666\u0667\u0668-5",
		"", "   ",
	}
	for _, in := range inputs {
		if bad := patternMismatches(in); bad != nil {
			t.Errorf("%q: patterns %v disagree with IsRutLike", in, bad)
		}
	}
}

func TestPatternsAgreeWithIsRutLikeExhaustively(t *testing.T) {
	// Every string up to 9 characters over a reduced alphabet: 2.4 million
	// strings covering every compact and hyphen boundary.
	maxLen := 9
	if testing.Short() {
		maxLen = 7
	}
	alphabet := []byte{'0', '1', 'K', '-', '.'}
	buf := make([]byte, 0, maxLen)
	mismatches := 0
	var walk func()
	walk = func() {
		if bad := patternMismatches(string(buf)); bad != nil {
			if mismatches < 10 {
				t.Errorf("%q: patterns %v disagree with IsRutLike", buf, bad)
			}
			mismatches++
		}
		if len(buf) == maxLen {
			return
		}
		for _, c := range alphabet {
			buf = append(buf, c)
			walk()
			buf = buf[:len(buf)-1]
		}
	}
	walk()
	if mismatches > 0 {
		t.Errorf("%d mismatches in total", mismatches)
	}
}

func TestGeneratedRutsMatchTheirPattern(t *testing.T) {
	for format, shape := range map[GenerateFormat]string{Compact: "compact", Hyphen: "hyphen", Dotted: "dotted"} {
		for _, length := range []int{7, 8} {
			for _, rut := range GenerateN(200, WithFormat(format), WithBodyLength(length)) {
				if !patterns[shape].MatchString(rut) || !patterns["any"].MatchString(rut) {
					t.Fatalf("%q does not match the %s pattern", rut, shape)
				}
			}
		}
	}
}

func TestPatternsDocumentedLimits(t *testing.T) {
	// Shape only, never the verifier digit.
	if !patterns["hyphen"].MatchString("12345678-0") || Validate("12345678-0") {
		t.Error("a pattern checks shape only, never the verifier")
	}
	// Any and compact match any 9-digit number, such as a phone number.
	if !patterns["compact"].MatchString("912345678") || patterns["hyphen"].MatchString("912345678") {
		t.Error("compact must match a bare 9-digit number and hyphen must not")
	}
	// The 64-unit cap belongs to IsRutLike, not to the patterns.
	padded := strings.Repeat(" ", 60) + "12345678-5"
	if IsRutLike(padded) || !patterns["any"].MatchString(strings.TrimSpace(padded)) {
		t.Error("past the cap IsRutLike must refuse while the trimmed shape still matches")
	}
}
