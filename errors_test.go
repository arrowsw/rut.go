package rut

import "testing"

func TestErrInvalidRutHasAConstantGenericMessage(t *testing.T) {
	if got, want := ErrInvalidRut.Error(), "invalid RUT input"; got != want {
		t.Fatalf("ErrInvalidRut.Error() = %q, want %q", got, want)
	}
}
