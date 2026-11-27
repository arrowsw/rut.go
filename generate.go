package rut

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

// GenerateFormat is the output shape of Generate.
type GenerateFormat string

// The shapes Generate can emit.
const (
	Dotted  GenerateFormat = "dotted"  // 12.345.678-5, the default
	Compact GenerateFormat = "compact" // 123456785
	Hyphen  GenerateFormat = "hyphen"  // 12345678-5
)

// GenerateOption configures Generate and GenerateN.
type GenerateOption func(*generateOptions)

type generateOptions struct {
	bodyLength int
	format     GenerateFormat
}

// WithBodyLength sets the number of body digits: 7, or 8 (the default). Any
// other value means 8, as in rut.ts.
func WithBodyLength(n int) GenerateOption {
	return func(o *generateOptions) { o.bodyLength = n }
}

// WithFormat sets the output shape: Dotted (the default), Compact or Hyphen.
// An unknown value means Dotted, as in rut.ts.
func WithFormat(f GenerateFormat) GenerateOption {
	return func(o *generateOptions) { o.format = f }
}

// Generate returns a random valid RUT, 12.345.678-5 style by default.
//
// Randomness comes from crypto/rand, with unbiased sampling, and
// repeated-digit bodies such as 11.111.111 are never produced, so every
// result also passes Validate with WithStrict(true).
//
// Generated RUTs cover the real RUT range, so any one of them may belong to a
// real person: use them as test data, not as published sample identities.
//
// If crypto/rand fails to return random bytes, Generate panics with an error
// rather than return a RUT that is not random. crypto/rand documents that its
// operating-system sources do not fail, except on legacy Linux systems.
func Generate(opts ...GenerateOption) string {
	return generateOne(rand.Reader, collectGenerateOptions(opts))
}

// GenerateN returns n random valid RUTs, configured like Generate. It returns
// an empty slice when n is zero or negative.
//
// There is no upper limit: if n comes from user input, cap it before calling.
// GenerateN allocates all n results at once, so a count too large to allocate
// panics, as make does. It also panics, like Generate, if crypto/rand fails.
func GenerateN(n int, opts ...GenerateOption) []string {
	return generateN(rand.Reader, n, collectGenerateOptions(opts))
}

func collectGenerateOptions(opts []GenerateOption) generateOptions {
	o := generateOptions{bodyLength: 8, format: Dotted}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}

// generateN and generateOne take the random source as a parameter so that
// tests can feed them a fixed stream of draws.
func generateN(r io.Reader, n int, o generateOptions) []string {
	if n < 0 {
		n = 0
	}
	ruts := make([]string, n)
	for i := range ruts {
		ruts[i] = generateOne(r, o)
	}
	return ruts
}

func generateOne(r io.Reader, o generateOptions) string {
	lo, hi := 10000000, 99999999
	if o.bodyLength == 7 {
		lo, hi = 1000000, 9999999
	}
	body := strconv.Itoa(randomIntInclusive(r, lo, hi))
	for isSuspicious(body) {
		body = strconv.Itoa(randomIntInclusive(r, lo, hi))
	}
	verifier := calculateVerifierForBody(body)
	switch o.format {
	case Compact:
		return body + verifier
	case Hyphen:
		return body + "-" + verifier
	default:
		return formatBody(body, true) + "-" + verifier
	}
}

// randomIntInclusive draws a uniform integer in [lo, hi] from r. Draws in the
// top of the uint32 range that would bias the result are discarded and
// redrawn, as in rut.ts. It panics with an error wrapping the read error if r
// fails, which crypto/rand.Reader does not do in practice.
func randomIntInclusive(r io.Reader, lo, hi int) int {
	span := uint64(hi - lo + 1)
	maxUnbiased := (1 << 32) / span * span
	var buf [4]byte
	for {
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			panic(fmt.Errorf("rut: reading random bytes: %w", err))
		}
		if v := uint64(binary.LittleEndian.Uint32(buf[:])); v < maxUnbiased {
			return lo + int(v%span)
		}
	}
}
