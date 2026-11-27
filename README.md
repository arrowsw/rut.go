# rut.go

Handle Chilean RUT values in Go: validate, clean, format, mask, compare and generate them.

A faithful port of [rut.ts](https://github.com/arrowsw/rut.ts). Every function gives the same answer as rut.ts 5.1.2 for the same input, and the test suite proves it against more than 60,000 cases taken from rut.ts itself. No dependencies.

```bash
go get github.com/arrowsw/rut.go
```

Requires Go 1.22 or later. The import path ends in `.go`; the package is `rut`.

## Quick start

```go
import rut "github.com/arrowsw/rut.go"

rut.Validate("12.345.678-5")                      // true
rut.Validate("12.345.678-9")                      // false: wrong verifier
rut.Validate("11.111.111-1", rut.WithStrict(true)) // false: placeholder

c, err := rut.Clean("RUT: 12.345.678-k")          // "12345678K", nil
f, err := rut.Format("123456785")                 // "12.345.678-5", nil
f, err  = rut.Format("123456785", rut.WithDots(false)) // "12345678-5", nil
m, err := rut.Mask("12.345.678-5")                // "12.***.***-5", nil
v, err := rut.CalculateVerifier("12.345.678")     // "5", nil

rut.Equals("12.345.678-5", "123456785")           // true

if errors.Is(err, rut.ErrInvalidRut) { /* … */ }
```

To accept messy input (leading zeros, embedded text) as a real RUT, clean it first and then validate it:

```go
if c, err := rut.Clean(raw); err == nil && rut.Validate(c) {
	// c is a valid RUT in compact form, ready to store
}
```

## API

| rut.ts | rut.go | Notes |
|---|---|---|
| `validate(rut, { strict })` | `Validate(rut, WithStrict(b)) bool` | |
| `isValidRut(rut)` | `Validate` | Only a TypeScript type guard; same answer |
| `isRutLike(rut)` | `IsRutLike(rut) bool` | |
| `clean(rut)` | `Clean(rut) (string, error)` | |
| `getBody(rut)` | `GetBody(rut) (string, error)` | |
| `getVerifier(rut)` | `GetVerifier(rut) (string, error)` | `"0"`–`"9"` or `"K"` |
| `decompose(rut)` | `Decompose(rut) (DecomposedRut, error)` | `DecomposedRut{Body, Verifier}` |
| `calculateVerifier(body)` | `CalculateVerifier(body) (string, error)` | |
| `format(rut, { dots })` | `Format(rut, WithDots(b)) (string, error)` | No `incremental` mode |
| `mask(rut)` | `Mask(rut) (string, error)` | |
| `equals(a, b, { requireValid })` | `Equals(a, b, WithRequireValid(b)) bool` | |
| `generate({ bodyLength, format })` | `Generate(WithBodyLength(n), WithFormat(f)) string` | `Dotted`, `Compact`, `Hyphen` |
| `generate({ count })` | `GenerateN(n, ...) []string` | |
| `RUT_PATTERN.any` / `.compact` / `.hyphen` / `.dotted` | `PatternAny` / `PatternCompact` / `PatternHyphen` / `PatternDotted` | Same strings, byte for byte |
| `InvalidRutError` | `ErrInvalidRut` | Check with `errors.Is` |

Options are functional options named after the rut.ts keys, so a value can come from a variable: `rut.WithStrict(cfg.Strict)`. When an option is repeated, the last one wins.

## Behaviour

Two families of functions, as in rut.ts:

- **Acceptance** (`Validate`, `IsRutLike`) accepts only the three canonical shapes, `123456785`, `12345678-5` and `12.345.678-5`, with optional surrounding whitespace and a `k` or `K` verifier. Leading zeros, other dot groupings and inner spaces are rejected. The body must have 7 or 8 digits.
- **Recovery** (`Clean`, `GetBody`, `GetVerifier`, `Decompose`, `Mask`, `Format`, `Equals`) strips every character outside `[0-9kK]` wherever it sits, then leading zeros, and uppercases `k`. Only `Format` and `Equals` (by default) also check the Modulo 11 verifier.

A few details are easy to miss:

- `Format("")` returns `""` and no error, as in rut.ts.
- Errors are always the bare `ErrInvalidRut`, whose message is the constant `invalid RUT input`. The rejected value never appears in an error, so RUTs do not leak into logs.
- `Mask` is not anonymization: about 90,000 RUTs share each masked value.
- `Generate` uses `crypto/rand` and never emits repeated-digit bodies. Its results cover the real RUT range, so any of them may belong to a real person: use them as test data only.

### Rules that come from JavaScript

rut.ts runs on JavaScript, and three of its rules depend on how JavaScript treats text. rut.go reproduces them on purpose, because the obvious Go equivalents give different answers:

| Rule | rut.ts | Obvious Go code (not used) |
|---|---|---|
| Surrounding whitespace that `Validate` and `IsRutLike` trim | The 25 characters of JavaScript's `trim()`, `U+FEFF` included | `strings.TrimSpace` also trims `U+0085` and does not trim `U+FEFF` |
| Separators that `CalculateVerifier` ignores | `.`, `-` and JavaScript's `\s`, the same 25 characters | Go's regexp `\s` covers only 5 ASCII characters, not even `\v` or `U+00A0` |
| The 64-character input cap | Counts UTF-16 code units: `ñ` is 1, `😀` is 2 | `len` counts bytes: `ñ` is 2, `😀` is 4 |

Go strings may also hold bytes that are not valid UTF-8, which JavaScript strings cannot. Each such byte counts as one unit for the cap and is treated as an ordinary non-RUT character.

## How parity with rut.ts is verified

- **Golden fixture.** [`scripts/golden/generate.mjs`](scripts/golden/generate.mjs) runs the pinned rut.ts release from npm over a reproducible corpus of 48,537 inputs and 11,000 pairs, and stores every answer in [`testdata/golden.json.gz`](testdata). [`golden_test.go`](golden_test.go) replays all of them against the Go port. The corpus covers every ASCII and Latin-1 character, Unicode spaces, zero-width characters, other digit systems, look-alikes of `K` and characters outside the BMP, placed before, after and inside RUTs. It also covers the 64-unit cap from both sides, every body one digit away from a repeated-digit one, random strings, mutated valid and wrong RUTs, and pairs that differ only in the verifier or in one digit. Before it was committed, each of the three JavaScript traps above was put into the code on purpose, and the fixture caught each one.
- **Generate.** The fixture also records what rut.ts's `generate()` returns in 3,000 calls when its random source is replaced by a fixed stream of draws, with draws at both ends of each range, biased draws and draws that land on repeated-digit bodies among them. The Go test feeds the same stream to `Generate` and requires the same RUTs from the same draws.
- **Patterns.** The fixture stores rut.ts's `RUT_PATTERN`, and the test requires the `Pattern` constants to match it.
- **CI** regenerates the fixture from the pinned rut.ts and fails if it is stale.
- **Fuzz tests** ([`fuzz_test.go`](fuzz_test.go)) check the rules that tie the functions together on any input, invalid UTF-8 included.
- **Unit tests** port the cases of the rut.ts test suite, plus the JavaScript-specific ones.

When rut.ts publishes a new version, bump it in [`scripts/golden/package.json`](scripts/golden/package.json), then run:

```bash
cd scripts/golden && npm install && node generate.mjs && cd ../.. && go test ./...
```

## Not ported

- **`format(rut, { incremental: true })`**, the format-as-you-type mode. It serves input fields in a browser, not a backend.
- **`throwOnError`.** Go returns errors instead of throwing them, so every function that can fail returns `ErrInvalidRut` and the option has nothing left to do.
- **Non-string input.** rut.ts accepts `unknown` and returns `false` or `null` for numbers, `null` or objects. The Go compiler only lets strings in.
- **The cross-copy `instanceof` brand** of `InvalidRutError`. Go links one copy of each major version.
- **The prototype-pollution guard** on options. Go options are functions, not objects with a prototype.

## Performance

Measured with `go test -bench . -benchmem` on an Apple M-series laptop: `Validate` takes about 120 ns with 2 allocations, `Clean` about 37 ns, `Format` about 57 ns and `CalculateVerifier` about 28 ns.

## License

MIT
