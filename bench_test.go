package rut

import "testing"

// Run with: go test -run '^$' -bench . -benchmem

var (
	sinkString string
	sinkBool   bool
)

func BenchmarkValidate(b *testing.B) {
	for _, in := range []string{"12.345.678-5", "123456785", "not-a-rut"} {
		b.Run(in, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sinkBool = Validate(in)
			}
		})
	}
}

func BenchmarkIsRutLike(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = IsRutLike("12.345.678-5")
	}
}

func BenchmarkClean(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkString, _ = Clean("12.345.678-5")
	}
}

func BenchmarkFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkString, _ = Format("123456785")
	}
}

func BenchmarkCalculateVerifier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkString, _ = CalculateVerifier("12345678")
	}
}

func BenchmarkEquals(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = Equals("12.345.678-5", "123456785")
	}
}

func BenchmarkMask(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkString, _ = Mask("12.345.678-5")
	}
}

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkString = Generate()
	}
}
