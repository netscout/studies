//go:build goexperiment.simd

package kernels

import (
	"runtime"
	"simd"
	"sync"
)

func SumSIMDSingle(x []float32) float32 {
	var acc simd.Float32s // accumulator vector. 실행 환경이 지원하는 SIMD lane 수만큼 누적 슬롯을 가진다. 초깃값은 Mac의 경우 [0, 0, 0, 0]
	lanes := acc.Len()    // NEON은 4, AVX는 8 또는 16

	i := 0
	for ; i+lanes <= len(x); i += lanes { // lane 수만큼 딱 떨어지는 구간을 동시에 계산
		acc = acc.Add(simd.LoadFloat32s(x[i : i+lanes])) // 주어진 슬라이스에서 lane 수만큼 잘라서 각 lane별로 누적
	}

	var tmp [16]float32    // 스택에 배열을 생성(AVX512의 경우 최대 float32를 16개 담을 수 있으므로 최댓값으로 생성)
	acc.Store(tmp[:lanes]) // 각 lane의 값을 배열에 저장
	var s float32
	for j := 0; j < lanes; j++ { // 모든 lane의 결과 값을 더하기
		s += tmp[j]
	}

	for ; i < len(x); i++ { // lane 수로 딱 떨어지지 않고 남았던 인덱스에 대해 계산
		s += x[i]
	}
	return s
}

func SumSIMDUnrolled4(x []float32) float32 {
	var a0, a1, a2, a3 simd.Float32s // 4개의 accumulator vector를 생성
	lanes := a0.Len()                // vector 하나가 담는 lane 수 확인
	step := lanes * 4                // 4개의 accumulator vector -> Mac의 경우 4 * 4 => 16

	i := 0
	for ; i+step <= len(x); i += step { // step 단위로 딱 떨어지는 구간을 계산 (Mac은 한 번에 16개)
		a0 = a0.Add(simd.LoadFloat32s(x[i : i+lanes])) // 각 accumulator vector에 lane 수만큼 끊어서 누적
		a1 = a1.Add(simd.LoadFloat32s(x[i+lanes : i+2*lanes]))
		a2 = a2.Add(simd.LoadFloat32s(x[i+2*lanes : i+3*lanes]))
		a3 = a3.Add(simd.LoadFloat32s(x[i+3*lanes : i+4*lanes]))
	}

	// step만큼 끊고 남은 나머지 중에서 lane 수만큼 끊을 수 있는 구간을 계산
	for ; i+lanes <= len(x); i += lanes {
		a0 = a0.Add(simd.LoadFloat32s(x[i : i+lanes]))
	}

	// (a0+a1)과 (a2+a3)를 먼저 묶어서 안쪽 Add 두 개가 겹쳐 실행되게 한다
	acc := a0.Add(a1).Add(a2.Add(a3))

	var tmp [16]float32 // 스택에 배열을 생성(AVX512의 경우 최대 float32를 16개 담을 수 있으므로 최댓값으로 생성)
	acc.Store(tmp[:lanes])
	var s float32
	for j := 0; j < lanes; j++ {
		s += tmp[j]
	}

	for ; i < len(x); i++ {
		s += x[i]
	}
	return s
}

