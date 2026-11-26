// Generates ../../testdata/golden.json.gz: what the pinned rut.ts release
// answers for a large, reproducible corpus of inputs. golden_test.go replays
// every case against the Go port and requires the same answer.
//
//   npm ci && node generate.mjs           # rewrite the fixture
//   npm ci && node generate.mjs --check   # fail if the fixture is stale
//
// The corpus comes from a seeded PRNG, so every run produces the same file.
// It never contains lone UTF-16 surrogates: JSON cannot carry them to Go,
// whose decoder would replace them with U+FFFD.

import { readFileSync, writeFileSync } from 'node:fs'
import { createRequire } from 'node:module'
import { gunzipSync, gzipSync } from 'node:zlib'

import * as rut from 'rut.ts'

const OUT = new URL('../../testdata/golden.json.gz', import.meta.url)
const SEED = 0x52555447 // "RUTG"
const { version } = createRequire(import.meta.url)('rut.ts/package.json')

// --- Seeded PRNG (mulberry32) ---
let state = SEED
const random = () => {
  state = (state + 0x6d2b79f5) | 0
  let t = Math.imul(state ^ (state >>> 15), 1 | state)
  t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}
const int = (n) => Math.floor(random() * n)
const pick = (list) => list[int(list.length)]
const range = (from, to) => Array.from({ length: to - from + 1 }, (_, i) => from + i)

// A RUT with a seeded body; its verifier is right unless `wrong` is set.
const seededRut = (bodyLength, wrong = false) => {
  const body = String(bodyLength === 7 ? 1000000 + int(9000000) : 10000000 + int(90000000))
  let verifier = rut.calculateVerifier(body)
  if (wrong) verifier = pick([...'0123456789K'].filter((v) => v !== verifier))
  return { body, verifier }
}
const spell = ({ body, verifier }, shape) => {
  if (shape === 'compact') return body + verifier
  if (shape === 'hyphen') return `${body}-${verifier}`
  return `${body.replace(/\B(?=(\d{3})+$)/g, '.')}-${verifier}`
}
const SHAPES = ['dotted', 'hyphen', 'compact']

// --- Characters worth testing next to a RUT ---
const codePoints = [
  ...range(0x00, 0xff), // ASCII and Latin-1, U+0085 and U+00A0 included
  0x1680, 0x180e, ...range(0x2000, 0x200f), 0x2028, 0x2029, ...range(0x202a, 0x202f),
  0x205f, 0x2060, 0x3000, 0xfeff, 0xfffd, // spaces, zero-widths, BOM, replacement
  ...range(0x0660, 0x0669), ...range(0x0966, 0x096f), ...range(0xff10, 0xff19), // other digits
  0x041a, 0x043a, 0x212a, 0x039a, // look-alikes of K: Cyrillic, Kelvin sign, Greek
  0x1f600, 0x10000, 0x1d7ce, // outside the BMP (two UTF-16 units)
]
for (let i = 0; i < 200; i++) {
  let c
  do c = 0x100 + int(0xfff0)
  while (c >= 0xd800 && c <= 0xdfff)
  codePoints.push(c)
}
const specials = codePoints.map((c) => String.fromCodePoint(c))

const inputs = new Set()
const add = (s) => inputs.add(s)

// 1. Hand-picked inputs from the rut.ts and rut.go test suites.
;[
  '', ' ', '   ', '\t\n', '0', 'K', 'k', '-', '.', '0000000', '000000000',
  '12.345.678-5', '12345678-5', '123456785', '1.234.567-4', '1234567-4', '12345674',
  '14.625.621-K', '14.625.621-k', '9.068.826-K', '9068826k', '10.000.004-0', '10000130',
  '1000000-9', '99.999.999-9', '100000000-7', '999999-K', '999.999-K', '1-9',
  '009.068.826-K', '0009068826K', '01234567-4', '012345678-5', '0012345678-9',
  '12.345.6785', '12.345678-5', '12345.678-5', '1.2.3.4.5.6.7.8-5', '12 345 678 5',
  '12,345,678-5', 'K2345678-5', '1234K678-5', '1234567K-K', 'KK345678-5', '1234567KK',
  'RUT: 12.345.678-5', '(12.345.678-K)', '12#34%56&@78K', '12345678--5', '11.111.111-1',
  '8.888.888-K', '8888888K', ' 0 12345678', '18-264-958', '0018264958', 'abcdefgh',
].forEach(add)

// 2. Every special character before, after, inside and in place of the verifier.
for (const base of ['12.345.678-5', '12345678-5', '123456785', '9.068.826-K']) {
  for (const c of specials) {
    add(c + base)
    add(base + c)
    add(c + base + c)
    add(base.slice(0, 3) + c + base.slice(3))
    add(base.slice(0, -1) + c)
  }
}
for (const c of specials) add(`12${c}345${c}678`) // CalculateVerifier separators

