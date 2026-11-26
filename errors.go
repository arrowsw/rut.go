package rut

import "errors"

// ErrInvalidRut is the error returned by every function that cannot produce a
// result from its input. Check it with [errors.Is].
//
// Its message is the constant "invalid RUT input" and it never includes the
// value that was rejected, so RUTs do not leak into logs or traces. It is the
// Go counterpart of rut.ts's InvalidRutError.
var ErrInvalidRut = errors.New("invalid RUT input")