func SumSIMDUnrolled8(x []float32) float32 {
	var a0, a1, a2, a3, a4, a5, a6, a7 simd.Float32s // 8개의 accumulator vector를 생성
	lanes := a0.Len()                                // vector 하나가 담는 lane 수 확인
	step := lanes * 8                                // 8개의 accumulator vector -> Mac의 경우 4 * 8 => 32

	i := 0
	for ; i+step <= len(x); i += step { // step 단위로 딱 떨어지는 구간을 계산 (Mac은 한 번에 32개)
		a0 = a0.Add(simd.LoadFloat32s(x[i : i+lanes])) // 각 accumulator vector에 lane 수만큼 끊어서 누적
		a1 = a1.Add(simd.LoadFloat32s(x[i+lanes : i+2*lanes]))
		a2 = a2.Add(simd.LoadFloat32s(x[i+2*lanes : i+3*lanes]))
		a3 = a3.Add(simd.LoadFloat32s(x[i+3*lanes : i+4*lanes]))
		a4 = a4.Add(simd.LoadFloat32s(x[i+4*lanes : i+5*lanes]))
		a5 = a5.Add(simd.LoadFloat32s(x[i+5*lanes : i+6*lanes]))
		a6 = a6.Add(simd.LoadFloat32s(x[i+6*lanes : i+7*lanes]))
		a7 = a7.Add(simd.LoadFloat32s(x[i+7*lanes : i+8*lanes]))
	}

	// step만큼 끊고 남은 나머지 중에서 lane 수만큼 끊을 수 있는 구간을 계산
	for ; i+lanes <= len(x); i += lanes {
		a0 = a0.Add(simd.LoadFloat32s(x[i : i+lanes]))
	}

	// (a0+a1) (a2+a3) (a4+a5) (a6+a7)
	acc := a0.Add(a1).Add(a2.Add(a3)).Add(a4.Add(a5).Add(a6.Add(a7)))

	var tmp [16]float32 // 스택에 배열을 생성(AVX512의 경우 최대 float32를 16개 담을 수 있으므로 최댓값으로 생성)
	acc.Store(tmp[:lanes])
	var s float32
	for j := 0; j < lanes; j++ {
		s += tmp[j]
	}

	for ; i < len(x); i++ {
		s += x[i]
	}
	return s
}

// SumSIMDParallel은 슬라이스를 core 수만큼 나눠서 goroutine마다 SumSIMDUnrolled4를 돌린다.
// SIMD(lane)와 ILP(accumulator)에 이어 세 번째 병렬화 요소인 core 병렬화를 추가하는 구현이다.
func SumSIMDParallel(x []float32) float32 {
	workers := runtime.GOMAXPROCS(0)
	// 작은 입력(65535 이하)는 goroutine을 만들고 기다리는 비용이 계산 비용보다 크다
	if len(x) < 1<<16 || workers == 1 {
		return SumSIMDUnrolled4(x)
	}

	// 각 워커가 처리해야 할 데이터의 크기
	// 남는 데이터가 발생하지 않도록 worker-1을 더하고 worker로 나눠준다.
	chunk := (len(x) + workers - 1) / workers
	// 각 worker의 출력을 저장할 슬라이스
	partial := make([]float32, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		lo := w * chunk
		// 남는 데이터가 발생하지 않도록 chunk의 크기를 조정했기 때문에, 후반부 작업에서는 worker의 시작 지점이 데이터의 길이를 넘어서는 경우가 있음.
		if lo >= len(x) {
			break
		}
		hi := min(lo+chunk, len(x))
		wg.Add(1)
		go func(w, lo, hi int) {
			defer wg.Done()
			partial[w] = SumSIMDUnrolled4(x[lo:hi])
		}(w, lo, hi)
	}
	wg.Wait()

	var s float32
	// 부분합을 모두 더하기
	for _, p := range partial {
		s += p
	}
	return s
}

// splat은 모든 lane을 v로 채운 vector를 만든다.
// hot loop 밖에서만 호출하므로 이 작은 루프의 비용은 무시해도 된다.
func splat(v float32, lanes int) simd.Float32s {
	b := make([]float32, lanes)
	for i := 0; i < lanes; i++ {
		b[i] = v
	}
	return simd.LoadFloat32s(b[:lanes])
}