// 3. The 64-unit cap, approached from both sides with one- and two-unit pads.
for (const base of ['123456785', '12.345.678-5', '12345678']) {
  for (const pad of [' ', '\u{a0}', '0', '\u{3000}', '\u{feff}', 'x', '.', '\u{1f600}']) {
    for (let total = 60; total <= 68; total++) {
      const n = Math.ceil((total - base.length) / pad.length)
      add(pad.repeat(n) + base)
      add(base + pad.repeat(n))
      add(pad.repeat(Math.floor(n / 2)) + base + pad.repeat(Math.ceil(n / 2)))
    }
  }
}

// 4. Random strings over a RUT-heavy alphabet.
const alphabet = [
  ...'0123456789'.repeat(3), ...'kK..--', ' ', ' ', '\t', '\n', '\u{a0}', '\u{feff}', '\u{3000}',
  '\u0085', '\u{200b}', 'x', '#', ',', '_', '/', '\u{e9}', '\u{1f600}', '\u{ff15}',
]
for (let i = 0; i < 15000; i++) {
  let s = ''
  for (let n = int(17); n > 0; n--) s += pick(alphabet)
  add(s)
}

// 5. Valid and wrong RUTs mutated by one to three insertions, deletions or
// replacements.
const noise = [...'0123456789kK-. \t\n,_/aX', '\u{a0}', '\u{feff}', '\u{661}', '\u{ff11}', '\u0085']
for (let i = 0; i < 15000; i++) {
  const chars = [...spell(seededRut(pick([7, 8]), random() < 0.3), pick(SHAPES))]
  for (let m = int(4); m > 0; m--) {
    const at = int(chars.length + 1)
    const kind = int(3)
    if (kind === 0) chars.splice(at, 0, pick(noise))
    else if (kind === 1) chars.splice(at, 1)
    else chars.splice(at, 1, pick(noise))
  }
  add(chars.join(''))
}

// 6. Bodies with separators and leading zeros, for CalculateVerifier.
const separators = ['.', '-', ' ', '\u{a0}', '\u{feff}', '\u0085', '\t', '\u{3000}', '_', '', '', '']
for (let i = 0; i < 5000; i++) {
  let s = '0'.repeat(int(3) === 0 ? int(4) : 0)
  for (let n = 4 + int(7); n > 0; n--) s += String(int(10)) + pick(separators)
  add(s)
}

// --- Answers ---
const safe = { throwOnError: false }
const flags = (...bits) => bits.reduce((acc, bit, i) => acc | (bit ? 1 << i : 0), 0)

const cases = [...inputs].map((s) => {
  const clean = rut.clean(s, safe)
  // Go derives GetBody, GetVerifier and Decompose from Clean: make sure that
  // holds in rut.ts itself before relying on it.
  const body = rut.getBody(s, safe)
  const verifier = rut.getVerifier(s, safe)
  const decomposed = rut.decompose(s, safe)
  if (
    (clean === null) !== (body === null) ||
    (clean !== null && (body + verifier !== clean || decomposed.body !== body || decomposed.verifier !== verifier)) ||
    rut.isValidRut(s) !== rut.validate(s)
  ) {
    throw new Error(`rut.ts helpers disagree on ${JSON.stringify(s)}`)
  }
  return [
    s,
    clean,
    rut.calculateVerifier(s, safe),
    rut.format(s, safe),
    rut.format(s, { dots: false, throwOnError: false }),
    rut.mask(s, safe),
    flags(rut.isRutLike(s), rut.validate(s), rut.validate(s, { strict: true })),
  ]
})

// Pairs for Equals: one RUT spelled two ways, and unrelated or mutated pairs.
const respell = (r) => {
  const s = spell(r, pick(SHAPES))
  switch (int(5)) {
    case 0:
      return s.toLowerCase()
    case 1:
      return '0'.repeat(1 + int(3)) + s
    case 2:
      return `  ${s}\u{a0}`
    case 3:
      return `RUT: ${s}`
    default:
      return s
  }
}
const pool = [...inputs]
const pairs = []
for (let i = 0; i < 8000; i++) {
  let a, b
  if (i % 2 === 0) {
    const r = seededRut(pick([7, 8]), random() < 0.4)
    a = respell(r)
    b = respell(r)
  } else {
    a = pick(pool)
    b = random() < 0.5 ? pick(pool) : a
  }
  pairs.push([a, b, flags(rut.equals(a, b), rut.equals(a, b, { requireValid: false }))])
}

const json =
  JSON.stringify({
    rutTs: version,
    seed: SEED,
    fields: ['input', 'clean', 'calculateVerifier', 'format', 'formatNoDots', 'mask', 'isRutLike|validate|strict'],
    pairFields: ['a', 'b', 'equals|equalsShapeOnly'],
    cases,
    pairs,
  }) + '\n'

if (process.argv.includes('--check')) {
  const committed = gunzipSync(readFileSync(OUT)).toString('utf8')
  if (committed !== json) {
    console.error('testdata/golden.json.gz is stale: run `node generate.mjs` in scripts/golden.')
    process.exit(1)
  }
  console.log(`golden fixture is up to date (rut.ts ${version}, ${cases.length} inputs, ${pairs.length} pairs)`)
} else {
  writeFileSync(OUT, gzipSync(json, { level: 9 }))
  console.log(`wrote golden fixture (rut.ts ${version}, ${cases.length} inputs, ${pairs.length} pairs)`)
}
