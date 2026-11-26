package rut_test

import (
	"errors"
	"fmt"

	rut "github.com/arrowsw/rut.go"
)

func ExampleValidate() {
	fmt.Println(rut.Validate("12.345.678-5"))
	fmt.Println(rut.Validate("12345678-5"))
	fmt.Println(rut.Validate("12.345.678-9"))  // wrong verifier
	fmt.Println(rut.Validate("012.345.678-5")) // leading zero
	fmt.Println(rut.Validate("11.111.111-1", rut.WithStrict(true)))
	// Output:
	// true
	// true
	// false
	// false
	// false
}

func ExampleClean() {
	c, err := rut.Clean("RUT: 12.345.678-k")
	fmt.Println(c, err)

	_, err = rut.Clean("not a rut")
	fmt.Println(errors.Is(err, rut.ErrInvalidRut))
	// Output:
	// 12345678K <nil>
	// true
}

func ExampleClean_thenValidate() {
	// Clean accepts messy input but does not check the verifier; Validate does.
	raw := "009.068.826-k"
	if c, err := rut.Clean(raw); err == nil && rut.Validate(c) {
		fmt.Println("valid:", c)
	}
	// Output:
	// valid: 9068826K
}

func ExampleFormat() {
	dotted, _ := rut.Format("123456785")
	plain, _ := rut.Format("12.345.678-5", rut.WithDots(false))
	_, err := rut.Format("123456789") // wrong verifier
	fmt.Println(dotted)
	fmt.Println(plain)
	fmt.Println(err)
	// Output:
	// 12.345.678-5
	// 12345678-5
	// invalid RUT input
}

func ExampleCalculateVerifier() {
	v, _ := rut.CalculateVerifier("12.345.678")
	fmt.Println(v)
	// Output:
	// 5
}

func ExampleDecompose() {
	d, _ := rut.Decompose("9.068.826-k")
	fmt.Println(d.Body, d.Verifier)
	// Output:
	// 9068826 K
}

func ExampleMask() {
	m, _ := rut.Mask("12.345.678-5")
	fmt.Println(m)
	// Output:
	// 12.***.***-5
}

func ExampleEquals() {
	fmt.Println(rut.Equals("12.345.678-5", "123456785"))
	fmt.Println(rut.Equals("12345678-9", "12345678-9")) // not a valid RUT
	fmt.Println(rut.Equals("12345678-9", "12345678-9", rut.WithRequireValid(false)))
	// Output:
	// true
	// false
	// true
}

func ExampleGenerate() {
	r := rut.Generate(rut.WithFormat(rut.Hyphen), rut.WithBodyLength(7))
	fmt.Println(rut.Validate(r), len(r))
	// Output:
	// true 9
}

func ExampleIsRutLike() {
	fmt.Println(rut.IsRutLike("12.345.678-0")) // right shape, wrong verifier
	fmt.Println(rut.IsRutLike("12.345678-5"))  // non-canonical grouping
	// Output:
	// true
	// false
}