// 호너의 방법을 사용하여 거듭제곱 계산을 (곱셈, 덧셈)의 조합으로 변환
// 0.5v³ + 1.2v² − 3.1v + 0.7
// = v(0.5v² + 1.2v − 3.1) + 0.7
// = v(v(0.5v + 1.2) − 3.1) + 0.7
// = ((0.5v + 1.2)·v − 3.1)·v + 0.7
func PolySIMD(x, out []float32) {
	var probe simd.Float32s
	lanes := probe.Len() // lane 수를 확인하는 용도로만 사용

	// 4개의 다항식 계수를 lane 수만큼 펼쳐서 레지스터에 로드(ex: [0.5, 0.5, 0.5, 0.5])
	c3 := splat(0.5, lanes)
	c2 := splat(1.2, lanes)
	c1 := splat(-3.1, lanes)
	c0 := splat(0.7, lanes)

	i := 0
	for ; i+lanes <= len(x); i += lanes {
		v := simd.LoadFloat32s(x[i : i+lanes]) // 레지스터에 lane 수만큼 로드
		r := c3.Mul(v).Add(c2)                 // c3 * v + c2 => [0.5*v1+1.2, ... , 0.5*v4+1.2]
		r = r.Mul(v).Add(c1)
		r = r.Mul(v).Add(c0)
		r.Store(out[i : i+lanes]) // 출력 슬라이스의 해당되는 인덱스에 결과값을 저장
	}
	for ; i < len(x); i++ { // lane 수만큼 딱 떨어지지 않고 남은 항목에 대해 계산
		v := x[i]
		out[i] = ((0.5*v+1.2)*v-3.1)*v + 0.7
	}
}

// 추가 실험
// PolySIMDUnrolled4는 서로 다른 원소 묶음 4개를 한 바퀴에 처리한다.
// 묶음끼리는 독립이라 호너 각 단계의 대기 시간 동안 다른 묶음의 계산이 겹쳐 실행된다.
func PolySIMDUnrolled4(x, out []float32) {
	var probe simd.Float32s
	lanes := probe.Len()
	step := lanes * 4

	c3 := splat(0.5, lanes)
	c2 := splat(1.2, lanes)
	c1 := splat(-3.1, lanes)
	c0 := splat(0.7, lanes)

	i := 0
	for ; i+step <= len(x); i += step {
		v0 := simd.LoadFloat32s(x[i : i+lanes])
		v1 := simd.LoadFloat32s(x[i+lanes : i+2*lanes])
		v2 := simd.LoadFloat32s(x[i+2*lanes : i+3*lanes])
		v3 := simd.LoadFloat32s(x[i+3*lanes : i+4*lanes])

		// 원소별이 아니라 단계별로 묶어서 독립 사슬 4개가 겹쳐 실행되게 한다
		r0 := c3.Mul(v0).Add(c2)
		r1 := c3.Mul(v1).Add(c2)
		r2 := c3.Mul(v2).Add(c2)
		r3 := c3.Mul(v3).Add(c2)

		r0 = r0.Mul(v0).Add(c1)
		r1 = r1.Mul(v1).Add(c1)
		r2 = r2.Mul(v2).Add(c1)
		r3 = r3.Mul(v3).Add(c1)

		r0 = r0.Mul(v0).Add(c0)
		r1 = r1.Mul(v1).Add(c0)
		r2 = r2.Mul(v2).Add(c0)
		r3 = r3.Mul(v3).Add(c0)

		r0.Store(out[i : i+lanes])
		r1.Store(out[i+lanes : i+2*lanes])
		r2.Store(out[i+2*lanes : i+3*lanes])
		r3.Store(out[i+3*lanes : i+4*lanes])
	}

	// step만큼 끊고 남은 나머지 중에서 lane 수만큼 끊을 수 있는 구간은 기존 방식으로 계산
	for ; i+lanes <= len(x); i += lanes {
		v := simd.LoadFloat32s(x[i : i+lanes])
		r := c3.Mul(v).Add(c2)
		r = r.Mul(v).Add(c1)
		r = r.Mul(v).Add(c0)
		r.Store(out[i : i+lanes])
	}

	for ; i < len(x); i++ {
		v := x[i]
		out[i] = ((0.5*v+1.2)*v-3.1)*v + 0.7
	}
}
