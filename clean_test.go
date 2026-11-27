package rut

import (
	"strings"
	"testing"
)

// wantInvalid fails the test unless err is exactly ErrInvalidRut: not wrapped,
// so no input can have been added to its message.
func wantInvalid(t *testing.T, call string, err error) {
	t.Helper()
	if err != ErrInvalidRut {
		t.Errorf("%s: err = %v, want ErrInvalidRut", call, err)
	}
}

func TestCleanNormalizesPermissively(t *testing.T) {
	// Clean strips everything outside [0-9kK], leading zeros and case. It
	// does not check the verifier digit.
	cases := []struct{ name, in, want string }{
		{"canonical dotted, lowercase k", "12.345.678-k", "12345678K"},
		{"canonical dotted, uppercase K", "12.345.678-K", "12345678K"},
		{"many leading zeros", "0000012.345.678-K", "12345678K"},
		{"7-digit body and leading zeros", "00009.068.826-K", "9068826K"},
		{"compact and many leading zeros", "00000012345678K", "12345678K"},
		{"compact, already clean", "123456789", "123456789"},
		{"compact with K, already clean", "12345678K", "12345678K"},
		{"scattered whitespace and hyphens", " 12 345 6 78 - 9 ", "123456789"},
		{"surrounding whitespace", "   12345678K   ", "12345678K"},
		{"extra hyphens", "12-34-56-789", "123456789"},
		{"extra hyphens with K", "12-345-678-K", "12345678K"},
		{"special characters interleaved", "12#34%56&@78K", "12345678K"},
		{"garbage before the verifier", "12.345.678@#$-K", "12345678K"},
		{"trailing garbage", "12345678#", "12345678"},
		{"trailing garbage after K", "12345678K###", "12345678K"},
		{"parentheses", "(12.345.678-K)", "12345678K"},
		{"dot before the verifier", "12.345.678.K", "12345678K"},
		{"lowercase k uppercased", "9068826k", "9068826K"},
		{"shortest length", "10000002", "10000002"},
		{"longest length", "999999996", "999999996"},
		{"embedded text", "RUT: 12.345.678-5", "123456785"},
		{"non-ASCII is stripped", "12\u00a0345\u2013678\u20145", "123456785"},
		{"full-width digits are stripped", "１12345678５5", "123456785"},
		{"invalid UTF-8 is stripped", "12\xff345\xfe6785", "123456785"},
		{"wrong verifier is kept", "12.345.678-9", "123456789"},
	}
	for _, c := range cases {
		got, err := Clean(c.in)
		if err != nil || got != c.want {
			t.Errorf("%s: Clean(%q) = %q, %v; want %q, nil", c.name, c.in, got, err, c.want)
		}
	}
}

func TestCleanRejectsWhatCannotBeARut(t *testing.T) {
	cases := []struct{ name, in string }{
		{"empty", ""},
		{"whitespace only", "   "},
		{"too short after normalization", "1.2.34-K"},
		{"too short", "123"},
		{"7 characters", "1234567"},
		{"too long after normalization", "12.345.6789012-3"},
		{"11 characters", "12345678901"},
		{"leading K", "K1234567"},
		{"K in the middle", "1234K567"},
		{"two K", "KK345678"},
		{"trailing KK", "1234567KK"},
		{"only invalid characters", "####"},
		{"only letters", "abcdefgh"},
		{"only separators", "--------"},
		{"only zeros", "000000000"},
	}
	for _, c := range cases {
		got, err := Clean(c.in)
		if got != "" {
			t.Errorf("%s: Clean(%q) = %q, want \"\"", c.name, c.in, got)
		}
		wantInvalid(t, "Clean("+c.in+")", err)
	}
}

func TestCleanCapsInputAt64UTF16Units(t *testing.T) {
	for _, pad := range []string{"0", " ", "\u00a0", "\u3000"} {
		atCap := strings.Repeat(pad, 55) + "123456785"
		if got, err := Clean(atCap); err != nil || got != "123456785" {
			t.Errorf("64 units padded with %q: got %q, %v; want \"123456785\", nil", pad, got, err)
		}
		_, err := Clean(pad + atCap)
		wantInvalid(t, "65 units padded with "+pad, err)
	}

	// A rune outside the Basic Multilingual Plane is two UTF-16 units.
	emoji := strings.Repeat("\U0001F600", 27) + "123456785" // 54 + 9 = 63 units
	if got, err := Clean(emoji); err != nil || got != "123456785" {
		t.Errorf("63 units with emoji: got %q, %v; want \"123456785\", nil", got, err)
	}
	_, err := Clean("\U0001F600" + emoji) // 65 units
	wantInvalid(t, "65 units with emoji", err)
}

func TestCleanCountsUnitsAtTheCapEdges(t *testing.T) {
	// Each pad is repeated, and topped up with spaces, to exactly 64 units;
	// one more space must take the input over the cap.
	cases := []struct {
		name  string
		pad   string
		units int
	}{
		{"last BMP character", "\uffff", 1},
		{"replacement character", "\ufffd", 1},
		{"first character outside the BMP", "\U00010000", 2},
		{"last code point", "\U0010ffff", 2},
		// Go strings can hold bytes that are not UTF-8, which JavaScript strings
		// cannot. Each such byte counts as one unit.
		{"invalid byte", "\xff", 1},
		{"truncated 3-byte sequence", "\xe2\x82", 2},
		{"UTF-8-encoded surrogate", "\xed\xa0\x80", 3},
	}
	const want = "123456785"
	for _, c := range cases {
		room := maxInputLength - len(want)
		n := room / c.units
		atCap := strings.Repeat(c.pad, n) + strings.Repeat(" ", room-n*c.units) + want
		if got, err := Clean(atCap); err != nil || got != want {
			t.Errorf("%s, 64 units: got %q, %v; want %q, nil", c.name, got, err, want)
		}
		_, err := Clean(" " + atCap)
		wantInvalid(t, c.name+", 65 units", err)
	}
}

func TestCleanIsIdempotent(t *testing.T) {
	for _, in := range []string{"12.345.678-5", "14.625.621-k", "009.068.826-K", " 12 345 6 78 - 9 ", "(18.972.631-7)"} {
		once, err := Clean(in)
		if err != nil {
			t.Fatalf("Clean(%q): %v", in, err)
		}
		if twice, err := Clean(once); err != nil || twice != once {
			t.Errorf("Clean(Clean(%q)) = %q, %v; want %q", in, twice, err, once)
		}
	}
}
