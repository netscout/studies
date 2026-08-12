package kernels

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
)

func data(n int) []float32 {
	x := make([]float32, n)
	for i := range x {
		x[i] = rand.Float32()
	}
	return x
}

// 정확성 우선. 허용 오차를 두는 이유: SIMD는 스칼라와 다른 순서로 더하므로
// 완전히 같은 값을 기대하는 검증은 잘못된 검증이다.
func TestSumImplsMatchScalar(t *testing.T) {
	impls := []struct {
		name string
		fn   func([]float32) float32
	}{
		{"unrolled4", SumUnrolled4},
		{"simd-single", SumSIMDSingle},
		{"simd-unrolled4", SumSIMDUnrolled4},
		{"simd-unrolled8", SumSIMDUnrolled8},
		{"simd-parallel", SumSIMDParallel},
	}
	// 모든 분기를 지나도록 길이를 고른다.
	// 32의 배수는 8-vector 본 루프만, 그 사이 값은 중간 단계와 스칼라 tail을 지난다.
	// 0과 1은 본 루프에 한 번도 들어가지 않는 극단값이다.
	for _, n := range []int{0, 1, 3, 4, 5, 7, 8, 15, 16, 17, 31, 32, 33, 63, 64, 65, 1000, 10_037} {
		x := data(n)
		want := SumScalar(x)
		for _, impl := range impls {
			t.Run(fmt.Sprintf("%s/n=%d", impl.name, n), func(t *testing.T) {
				got := impl.fn(x)
				// 상대 오차 1%에 절대 오차 1e-6을 더한다. 합이 0에 가까운 극단값 대비.
				if math.Abs(float64(want-got)) > 1e-2*math.Abs(float64(want))+1e-6 {
					t.Errorf("scalar=%v got=%v", want, got)
				}
			})
		}
	}
}

// Poly 구현들을 PolyScalar와 원소 단위로 비교한다.
// vector 경로는 Mul과 Add를 따로 실행하고 스칼라 경로는 FMA로 붙을 수 있어서
// 마지막 자리가 다를 수 있다. 그래서 아주 작은 허용 오차를 둔다.
func TestPolyImplsMatchScalar(t *testing.T) {
	impls := []struct {
		name string
		fn   func(x, out []float32)
	}{
		{"simd", PolySIMD},
		{"simd-unrolled4", PolySIMDUnrolled4},
	}
	for _, n := range []int{0, 1, 3, 4, 5, 7, 8, 15, 16, 17, 31, 32, 33, 63, 64, 65, 1000, 10_037} {
		x := data(n)
		want := make([]float32, n)
		PolyScalar(x, want)
		for _, impl := range impls {
			t.Run(fmt.Sprintf("%s/n=%d", impl.name, n), func(t *testing.T) {
				got := make([]float32, n)
				impl.fn(x, got)
				for j := range want {
					if math.Abs(float64(want[j]-got[j])) > 1e-4*math.Abs(float64(want[j]))+1e-5 {
						t.Fatalf("j=%d x=%v want=%v got=%v", j, x[j], want[j], got[j])
					}
				}
			})
		}
	}
}

func BenchmarkSum(b *testing.B) {
	for _, n := range []int{1024, 65536, 1 << 20} {
		x := data(n)
		for _, impl := range []struct {
			name string
			fn   func([]float32) float32
		}{
			{"scalar", SumScalar},
			{"unrolled4", SumUnrolled4},
			{"simd-single", SumSIMDSingle},
			{"simd-unrolled4", SumSIMDUnrolled4},
			{"simd-unrolled8", SumSIMDUnrolled8},
		} {
			b.Run(fmt.Sprintf("%s/n=%d", impl.name, n), func(b *testing.B) {
				b.SetBytes(int64(n * 4)) // → MB/s in the output
				for b.Loop() {           // Go 1.24+: keeps the call alive
					impl.fn(x)
				}
			})
		}
	}
}

// core 병렬화가 언제부터 이득인지 확인하는 벤치마크.
// 작은 입력은 goroutine 비용이 이기고, 큰 입력은 core 수만큼 벌 수 있는지 본다.
func BenchmarkSumBig(b *testing.B) {
	for _, n := range []int{1 << 20, 1 << 24} {
		x := data(n)
		for _, impl := range []struct {
			name string
			fn   func([]float32) float32
		}{
			{"simd-unrolled4", SumSIMDUnrolled4},
			{"simd-parallel", SumSIMDParallel},
		} {
			b.Run(fmt.Sprintf("%s/n=%d", impl.name, n), func(b *testing.B) {
				b.SetBytes(int64(n * 4))
				for b.Loop() {
					impl.fn(x)
				}
			})
		}
	}
}

func BenchmarkPoly(b *testing.B) {
	for _, n := range []int{1024, 65536, 1 << 20} {
		x, out := data(n), make([]float32, n)
		for _, impl := range []struct {
			name string
			fn   func(x, out []float32)
		}{
			{"scalar", PolyScalar},
			{"simd", PolySIMD},
			{"simd-unrolled4", PolySIMDUnrolled4},
		} {
			b.Run(fmt.Sprintf("%s/n=%d", impl.name, n), func(b *testing.B) {
				b.SetBytes(int64(n * 4))
				for b.Loop() {
					impl.fn(x, out)
				}
			})
		}
	}
}
