package rut

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrInvalidRutHasAConstantGenericMessage(t *testing.T) {
	if got, want := ErrInvalidRut.Error(), "invalid RUT input"; got != want {
		t.Fatalf("ErrInvalidRut.Error() = %q, want %q", got, want)
	}
}

func TestErrInvalidRutMatchesThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("parsing customer: %w", ErrInvalidRut)
	if !errors.Is(wrapped, ErrInvalidRut) {
		t.Fatal("errors.Is(wrapped, ErrInvalidRut) = false, want true")
	}
}
