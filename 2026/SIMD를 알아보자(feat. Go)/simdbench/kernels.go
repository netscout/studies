package kernels

func SumScalar(x []float32) float32 {
	var s float32
	for _, v := range x {
		s += v
	}
	return s
}

// 호너의 방법을 사용하여 거듭제곱 계산을 (곱셈, 덧셈)의 조합으로 변환
// 0.5v³ + 1.2v² − 3.1v + 0.7
// = v(0.5v² + 1.2v − 3.1) + 0.7
// = v(v(0.5v + 1.2) − 3.1) + 0.7
// = ((0.5v + 1.2)·v − 3.1)·v + 0.7
func PolyScalar(x, out []float32) {
	// 입력값을 체크한다. 동시에 컴파일러가 prove 과정에서 len(x) <= len(out)임을 확인한다.
	// 그러면 컴파일러는 아래 for 루프 안의 out[i]에 인덱스 접근 가드를 넣지 않는다.
	if len(out) < len(x) {
		panic("PolyScalar: out이 x보다 짧습니다.")
	}
	for i, v := range x {
		out[i] = ((0.5*v+1.2)*v-3.1)*v + 0.7
	}
}

// SumUnrolled4는 독립된 accumulator 4개를 사용한다.
// 각 덧셈이 앞 덧셈의 결과를 기다리지 않는다. 그래서 CPU가 덧셈 4개를 겹쳐 실행한다.
func SumUnrolled4(x []float32) float32 {
	var s0, s1, s2, s3 float32
	n := len(x)
	m := n - n%4 // 동시에 4개의 계산을 진행하니까 4로 나눠떨어지는 구간을 구하기
	i := 0
	for ; i < m; i += 4 {
		s0 += x[i]
		s1 += x[i+1]
		s2 += x[i+2]
		s3 += x[i+3]
	}
	s := s0 + s1 + s2 + s3
	for ; i < n; i++ { // 4로 나눠 남은 나머지 구간을 계산
		s += x[i]
	}
	return s
}
