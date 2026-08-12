//go:build !goexperiment.simd

package kernels

func SumSIMDUnrolled8(x []float32) float32 { return SumUnrolled4(x) }
func SumSIMDUnrolled4(x []float32) float32 { return SumUnrolled4(x) }
func SumSIMDSingle(x []float32) float32    { return SumUnrolled4(x) }
func SumSIMDParallel(x []float32) float32  { return SumUnrolled4(x) }
func PolySIMD(x, out []float32)            { PolyScalar(x, out) }
func PolySIMDUnrolled4(x, out []float32)   { PolyScalar(x, out) }
