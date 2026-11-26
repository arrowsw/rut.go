// Package rut handles Chilean RUT values: validation, cleaning, formatting,
// masking and generation.
//
// It is a port of the rut.ts library (https://github.com/arrowsw/rut.ts) and
// mirrors its behavior exactly, character for character, except for the
// format-as-you-type mode of rut.ts, which this package does not include.
//
// Three details of rut.ts come from JavaScript and are reproduced on purpose:
//
//   - "Whitespace" is the set of 25 characters that JavaScript's trim() and
//     \s accept (ASCII spaces and tabs, U+00A0, U+FEFF, U+3000 and the other
//     Unicode spaces). It is not the set used by [strings.TrimSpace].
//   - The 64-character input cap counts UTF-16 code units, as JavaScript's
//     length does, not bytes. Each byte that is not valid UTF-8 counts as one.
//   - Error values never contain the input, so RUTs do not leak into logs.
package rut
